package driveService

import (
	"net/http"

	"google.golang.org/api/drive/v3"
)

type FileService struct {
	instance *drive.Service
}

//Singleton pattern for getting the fileservice instance.
func (fs *FileService) GetInstance(client *http.Client) (*drive.Service, error) {
	if fs.instance != nil {
		return fs.instance, nil
	} else {
		var err error
		fs.instance, err = drive.New(client)
		return fs.instance, err
	}
}
