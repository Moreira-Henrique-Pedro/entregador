// package entities contém as definições de entidades do domínio, representando os dados e comportamentos principais do sistema.
package entities

import "time"

type Delivery struct {
	ID          string
	ApNum       string
	PackageType string
	Urgency     string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeleteAt    time.Time
}
