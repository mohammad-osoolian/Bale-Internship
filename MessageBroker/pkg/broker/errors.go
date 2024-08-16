package broker

import "errors"

var (
	// Use this error for the calls that are coming after that
	// When id is used befor in publishing message
	ErrAlreadyExistID = errors.New("message with this id already exists")
	// server is shutting down
	ErrUnavailable = errors.New("service is unavailable")
	// Use this error when the message with provided id is not available
	ErrInvalidID = errors.New("message with id provided is not valid or never published")
	// Use this error when message had been published, but it is not
	// available anymore because the expiration time has reached.
	ErrExpiredID = errors.New("message with id provided is expired")

	// Openning connection failed
	ErrDBConnect = errors.New("failed to open db connection")
	// Closing connection failed
	ErrDBClose = errors.New("failed to close db connection")
	// Query failed
	ErrRunQuery = errors.New("failed to run the query")
	// Clear data
	ErrClearData = errors.New("failed to clear data")
)
