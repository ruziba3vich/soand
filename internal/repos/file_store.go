package repos

import "mime/multipart"

type (
	IFIleStoreService interface {
		DeleteFile(string) error
		GetFile(string) (string, error)
		UploadFile(*multipart.FileHeader) (string, error)
	}
)
