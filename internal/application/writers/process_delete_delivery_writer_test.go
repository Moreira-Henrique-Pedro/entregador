package writers

import (
	"context"
	"errors"
	"testing"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/commands"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
)

func TestProcessDeleteDelivery(t *testing.T) {
	tests := []struct {
		name      string
		deleteErr error
		wantErr   bool
	}{
		{name: "marks as deleted"},
		{name: "not found is ignored", deleteErr: entities.ErrEntityNotFound},
		{name: "repository error is returned", deleteErr: errors.New("mongo down"), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deliveries := &fakeDeliveryRepository{deleteErr: tt.deleteErr}
			writer := NewProcessDeleteDelivery(deliveries)

			err := writer.Handle(context.Background(), &commands.ProcessDeleteDeliveryCommand{CommandID: "cmd-1", DeliveryID: "d1"})
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if len(deliveries.deletedIDs) != 1 || deliveries.deletedIDs[0] != "d1" {
				t.Errorf("deleted ids = %v, want [d1]", deliveries.deletedIDs)
			}
		})
	}
}
