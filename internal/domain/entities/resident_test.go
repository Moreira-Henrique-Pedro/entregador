package entities

import "testing"

func TestOtherResidentID(t *testing.T) {
	tests := []struct {
		name      string
		apartment string
		want      string
	}{
		{name: "apartment", apartment: "101", want: "other-101"},
		{name: "empty apartment", apartment: "", want: "other-"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := OtherResidentID(tt.apartment); got != tt.want {
				t.Errorf("OtherResidentID(%q) = %q, want %q", tt.apartment, got, tt.want)
			}
			if again := OtherResidentID(tt.apartment); again != tt.want {
				t.Errorf("OtherResidentID(%q) is not deterministic: %q", tt.apartment, again)
			}
		})
	}
	if OtherResidentID("101") == OtherResidentID("102") {
		t.Error("different apartments must have different other resident ids")
	}
}

func TestNewOtherResident(t *testing.T) {
	tests := []struct {
		name      string
		apartment string
	}{
		{name: "apartment 101", apartment: "101"},
		{name: "apartment 2B", apartment: "2B"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewOtherResident(tt.apartment)
			wantID := OtherResidentID(tt.apartment)

			if got.ID != wantID || got.ResidentID != wantID {
				t.Errorf("ID/ResidentID = %q/%q, want %q", got.ID, got.ResidentID, wantID)
			}
			if got.Apartment != tt.apartment {
				t.Errorf("Apartment = %q, want %q", got.Apartment, tt.apartment)
			}
			if got.Name != "Outro" {
				t.Errorf("Name = %q, want %q", got.Name, "Outro")
			}
			if got.Phone != "" {
				t.Errorf("Phone = %q, want empty", got.Phone)
			}
			if got.Type != ResidentTypeOther || !got.IsOther() {
				t.Errorf("Type = %q, want %q", got.Type, ResidentTypeOther)
			}
			if got.Status != ResidentStatusCreated {
				t.Errorf("Status = %q, want %q", got.Status, ResidentStatusCreated)
			}
		})
	}
}

func TestResidentIsOther(t *testing.T) {
	tests := []struct {
		name     string
		resident *Resident
		want     bool
	}{
		{name: "other", resident: &Resident{Type: ResidentTypeOther}, want: true},
		{name: "primary", resident: &Resident{Type: ResidentTypePrimary}, want: false},
		{name: "secondary", resident: &Resident{Type: ResidentTypeSecondary}, want: false},
		{name: "empty type", resident: &Resident{}, want: false},
		{name: "other id prefix but secondary type", resident: &Resident{ResidentID: OtherResidentID("101"), Type: ResidentTypeSecondary}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.resident.IsOther(); got != tt.want {
				t.Errorf("IsOther() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResidentIsPrimary(t *testing.T) {
	tests := []struct {
		name     string
		resident *Resident
		want     bool
	}{
		{name: "primary", resident: &Resident{Type: ResidentTypePrimary}, want: true},
		{name: "secondary", resident: &Resident{Type: ResidentTypeSecondary}, want: false},
		{name: "other", resident: &Resident{Type: ResidentTypeOther}, want: false},
		{name: "empty type", resident: &Resident{}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.resident.IsPrimary(); got != tt.want {
				t.Errorf("IsPrimary() = %v, want %v", got, tt.want)
			}
		})
	}
}
