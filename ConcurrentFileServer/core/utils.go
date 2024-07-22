package core

import (
	"io"
	"os"
	"sync"
)

func ConcurrentRead(filepath string, chunkSize int) (*[]byte, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	fileInfo, _ := file.Stat()
	fileSize := fileInfo.Size()

	data := make([]byte, fileSize)
	var wg sync.WaitGroup
	for offset := int64(0); offset < fileSize; offset += int64(chunkSize) {
		wg.Add(1)
		if offset+int64(chunkSize) > fileSize {
			chunkSize = int(fileSize - offset)
		}
		go func() {
			file.ReadAt(data[offset:offset+int64(chunkSize)], offset)
			wg.Done()
		}()
	}
	wg.Wait()
	return &data, nil
}

func ConcurrentWrite(filepath string, data []byte, chunkSize int) error {
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	fileSize := int64(len(data))
	var wg sync.WaitGroup
	for offset := int64(0); offset < fileSize; offset += int64(chunkSize) {
		wg.Add(1)
		if offset+int64(chunkSize) > fileSize {
			chunkSize = int(fileSize - offset)
		}
		go func() {
			file.WriteAt(data[offset:offset+int64(chunkSize)], offset)
			wg.Done()
		}()
	}
	wg.Wait()
	return nil
}

func SequentialWrite(filepath string, data []byte) error {
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func SequentialRead(filepath string) (*[]byte, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	return &data, nil
}
