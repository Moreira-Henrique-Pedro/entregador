package providers

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Moreira-Henrique-Pedro/entregador/api"
	"github.com/Moreira-Henrique-Pedro/entregador/config"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/controllers"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/usecases"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/repositories"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/services"
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

	handler := server.New(log, env.HTTP.CORSAllowedOrigins, newControllers(env, apiDependencies{
		residents:  repos.residentRepository,
		deliveries: repos.deliveryRepository,
		scheduler:  pubSub.scheduler,
		notifier:   deliveryNotifier,
		users:      identityProvider,
		authorizer: authorizer,
		pushAuth:   pushAuth,
	})...)

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

type apiDependencies struct {
	residents  repositories.ResidentRepository
	deliveries repositories.DeliveryRepository
	scheduler  services.NotificationScheduler
	notifier   services.Notifier
	users      services.UserRegistry
	authorizer controllers.Authorizer
	pushAuth   gin.HandlerFunc
}

func newControllers(env *config.Environment, deps apiDependencies) []server.Controller {
	routes := []server.Controller{
		controllers.NewHealthController(),
		controllers.NewResidentsController(controllers.ResidentsControllerDependencies{
			Authorizer:               deps.authorizer,
			ListResidentsByApartment: usecases.NewListResidentsByApartment(deps.residents),
			ListResidentsByPhone:     usecases.NewListResidentsByPhone(deps.residents),
			CreateResident:           usecases.NewCreateResident(deps.residents),
			UpdateResident:           usecases.NewUpdateResident(deps.residents),
			DeleteResident:           usecases.NewDeleteResident(deps.residents),
		}),
		controllers.NewDeliveriesController(controllers.DeliveriesControllerDependencies{
			Authorizer:       deps.authorizer,
			ListDeliveries:   usecases.NewListDeliveries(deps.deliveries, deps.residents),
			RegisterDelivery: usecases.NewRegisterDelivery(deps.deliveries, deps.residents, deps.scheduler),
			DeleteDelivery:   usecases.NewDeleteDelivery(deps.deliveries, deps.scheduler),
		}),
		controllers.NewApartmentsController(controllers.ApartmentsControllerDependencies{
			Authorizer:     deps.authorizer,
			ListApartments: usecases.NewListApartments(deps.residents),
		}),
		controllers.NewUsersController(controllers.UsersControllerDependencies{
			Authorizer: deps.authorizer,
			CreateUser: usecases.NewCreateUser(deps.users),
		}),
		controllers.NewNotificationsController(
			usecases.NewNotifyDelivery(deps.deliveries, deps.residents, deps.notifier),
			deps.pushAuth,
		),
	}

	if !env.IsProduction() {
		routes = append(routes, controllers.NewDocsController(api.OpenAPISpec))
	}
	return routes
}
