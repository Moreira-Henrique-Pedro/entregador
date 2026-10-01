package entities

import "errors"

// ErrEntityNotFound is returned by repositories when the requested entity does not exist.
// Wrap it with the entity and id, e.g. fmt.Errorf("resident %s: %w", id, ErrEntityNotFound).
var ErrEntityNotFound = errors.New("entity not found")

// ErrNoResidentInApartment is returned when a delivery targets an apartment that has
// no registered resident; the apartment's "other" resident does not count.
var ErrNoResidentInApartment = errors.New("não existe um morador cadastrado para este apartamento")
