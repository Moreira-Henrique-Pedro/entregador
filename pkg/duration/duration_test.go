package duration

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDurationMarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		in   Duration
		want string
	}{
		{name: "seconds", in: Duration(4 * time.Second), want: `"4s"`},
		{name: "minutes", in: Duration(90 * time.Second), want: `"1m30s"`},
		{name: "zero", in: 0, want: `"0s"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.in)
			require.NoError(t, err)
			assert.JSONEq(t, tt.want, string(got))
		})
	}
}

func TestDurationUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    time.Duration
		wantErr bool
	}{
		{name: "seconds", in: `"4s"`, want: 4 * time.Second},
		{name: "milliseconds", in: `"250ms"`, want: 250 * time.Millisecond},
		{name: "composite", in: `"1h2m"`, want: time.Hour + 2*time.Minute},
		{name: "number not supported", in: `5`, wantErr: true},
		{name: "invalid string", in: `"soon"`, wantErr: true},
		{name: "invalid json", in: `"4s`, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d Duration
			err := json.Unmarshal([]byte(tt.in), &d)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, d.Duration())
		})
	}
}

func TestDurationInStruct(t *testing.T) {
	type cfg struct {
		Timeout Duration `json:"timeout"`
	}

	var c cfg
	require.NoError(t, json.Unmarshal([]byte(`{"timeout":"30s"}`), &c))
	assert.Equal(t, 30*time.Second, c.Timeout.Duration())
	assert.Equal(t, "30s", c.Timeout.String())

	out, err := json.Marshal(c)
	require.NoError(t, err)
	assert.JSONEq(t, `{"timeout":"30s"}`, string(out))
}
