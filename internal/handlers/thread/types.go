package thread

import "gorm.io/gorm"

type ThreadHandler struct {
	db *gorm.DB
}

func NewThreadHandler(db *gorm.DB) *ThreadHandler {
	return &ThreadHandler{db: db}
}
