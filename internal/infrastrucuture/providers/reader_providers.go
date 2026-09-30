package providers

import (
	"context"

	"github.com/Moreira-Henrique-Pedro/entregador/config"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/readers"
	"go.mongodb.org/mongo-driver/mongo"
)

type ReaderProviders struct {
	GetResidentsByApartment  *readers.GetResidentsByApartment
	GetResidentsByPhone      *readers.GetResidentsByPhone
	GetDeliveriesByApartment *readers.GetDeliveriesByApartment
	mongoClient              *mongo.Client
}

func NewReaderProviders(env *config.Environment) (*ReaderProviders, error) {
	repos, err := newRepositoryProviders(env)
	if err != nil {
		return nil, err
	}

	return &ReaderProviders{
		GetResidentsByApartment:  readers.NewGetResidentsByApartment(repos.residentRepository),
		GetResidentsByPhone:      readers.NewGetResidentsByPhone(repos.residentRepository),
		GetDeliveriesByApartment: readers.NewGetDeliveriesByApartment(repos.deliveryRepository),
		mongoClient:              repos.mongoClient,
	}, nil
}

func (r *ReaderProviders) Close(ctx context.Context) error {
	if r == nil || r.mongoClient == nil {
		return nil
	}
	return r.mongoClient.Disconnect(ctx)
}
