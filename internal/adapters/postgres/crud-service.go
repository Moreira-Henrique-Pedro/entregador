package postgre

// type BoxService struct {
// 	db            *gorm.DB
// 	twilioService *TwilioService
// }

// func NewBoxService(db *gorm.DB, twilioService *TwilioService) *BoxService {
// 	return &BoxService{
// 		db: db,
// 	}
// }

// func (b *BoxService) CreateBox(box model.Box) (uint64, error) {
// 	slog.Info("Creating box")

// 	result := b.db.Create(&box)
// 	if result.Error != nil {
// 		slog.Error("Error to creat box")
// 		return 0, result.Error
// 	}

// 	to := "11968358817"

// 	err := b.twilioService.SendWhatsAppMessage(to)
// 	if err != nil {
// 		slog.Error("Error sending WhatsApp message")
// 		return 0, err
// 	}

// 	slog.Info("Box created sucessfuly id")

// 	return uint64(box.ID), nil
// }

// func (b *BoxService) FindBoxByID(id uint64) (model.Box, error) {
// 	slog.Info("finding box")

// 	box := new(model.Box)
// 	resp := b.db.First(&box, id)
// 	if resp.Error != nil {
// 		slog.Error("Error to creat box")
// 		return model.Box{}, resp.Error
// 	}

// 	slog.Info("Box id founded sucessfuly")

// 	return *box, nil
// }

// func (b *BoxService) UpdateBox(box model.Box, id uint64) (model.Box, error) {
// 	slog.Info("Updating box ID")

// 	exist := new(model.Box)
// 	result := b.db.First(&exist, id)
// 	if result.Error != nil {
// 		slog.Error("Error to creat box")
// 		return model.Box{}, result.Error
// 	}

// 	exist.Status = box.Status

// 	resp := b.db.Save(&exist)
// 	if resp.Error != nil {
// 		slog.Error("Error to creat box")
// 		return model.Box{}, resp.Error
// 	}

// 	slog.Info("Box id founded sucessfuly")

// 	return *exist, nil
// }

// func (b *BoxService) DeleteBoxByID(id uint64) error {
// 	slog.Info("Deleting box ID")

// 	result := b.db.Delete(&model.Box{}, id)
// 	if result.Error != nil {
// 		slog.Error("Error to delete box")
// 		return result.Error
// 	}

// 	return nil
// }
