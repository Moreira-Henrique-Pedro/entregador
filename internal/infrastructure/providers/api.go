package providers

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/Moreira-Henrique-Pedro/entregador/config"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/controllers"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/usecases"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/infrastructure/server"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
)

type API struct {
	Handler      http.Handler
	repositories *mongoRepositories
	pubSub       *pubSub
}

func NewAPI(env *config.Environment, log logger.Logger) (*API, error) {
	deliveryNotifier, err := newNotifier(env)
	if err != nil {
		return nil, fmt.Errorf("create notifier: %w", err)
	}

	pushAuth, err := newPubSubPushAuth(env, log)
	if err != nil {
		return nil, err
	}

	identityProvider, err := NewIdentityProvider(context.Background(), env)
	if err != nil {
		return nil, err
	}

	authorizer, err := newAuthorizer(env, log, identityProvider)
	if err != nil {
		return nil, err
	}

	repos, err := newMongoRepositories(env)
	if err != nil {
		return nil, err
	}

	pubSub, err := newPubSub(env)
	if err != nil {
		_ = repos.close(context.Background())
		return nil, err
	}

	residents, deliveries := repos.residentRepository, repos.deliveryRepository
	scheduler := pubSub.scheduler

	handler := server.New(log,
		controllers.NewHealthController(),
		controllers.NewResidentsController(controllers.ResidentsControllerDependencies{
			Authorizer:               authorizer,
			ListResidentsByApartment: usecases.NewListResidentsByApartment(residents),
			ListResidentsByPhone:     usecases.NewListResidentsByPhone(residents),
			CreateResident:           usecases.NewCreateResident(residents),
			UpdateResident:           usecases.NewUpdateResident(residents),
			DeleteResident:           usecases.NewDeleteResident(residents),
		}),
		controllers.NewDeliveriesController(controllers.DeliveriesControllerDependencies{
			Authorizer:       authorizer,
			ListDeliveries:   usecases.NewListDeliveriesByApartment(deliveries),
			RegisterDelivery: usecases.NewRegisterDelivery(deliveries, residents, scheduler),
			DeleteDelivery:   usecases.NewDeleteDelivery(deliveries, scheduler),
		}),
		controllers.NewUsersController(controllers.UsersControllerDependencies{
			Authorizer: authorizer,
			CreateUser: usecases.NewCreateUser(identityProvider),
		}),
		controllers.NewNotificationsController(
			usecases.NewNotifyDelivery(deliveries, residents, deliveryNotifier),
			pushAuth,
		),
	)

	return &API{
		Handler:      handler,
		repositories: repos,
		pubSub:       pubSub,
	}, nil
}

func (a *API) Close(ctx context.Context) error {
	return errors.Join(
		a.pubSub.close(),
		a.repositories.close(ctx),
	)
}
