package model

import "gorm.io/gorm"

type Package struct {
	gorm.Model         //ID, CreatedAt, UpdatedAt, DeletedAt
	BlockNum    string `gorm:"not null"`
	ApNum       string `gorm:"not null"`
	PackageType string `gorm:"not null"`
	Urgency     string `gorm:"not null"`
}
