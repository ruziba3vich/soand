package repos

import (
	"mime/multipart"

	"github.com/ruziba3vich/soand/internal/models"
)

type IBackgroundService interface {
	CreateBackground(file *multipart.FileHeader) (string, error)
	DeleteBackground(id string) error
	GetAllBackgrounds(page int64, pageSize int64) ([]models.Background, error)
	GetBackgroundByID(id string) (string, error)
}
