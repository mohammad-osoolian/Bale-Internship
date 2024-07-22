package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type ErrorJson struct {
	ErrorMsg string `json:"error"`
}

type ResponseJson struct {
	FileId string `json:"file_id"`
}

type CheckfileResponseJson struct {
	Ok bool `json:"ok"`
}

type RequestUploadJson struct {
	File string `json:"file"`
}

type RequestDownloadJson struct {
	FileId string `json:"file_id"`
}

func ErrorResponse(w http.ResponseWriter, msg string) {
	response := ErrorJson{ErrorMsg: msg}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func RetriveFile(r *http.Request) ([]byte, string, error) {
	url := r.FormValue("file")
	var body RequestUploadJson
	err := json.NewDecoder(r.Body).Decode(&body)
	if err == nil && body.File != "" {
		url = body.File
	}

	if url != "" {
		response, err := http.Get(url)
		if err != nil {
			return nil, "", errors.New("error: faild to request url")
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusOK {
			return nil, "", errors.New("error: faild to download")
		}
		data, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, "", errors.New("error: faild to read bytes")
		}
		mime := response.Header.Get("Content-Type")
		return data, mime, nil
	}

	r.ParseMultipartForm(10 << 20)
	file, handler, err := r.FormFile("file")
	if err != nil {
		return nil, "", errors.New("error: file is not specified")
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, "", errors.New("error: faild to read file")
	}
	mime := handler.Header["Content-Type"][0]
	return data, mime, nil
}
