package validation

import (
	"fmt"
	"strings"
)

func Required(kind error, field, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%w: %s is required", kind, field)
	}
	return nil
}

func First(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}
