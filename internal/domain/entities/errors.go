package entities

import "errors"

var ErrEntityNotFound = errors.New("entity not found")

var ErrNoResidentInApartment = errors.New("não existe um morador cadastrado para este apartamento")
