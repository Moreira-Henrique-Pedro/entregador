package entities

import "errors"

// ErrEntityNotFound is returned by repositories when the requested entity does not exist.
// Wrap it with the entity and id, e.g. fmt.Errorf("resident %s: %w", id, ErrEntityNotFound).
var ErrEntityNotFound = errors.New("entity not found")
