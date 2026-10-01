package models

import (
	"testing"
	"time"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
)

func TestResidentFromEntity(t *testing.T) {
	now := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	tests := []struct {
		name       string
		resident   *domain.Resident
		wantType   string
		wantStatus string
	}{
		{name: "empty type defaults to secondary", resident: &domain.Resident{ResidentID: "r1"}, wantType: "resident-secondary"},
		{name: "deleted status is kept", resident: &domain.Resident{ResidentID: "r1", Status: domain.ResidentStatusDeleted}, wantType: "resident-secondary", wantStatus: "deleted"},
		{name: "primary type", resident: &domain.Resident{ResidentID: "r1", Type: domain.ResidentTypePrimary}, wantType: "resident-primary"},
		{name: "secondary type", resident: &domain.Resident{ResidentID: "r1", Type: domain.ResidentTypeSecondary}, wantType: "resident-secondary"},
		{name: "other type", resident: &domain.Resident{ResidentID: "other-101", Type: domain.ResidentTypeOther}, wantType: "other"},
		{
			name: "copies all fields",
			resident: &domain.Resident{
				ID: "r1", ResidentID: "r1", Apartment: "101", Name: "Ana", Phone: "5511",
				Type: domain.ResidentTypePrimary, CreatedAt: now, UpdatedAt: now.Add(time.Hour), DeleteAt: now.Add(2 * time.Hour),
			},
			wantType: "resident-primary",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResidentFromEntity(tt.resident)
			wantStatus := tt.wantStatus
			if wantStatus == "" {
				wantStatus = "created"
			}
			want := Resident{
				ID: tt.resident.ID, ResidentID: tt.resident.ResidentID, Apartment: tt.resident.Apartment,
				Name: tt.resident.Name, Phone: tt.resident.Phone, Type: tt.wantType, Status: wantStatus,
				CreatedAt: tt.resident.CreatedAt, UpdatedAt: tt.resident.UpdatedAt, DeleteAt: tt.resident.DeleteAt,
			}
			if *got != want {
				t.Errorf("got %+v, want %+v", *got, want)
			}
		})
	}
}

func TestResidentToEntity(t *testing.T) {
	tests := []struct {
		name     string
		model    Resident
		wantType domain.ResidentType
	}{
		{name: "resident without type is secondary", model: Resident{ResidentID: "r1"}, wantType: domain.ResidentTypeSecondary},
		{name: "legacy resident type is secondary", model: Resident{ResidentID: "r1", Type: "resident"}, wantType: domain.ResidentTypeSecondary},
		{name: "primary", model: Resident{ResidentID: "r1", Type: "resident-primary"}, wantType: domain.ResidentTypePrimary},
		{name: "secondary", model: Resident{ResidentID: "r1", Type: "resident-secondary"}, wantType: domain.ResidentTypeSecondary},
		{name: "other", model: Resident{ResidentID: "other-101", Type: "other"}, wantType: domain.ResidentTypeOther},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.model.ToEntity()
			if got.Type != tt.wantType || got.ResidentID != tt.model.ResidentID {
				t.Errorf("got %+v, want type %q and id %q", got, tt.wantType, tt.model.ResidentID)
			}
		})
	}
}

func TestResidentToEntityStatus(t *testing.T) {
	deletedAt := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	tests := []struct {
		name       string
		model      Resident
		wantStatus domain.ResidentStatus
	}{
		{name: "created", model: Resident{Status: "created"}, wantStatus: domain.ResidentStatusCreated},
		{name: "deleted", model: Resident{Status: "deleted", DeleteAt: deletedAt}, wantStatus: domain.ResidentStatusDeleted},
		{name: "legacy active resident without status is created", model: Resident{}, wantStatus: domain.ResidentStatusCreated},
		{name: "legacy removed resident without status is deleted", model: Resident{DeleteAt: deletedAt}, wantStatus: domain.ResidentStatusDeleted},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.model.ToEntity().Status; got != tt.wantStatus {
				t.Errorf("Status = %q, want %q", got, tt.wantStatus)
			}
		})
	}
}

