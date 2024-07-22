package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	pb "messenger"
	"net/http"
	"reflect"
	"unicode"
)

func IsValidUsername(username string) bool {
	if len(username) <= 3 {
		return false
	}

	hasLetter := false
	hasDigit := false

	for _, char := range username {
		if unicode.IsLetter(char) {
			hasLetter = true
		}
		if unicode.IsDigit(char) {
			hasDigit = true
		}
		if hasLetter && hasDigit {
			return true
		}
	}

	return false
}

func IsUniqueUsername(s *Server, username string) bool {
	for _, user := range s.users {
		if user.Username == username {
			return false
		}
	}
	return true
}

func GetUserByID(s *Server, userId int32) (*User, error) {
	user, ok := s.users[userId]
	if !ok {
		return nil, errors.New("user_id was not found")
	}
	return user, nil
}

func GetUserByUsername(s *Server, username string) (*User, error) {
	var user *User
	user, err := nil, errors.New("username was not found")
	for _, v := range s.users {
		if v.Username == username {
			user, err = v, nil
			break
		}
	}
	return user, err
}

func GetUserByIDOrUsername(s *Server, user *pb.UserIdOrName) (*User, error) {
	log.Println(reflect.TypeOf(user.User))
	switch t := user.User.(type) {
	case *pb.UserIdOrName_UserId:
		return GetUserByID(s, t.UserId)
	case *pb.UserIdOrName_Username:
		return GetUserByUsername(s, t.Username)
	default:
		return nil, errors.New("argument is neather user_id nor username")
	}
}

type RequestBody struct {
	FileID string `json:"file_id"`
}

type ResponseBody struct {
	OK bool `json:"ok"`
}

func IsValidFile(s *Server, fileId string) (bool, error) {
	url := s.fileServerURL + "/checkfile"
	reqBody := RequestBody{FileID: fileId}
	reqBodyJSON, err := json.Marshal(reqBody)
	if err != nil {
		return false, fmt.Errorf("failed to marshal request body: %v", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqBodyJSON))
	if err != nil {
		return false, fmt.Errorf("failed to send POST request: %v", err)
	}
	defer resp.Body.Close()

	var respBody ResponseBody
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return false, fmt.Errorf("failed to decode response body: %v", err)
	}

	return respBody.OK, nil
}

func ValidateContent(s *Server, content *pb.Content) (string, error) {
	switch t := content.Content.(type) {
	case *pb.Content_Text:
		return t.Text, nil
	case *pb.Content_FileId:
		valid, err := IsValidFile(s, t.FileId)
		if err != nil {
			return "", errors.New("unable to validate file")
		}

		if valid {
			return t.FileId, nil
		} else {
			return "", errors.New("file is not valid")
		}
	default:
		return "", errors.New("invalid content type")
	}
}
