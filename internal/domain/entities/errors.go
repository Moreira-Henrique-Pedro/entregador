package entities

import "errors"

var ErrEntityNotFound = errors.New("entity not found")

var ErrNoResidentInApartment = errors.New("não existe um morador cadastrado para este apartamento")

var ErrInvalidResident = errors.New("invalid resident")

var ErrInvalidDelivery = errors.New("invalid delivery")

var ErrOtherResidentReadOnly = errors.New("other resident cannot be changed")

var ErrInvalidRecipient = errors.New("invalid notification recipient")

var ErrInvalidUser = errors.New("invalid user")

var ErrUserAlreadyExists = errors.New("user already exists")

var ErrUnauthenticated = errors.New("unauthenticated")
