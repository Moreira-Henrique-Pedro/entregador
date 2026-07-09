package duration

import (
	"encoding/json"
	"time"
)

type Duration time.Duration

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *Duration) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	val, err := time.ParseDuration(str)
	*d = Duration(val)
	return err
}

func (d Duration) String() string {
	return time.Duration(d).String()
}

func (d Duration) Duration() time.Duration {
	return time.Duration(d)
}
