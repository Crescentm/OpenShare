package imports

import "gorm.io/gorm"

type ImportRepository struct {
	db *gorm.DB
}

func NewImportRepository(db *gorm.DB) *ImportRepository {
	return &ImportRepository{db: db}
}
