package core

import (
	"ConcurrentFileServer/utils"
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"math/rand"
	"path/filepath"
	"strings"
)

type FileHandler interface {
	UploadFile(ctx context.Context, file []byte, mimeType string) (string, error)
	DownloadFile(ctx context.Context, fileID string) ([]byte, string, error)
	CheckFileExists(fileId string) bool
}

type FileHandlerImpl struct {
	RootDir   string
	FileCount int64
}

func NewFileHandlerImpl() FileHandler {
	return &FileHandlerImpl{RootDir: "../files/", FileCount: 0}
}

func (f *FileHandlerImpl) UploadFile(ctx context.Context, file []byte, mimeType string) (string, error) {
	//TODO implement me
	f.FileCount++ //wanted to use for generating access_hash but changed my mind. However it seemed useful so i kept it.

	access_hash := fmt.Sprintf("%.19d", rand.Int63())
	md5hash := md5.Sum(file)
	file_name := hex.EncodeToString(md5hash[:])
	file_id := fmt.Sprintf("%s-%s", access_hash, file_name)
	extention := utils.GetExtensionByMimeType(mimeType)
	path := fmt.Sprintf("%s%s.%s", f.RootDir, file_id, extention)
	if ctx.Err() != nil {
		return "", errors.New("error: request cancelled")
	}
	err := ConcurrentWrite(path, file, 1024*1024)
	return file_id, err
}

func (f *FileHandlerImpl) DownloadFile(ctx context.Context, fileID string) ([]byte, string, error) {
	//TODO implement me
	var extention string
	filepath.Walk(f.RootDir, func(path string, info fs.FileInfo, err error) error {
		if (!info.IsDir()) && (strings.Split(info.Name(), ".")[0] == fileID) {
			extention = strings.Split(info.Name(), ".")[1]
			return nil
		}
		return nil
	})
	if extention == "" {
		return nil, "", errors.New("error: file id not found")
	}

	mime_type := utils.GetMimeTypeByExtension(extention)
	path := fmt.Sprintf("%s%s.%s", f.RootDir, fileID, extention)
	if ctx.Err() != nil {
		return nil, "", errors.New("error: request cancelled")
	}
	file, err := ConcurrentRead(path, 1024*1024)
	return *file, mime_type, err
}

func (f *FileHandlerImpl) CheckFileExists(fileId string) bool {
	exists := false
	filepath.Walk(f.RootDir, func(path string, info fs.FileInfo, err error) error {
		if (!info.IsDir()) && (strings.Split(info.Name(), ".")[0] == fileId) {
			exists = true
			return nil
		}
		return nil
	})
	return exists
}
