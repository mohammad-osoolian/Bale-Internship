package config

import (
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

var (
	DATA_CONTROL    string
	SCYLLA_HOST     string
	SCYLLA_PORT     string
	SCYLLA_KEYSPACE string
	SCYLLA_FORGET   time.Duration

	POSTGRES_HOST   string
	POSTGRES_PORT   string
	POSTGRES_USER   string
	POSTGRES_PASS   string
	POSTGRES_DBNAME string

	GRPC_PORT string
)

func LoadConfig() error {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		return errors.New("failed to load .env file")
	}

	DATA_CONTROL = os.Getenv("DATA_CONTROL")
	if DATA_CONTROL == "scylla" {
		SCYLLA_HOST = os.Getenv("SCYLLA_HOST")
		SCYLLA_PORT = os.Getenv("SCYLLA_PORT")
		SCYLLA_KEYSPACE = os.Getenv("SCYLLA_KEYSPACE")
		forgetSeconds, err := strconv.Atoi(os.Getenv("SCYLLA_FORGET"))
		if err != nil {
			return errors.New("failed to convert SCYLLA_FORGET")
		}
		SCYLLA_FORGET = time.Duration(forgetSeconds * int(time.Second))
	}

	if DATA_CONTROL == "postgres" {
		POSTGRES_HOST = os.Getenv("POSTGRES_HOST")
		POSTGRES_PORT = os.Getenv("POSTGRES_PORT")
		POSTGRES_USER = os.Getenv("POSTGRES_USER")
		POSTGRES_PASS = os.Getenv("POSTGRES_PASS")
		POSTGRES_DBNAME = os.Getenv("POSTGRES_DBNAME")
	}

	GRPC_PORT = os.Getenv("GRPC_PORT")

	return nil
}
