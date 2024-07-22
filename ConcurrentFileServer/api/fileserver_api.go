package api

import (
	"ConcurrentFileServer/core"
	"ConcurrentFileServer/utils"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type API interface {
	SetupRoutes()
	Home(w http.ResponseWriter, r *http.Request)
	UploadFile(w http.ResponseWriter, r *http.Request)
}
type APIImpl struct {
	handler core.FileHandler
}

func NewAPI() API {
	return &APIImpl{handler: core.NewFileHandlerImpl()}
}

func (api *APIImpl) Home(w http.ResponseWriter, r *http.Request) {
	response := "Welcome\n\n/\t\tHelp page\n/upload\t\tupload file\n/download\tdownload file"
	fmt.Fprintf(w, "%s\n", response)
}

func (api *APIImpl) DownloadFile(w http.ResponseWriter, r *http.Request) {
	fileId := r.FormValue("file_id")
	var body RequestDownloadJson
	err := json.NewDecoder(r.Body).Decode(&body)
	if err == nil && body.FileId != "" {
		fileId = body.FileId
	}

	if fileId == "" {
		ErrorResponse(w, "Error: file_id not specified")
		return
	}

	ctx := context.Background()
	var data []byte
	var mime string
	data, mime, err = api.handler.DownloadFile(ctx, fileId)

	if err != nil {
		fmt.Println("Log |", err)
		ErrorResponse(w, "Error: Finding the File")
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+fileId+"."+utils.GetExtensionByMimeType(mime))
	w.Header().Set("Content-Type", mime)

	_, err = w.Write(data)
	if err != nil {
		fmt.Println("Log |", err)
		ErrorResponse(w, "Error: writing the File")
		return
	}
}

func (api *APIImpl) UploadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	ctx := context.Background()
	data, mimeType, err := RetriveFile(r)
	if err != nil {
		fmt.Println("Log |", err)
		ErrorResponse(w, "Error: faild retriving file")
		return
	}

	file_id, err := api.handler.UploadFile(ctx, data, mimeType)

	if err != nil {
		fmt.Println("Log |", err)
		ErrorResponse(w, "Error: faild Saving the File")
		return
	}

	response := ResponseJson{FileId: file_id}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (api *APIImpl) CheckFile(w http.ResponseWriter, r *http.Request) {
	fileId := r.FormValue("file_id")
	var body RequestDownloadJson
	err := json.NewDecoder(r.Body).Decode(&body)
	if err == nil && body.FileId != "" {
		fileId = body.FileId
	}

	if fileId == "" {
		ErrorResponse(w, "Error: file_id not specified")
		return
	}

	response := CheckfileResponseJson{Ok: api.handler.CheckFileExists(fileId)}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (api *APIImpl) SetupRoutes() {
	http.HandleFunc("/", api.Home)
	http.HandleFunc("/upload", api.UploadFile)
	http.HandleFunc("/download", api.DownloadFile)
	http.HandleFunc("/checkfile", api.CheckFile)
	fmt.Println("Listening...")
	http.ListenAndServe(":8000", nil)
}