func TestResidentRoundTrip(t *testing.T) {
	now := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	tests := []struct {
		name     string
		resident domain.Resident
	}{
		{name: "full resident", resident: domain.Resident{
			ID: "r1", ResidentID: "r1", Apartment: "101", Name: "Ana", Phone: "5511",
			Type: domain.ResidentTypePrimary, Status: domain.ResidentStatusDeleted, CreatedAt: now, UpdatedAt: now, DeleteAt: now,
		}},
		{name: "other resident", resident: *domain.NewOtherResident("101")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := ResidentFromEntity(&tt.resident)

			raw, err := bson.Marshal(model)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			var decoded Resident
			if err := bson.Unmarshal(raw, &decoded); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}

			got := decoded.ToEntity()
			if got.ID != tt.resident.ID || got.ResidentID != tt.resident.ResidentID || got.Apartment != tt.resident.Apartment ||
				got.Name != tt.resident.Name || got.Phone != tt.resident.Phone || got.Type != tt.resident.Type ||
				!got.CreatedAt.Equal(tt.resident.CreatedAt) || !got.UpdatedAt.Equal(tt.resident.UpdatedAt) || !got.DeleteAt.Equal(tt.resident.DeleteAt) {
				t.Errorf("round trip = %+v, want %+v", got, tt.resident)
			}
		})
	}
}

func TestResidentBSONFieldNames(t *testing.T) {
	raw, err := bson.Marshal(ResidentFromEntity(&domain.Resident{ResidentID: "r1"}))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	for _, key := range []string{"_id", "resident_id", "apartment", "name", "phone", "type", "createdat", "updatedat", "deleteat"} {
		t.Run(key, func(t *testing.T) {
			if _, err := bson.Raw(raw).LookupErr(key); err != nil {
				t.Errorf("field %q not found in %v", key, bson.Raw(raw))
			}
		})
	}
}

func TestDeliveryRoundTrip(t *testing.T) {
	now := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	tests := []struct {
		name     string
		delivery domain.Delivery
	}{
		{name: "empty delivery", delivery: domain.Delivery{}},
		{name: "pending delivery", delivery: domain.Delivery{
			ID: "d1", DeliveryID: "d1", Apartment: "101", ResidentID: "r1", PackageType: "box", Urgency: "high",
			Status: domain.DeliveryStatusPending, CreatedAt: now, UpdatedAt: now,
		}},
		{name: "notified and deleted delivery", delivery: domain.Delivery{
			ID: "d2", DeliveryID: "d2", Apartment: "102", ResidentID: "other-102", Status: domain.DeliveryStatusDeleted,
			CreatedAt: now, UpdatedAt: now.Add(3 * time.Hour), DeleteAt: now.Add(3 * time.Hour),
			ArrivalNotifiedAt: now.Add(time.Minute), PickupNotifiedAt: now.Add(3 * time.Hour),
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := DeliveryFromEntity(&tt.delivery)
			if model.Status != string(tt.delivery.Status) {
				t.Errorf("model status = %q, want %q", model.Status, tt.delivery.Status)
			}

			raw, err := bson.Marshal(model)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			var decoded Delivery
			if err := bson.Unmarshal(raw, &decoded); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}

			got := decoded.ToEntity()
			want := tt.delivery
			if got.ID != want.ID || got.DeliveryID != want.DeliveryID || got.Apartment != want.Apartment ||
				got.ResidentID != want.ResidentID || got.PackageType != want.PackageType || got.Urgency != want.Urgency ||
				got.Status != want.Status {
				t.Errorf("round trip = %+v, want %+v", got, want)
			}
			times := []struct {
				name      string
				got, want time.Time
			}{
				{"CreatedAt", got.CreatedAt, want.CreatedAt},
				{"UpdatedAt", got.UpdatedAt, want.UpdatedAt},
				{"DeleteAt", got.DeleteAt, want.DeleteAt},
				{"ArrivalNotifiedAt", got.ArrivalNotifiedAt, want.ArrivalNotifiedAt},
				{"PickupNotifiedAt", got.PickupNotifiedAt, want.PickupNotifiedAt},
			}
			for _, ts := range times {
				if !ts.got.Equal(ts.want) {
					t.Errorf("%s = %v, want %v", ts.name, ts.got, ts.want)
				}
			}
		})
	}
}

func TestDeliveryBSONFieldNames(t *testing.T) {
	raw, err := bson.Marshal(DeliveryFromEntity(&domain.Delivery{DeliveryID: "d1"}))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	keys := []string{
		"_id", "delivery_id", "apartment", "resident_id", "package_type", "urgency", "status",
		"createdat", "updatedat", "deleteat", "arrivalnotifiedat", "pickupnotifiedat",
	}
	for _, key := range keys {
		t.Run(key, func(t *testing.T) {
			if _, err := bson.Raw(raw).LookupErr(key); err != nil {
				t.Errorf("field %q not found in %v", key, bson.Raw(raw))
			}
		})
	}
}
