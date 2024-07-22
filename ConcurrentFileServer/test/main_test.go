package test

import (
	"ConcurrentFileServer/core"
	"ConcurrentFileServer/pkg"
	"ConcurrentFileServer/utils"
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUpload(t *testing.T) {
	var (
		data           = []byte("ali")
		mimeType       = "text/plain"
		filesDirectory = "../files"
		handler        = core.NewFileHandlerImpl()
		ctx            = context.Background()
	)

	fileId, err := handler.UploadFile(ctx, data, mimeType)
	assert.Nil(t, err)
	assert.NotEmpty(t, fileId)

	dir, err := os.ReadDir(filesDirectory)
	assert.Nil(t, err)
	assert.Len(t, dir, 1)
	assert.False(t, dir[0].IsDir())
	expectedFileName := fmt.Sprintf("%s.%s", fileId, utils.GetExtensionByMimeType(mimeType))
	assert.Equal(t, expectedFileName, dir[0].Name())
	file, err := os.Open(fmt.Sprintf("%s/%s", filesDirectory, dir[0].Name()))
	assert.Nil(t, err)
	buf := make([]byte, len(data))
	read, err := file.Read(buf)
	assert.Nil(t, err)
	assert.Equal(t, len(data), read)
	assert.True(t, bytes.Equal(data, buf))
}

func TestDownload(t *testing.T) {
	var (
		data = []byte("ali")

		mimeType       = "text/plain"
		filesDirectory = "../files"
		fileId         = "tmp"
		handler        = core.NewFileHandlerImpl()
		ctx            = context.Background()
	)

	create, err := os.Create(fmt.Sprintf("%s/%s.%s", filesDirectory, fileId, utils.GetExtensionByMimeType(mimeType)))
	assert.Nil(t, err)

	write, err := create.Write(data)
	assert.Nil(t, err)
	assert.NotEmpty(t, write)

	resultFile, resultMimeType, err := handler.DownloadFile(ctx, fileId)
	assert.Nil(t, err)
	assert.NotEmpty(t, resultFile)
	assert.NotEmpty(t, resultMimeType)
	assert.True(t, bytes.Equal(resultFile, data))
}

func TestUploadAndDownloadScenario(t *testing.T) {
	var (
		data     = []byte("ali")
		mimeType = "text/plain"
		handler  = core.NewFileHandlerImpl()
		ctx      = context.Background()
	)

	fileId, err := handler.UploadFile(ctx, data, mimeType)
	assert.Nil(t, err)
	assert.NotEmpty(t, fileId)

	file, downloadMimeType, err := handler.DownloadFile(ctx, fileId)
	assert.Nil(t, err)
	assert.NotEmpty(t, downloadMimeType)
	assert.Equal(t, mimeType, downloadMimeType)
	assert.True(t, bytes.Equal(file, data))
}

func TestUploadAndDownloadConcurrent(t *testing.T) {
	var (
		workerPool = pkg.NewWorkerPool(25)
		handler    = core.NewFileHandlerImpl()
		mimeType   = "text/plain"
	)

	ticker := time.NewTicker(100 * time.Millisecond)
	timer := time.NewTimer(5 * time.Second)

	condition := true
	for condition {
		select {
		case <-timer.C:
			condition = false
		case <-ticker.C:
			workerPool.SubmitJob(func() {
				ctx := context.Background()
				data := []byte(utils.RandStringRunes(1000))
				fileId, err := handler.UploadFile(ctx, data, mimeType)
				assert.Nil(t, err)
				time.Sleep(1 * time.Second)
				file, s, err := handler.DownloadFile(ctx, fileId)
				assert.Nil(t, err)
				assert.NotEmpty(t, s)
				assert.NotEmpty(t, file)
				assert.True(t, bytes.Equal(file, data))
			})
		}
	}

}

func TestReadPerformance(t *testing.T) {
	fileSize := int(1e8)
	dur, _ := SequentialReadPerformance(fileSize)
	fmt.Printf("Sequential read |\tsize: %dkb\ttime: %v\n", fileSize/1000, dur)
	dur, _ = ConcurrentReadPerformance(fileSize)
	fmt.Printf("Concurrent read |\tsize: %dkb\ttime: %v\n", fileSize/1000, dur)
}

// to see the real result, should run the test twice I think.
func TestWritePerformance(t *testing.T) {
	fileSize := int(1e8)
	dur, _ := ConcurrentWritePerformance(fileSize)
	fmt.Printf("Concurrent write |\tsize: %dkb\ttime: %v\n", fileSize/1000, dur)
	dur, _ = SequentialWritePerformance(fileSize)
	fmt.Printf("Sequential write |\tsize: %dkb\ttime: %v\n", fileSize/1000, dur)
}

func makeFakeFile(fileSize int, filePath string) error {
	err := os.WriteFile(filePath, []byte(utils.RandStringRunes(fileSize)), 0644)
	if err != nil {
		return err
	}
	return nil
}

func ConcurrentReadPerformance(fileSize int) (time.Duration, error) {
	filepath := "../files/fakefile.txt"
	makeFakeFile(fileSize, filepath)
	start := time.Now()
	_, err := core.ConcurrentRead(filepath, 1024*1024)
	if err != nil {
		return time.Duration(0), err
	}
	delta := time.Since(start)
	os.Remove(filepath)
	return delta, nil
}

func ConcurrentWritePerformance(fileSize int) (time.Duration, error) {
	filepath := "../files/fakefile.txt"
	data := []byte(utils.RandStringRunes(fileSize))
	start := time.Now()
	err := core.ConcurrentWrite(filepath, data, 1024*1024)
	if err != nil {
		return time.Duration(0), err
	}
	delta := time.Since(start)
	os.Remove(filepath)
	return delta, nil
}

func SequentialWritePerformance(fileSize int) (time.Duration, error) {
	filepath := "../files/fakefile.txt"
	data := []byte(utils.RandStringRunes(fileSize))
	start := time.Now()
	err := core.SequentialWrite(filepath, data)
	if err != nil {
		return time.Duration(0), err
	}
	delta := time.Since(start)
	os.Remove(filepath)
	return delta, nil
}

func SequentialReadPerformance(fileSize int) (time.Duration, error) {
	filepath := "../files/fakefile.txt"
	makeFakeFile(fileSize, filepath)
	start := time.Now()
	_, err := core.SequentialRead(filepath)
	delta := time.Since(start)
	if err != nil {
		return time.Duration(0), err
	}
	os.Remove(filepath)
	os.Remove(filepath)
	return delta, nil
}
