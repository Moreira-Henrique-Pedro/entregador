package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/Moreira-Henrique-Pedro/entregador/config"
	httpAdapter "github.com/Moreira-Henrique-Pedro/entregador/internal/adapters/in/http"
	kafkaOut "github.com/Moreira-Henrique-Pedro/entregador/internal/adapters/out/kafka"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/usecases"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/pubsub"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/watermill"
)

// API wires the HTTP adapter: residents, delivery registration/pickup and queries.
type API struct {
	Handler      http.Handler
	repositories *repositories
	publisher    pubsub.MessagePublisher[any]
}

func NewAPI(env *config.Environment, log logger.Logger) (*API, error) {
	repos, err := newRepositories(env)
	if err != nil {
		return nil, err
	}

	publisher, err := watermill.NewWatermillPublisher[any](env.Pubsub.DeliveryBrokersHosts, log)
	if err != nil {
		_ = repos.close(context.Background())
		return nil, fmt.Errorf("create kafka publisher: %w", err)
	}

	scheduler := kafkaOut.NewNotificationScheduler(publisher, env.Pubsub.CommandsTopic)

	residentHandler := httpAdapter.NewResidentHandler(httpAdapter.ResidentHandlerDependencies{
		ListResidentsByApartment: usecases.NewListResidentsByApartment(repos.residentRepository),
		ListResidentsByPhone:     usecases.NewListResidentsByPhone(repos.residentRepository),
		CreateResident:           usecases.NewCreateResident(repos.residentRepository),
		UpdateResident:           usecases.NewUpdateResident(repos.residentRepository),
		DeleteResident:           usecases.NewDeleteResident(repos.residentRepository),
	})

	deliveryHandler := httpAdapter.NewDeliveryHandler(httpAdapter.DeliveryHandlerDependencies{
		ListDeliveries:   usecases.NewListDeliveriesByApartment(repos.deliveryRepository),
		RegisterDelivery: usecases.NewRegisterDelivery(repos.deliveryRepository, repos.residentRepository, scheduler),
		DeleteDelivery:   usecases.NewDeleteDelivery(repos.deliveryRepository, scheduler),
	})

	return &API{
		Handler:      httpAdapter.NewRouter(residentHandler, deliveryHandler, log),
		repositories: repos,
		publisher:    publisher,
	}, nil
}

func (a *API) Close(ctx context.Context) error {
	return errors.Join(
		a.publisher.Close(ctx),
		a.repositories.close(ctx),
	)
}
