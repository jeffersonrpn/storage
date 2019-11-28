package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ncw/swift"
)

// BackupClient struct containing a conn with swift library
type BackupClient struct {
	conn swiftConnection
}

const (
	jusContainer = "dadosjusbr"
)

type swiftConnection interface {
	ObjectPut(container string, objectName string, contents io.Reader, checkHash bool, Hash string, contentType string, h swift.Headers) (headers swift.Headers, err error)
	ObjectDelete(container string, objectName string) error
}

// NewBackupClient Create a client connect with Cloud
func NewBackupClient(userName, apiKey, authURL, domain string) *BackupClient {
	return &BackupClient{conn: &swift.Connection{UserName: userName, ApiKey: apiKey, AuthUrl: authURL, Domain: domain}}
}

//UploadFile Store a file in cloud container and return a Backup file containing a URL and a Hash for that file.
func (bc *BackupClient) uploadFile(path string) (*Backup, error) {

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error Opening file at %s: %q", path, err)
	}
	defer f.Close()
	headers, err := bc.conn.ObjectPut(jusContainer, filepath.Base(path), f, true, "", "", nil)
	if err != nil {
		return nil, fmt.Errorf("error trying to upload file at %s to storage: %q\nHeaders: %v", path, err, headers)
	}
	return &Backup{URL: fmt.Sprintf("%s/%s/%s", bc.storageURL(), jusContainer, filepath.Base(path)), Hash: headers["Etag"]}, nil
}

func (bc *BackupClient) storageURL() string {
	if v, ok := bc.conn.(*swift.Connection); ok {
		return v.StorageUrl
	}
	return ""
}

//DeleteFile delete a file from cloud container.
func (bc *BackupClient) deleteFile(path string) error {

	err := bc.conn.ObjectDelete(jusContainer, filepath.Base(path))
	if err != nil {
		return fmt.Errorf("delete file error: 'BackupClient:deleteFile' %s to storage: %q\nHeaders", path, err)
	}
	return nil
}

//Backup is the API to make URL and HASH files to be used in mgo store function
func (bc *BackupClient) Backup(Files []string) ([]Backup, error) {
	if len(Files) == 0 {
		return nil, fmt.Errorf("There is no file to upload")
	}
	var backups []Backup
	for _, value := range Files {
		back, err := bc.uploadFile(value)
		if err != nil {
			return nil, fmt.Errorf("Error in BackupClient:backup upload file %v", err)
		}
		backups = append(backups, *back)

	}
	return backups, nil
}
