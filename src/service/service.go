package service

import (
	"github.com/Moreira-Henrique-Pedro/entregador/src/model"
	"gorm.io/gorm"
)

type BoxService struct {
	db *gorm.DB
}

func NewBoxService(db *gorm.DB) *BoxService {
	return &BoxService{
		db: db,
	}
}

func (b *BoxService) CreateBox(box model.Box) (uint64, error) {
	result := b.db.Create(&box)
	if result.Error != nil {
		return 0, result.Error
	}

	return uint64(box.ID), nil
}

func (b *BoxService) FindBoxByID(id uint64) (model.Box, error) {
	box := new(model.Box)
	resp := b.db.First(&box, id)
	if resp.Error != nil {
		return model.Box{}, resp.Error
	}

	return *box, nil
}

func (b *BoxService) UpdateBox(box model.Box, id uint64) (model.Box, error) {
	exist := new(model.Box)
	result := b.db.First(&exist, id)
	if result.Error != nil {
		return model.Box{}, result.Error
	}

	exist.Status = box.Status

	resp := b.db.Save(&exist)
	if resp.Error != nil {
		return model.Box{}, resp.Error
	}

	return *exist, nil
}

func (b *BoxService) DeleteBoxByID(id uint64) error {
	result := b.db.Delete(&model.Box{}, id)
	if result.Error != nil {
		return result.Error
	}

	return nil
}
