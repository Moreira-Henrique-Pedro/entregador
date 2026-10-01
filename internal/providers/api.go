package providers

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/Moreira-Henrique-Pedro/entregador/config"
	httpAdapter "github.com/Moreira-Henrique-Pedro/entregador/internal/adapters/in/http"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/usecases"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
)

// API wires the HTTP adapter: residents, delivery registration/pickup and queries. With Pub/Sub
// it also receives the push subscription and sends the notifications itself (single binary).
type API struct {
	Handler      http.Handler
	repositories *repositories
	messaging    *messaging
}

func NewAPI(env *config.Environment, log logger.Logger) (*API, error) {
	repos, err := newRepositories(env)
	if err != nil {
		return nil, err
	}

	messaging, err := newMessaging(env, log)
	if err != nil {
		_ = repos.close(context.Background())
		return nil, err
	}

	residentHandler := httpAdapter.NewResidentHandler(httpAdapter.ResidentHandlerDependencies{
		ListResidentsByApartment: usecases.NewListResidentsByApartment(repos.residentRepository),
		ListResidentsByPhone:     usecases.NewListResidentsByPhone(repos.residentRepository),
		CreateResident:           usecases.NewCreateResident(repos.residentRepository),
		UpdateResident:           usecases.NewUpdateResident(repos.residentRepository),
		DeleteResident:           usecases.NewDeleteResident(repos.residentRepository),
	})

	deliveryHandler := httpAdapter.NewDeliveryHandler(httpAdapter.DeliveryHandlerDependencies{
		ListDeliveries:   usecases.NewListDeliveriesByApartment(repos.deliveryRepository),
		RegisterDelivery: usecases.NewRegisterDelivery(repos.deliveryRepository, repos.residentRepository, messaging.scheduler),
		DeleteDelivery:   usecases.NewDeleteDelivery(repos.deliveryRepository, messaging.scheduler),
	})

	handlers := httpAdapter.Handlers{Residents: residentHandler, Deliveries: deliveryHandler}

	if env.Messaging.Provider == config.MessagingProviderPubSub {
		push, err := newPubSubPushHandler(env, log, repos)
		if err != nil {
			_ = messaging.close(context.Background())
			_ = repos.close(context.Background())
			return nil, err
		}
		handlers.PubSubPush = push
	}

	return &API{
		Handler:      httpAdapter.NewRouter(handlers, log),
		repositories: repos,
		messaging:    messaging,
	}, nil
}

func (a *API) Close(ctx context.Context) error {
	return errors.Join(
		a.messaging.close(ctx),
		a.repositories.close(ctx),
	)
}

func newPubSubPushHandler(env *config.Environment, log logger.Logger, repos *repositories) (http.Handler, error) {
	deliveryNotifier, err := newNotifier(env)
	if err != nil {
		return nil, fmt.Errorf("create notifier: %w", err)
	}

	notifyDelivery := usecases.NewNotifyDelivery(repos.deliveryRepository, repos.residentRepository, deliveryNotifier)
	return newPushHandler(env, log, notifyDelivery)
}
