package entities

import "testing"

func TestResidentStatusIsValid(t *testing.T) {
	tests := []struct {
		name   string
		status ResidentStatus
		want   bool
	}{
		{name: "created", status: ResidentStatusCreated, want: true},
		{name: "deleted", status: ResidentStatusDeleted, want: true},
		{name: "empty", status: "", want: false},
		{name: "unknown", status: "moved", want: false},
		{name: "case sensitive", status: "Created", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.IsValid(); got != tt.want {
				t.Errorf("ResidentStatus(%q).IsValid() = %v, want %v", tt.status, got, tt.want)
			}
		})
	}
}
