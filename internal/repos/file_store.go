package repos

import (
	"mime/multipart"

	dto "github.com/ruziba3vich/soand/internal/dtos"
)

type (
	IFIleStoreService interface {
		DeleteFile(string) error
		GetFile(string) (string, error)
		UploadFile(*multipart.FileHeader) (*dto.FileObject, error)
		UploadFileFromBytes([]byte, string) (string, error)
	}
)
