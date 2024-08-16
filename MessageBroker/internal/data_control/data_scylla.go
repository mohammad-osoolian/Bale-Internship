package datacontrol

import (
	"fmt"
	"log"
	"therealbroker/pkg/broker"
	"time"

	"github.com/gocql/gocql"
	"github.com/google/uuid"
)

type DataScylla struct {
	DataControl
	cluster  *gocql.ClusterConfig
	session  *gocql.Session
	host     string
	port     string
	keyspace string
	forget   time.Duration
}

func NewDataScylla(host, port, keyspace string, forget time.Duration) *DataScylla {
	return &DataScylla{
		cluster:  nil,
		session:  nil,
		host:     host,
		port:     port,
		keyspace: keyspace,
		forget:   forget,
	}
}

func (ds *DataScylla) Connect() error {
	ds.cluster = gocql.NewCluster(fmt.Sprintf("%s:%s", ds.host, ds.port))
	ds.cluster.Keyspace = ds.keyspace
	ds.cluster.Consistency = gocql.Quorum
	var err error
	ds.session, err = ds.cluster.CreateSession()
	if err != nil {
		log.Println("Failed to connect to ScyllaDB: ", err)
		return broker.ErrDBConnect
	}
	return nil
}

func (ds *DataScylla) Close() error {
	ds.session.Close()
	return nil
}

func (ds *DataScylla) SaveMessage(msg broker.Message) (string, error) {
	id := gocql.TimeUUID()
	expires_at := time.Now().Add(msg.Expiration)
	ttl := int((msg.Expiration + ds.forget).Seconds())
	query := `INSERT INTO messages (id, body, expiration_duration, expires_at)
              VALUES (?, ?, ?, ?)
			  USING TTL ?;`

	err := ds.session.Query(query, id, msg.Body, int(msg.Expiration.Seconds()), expires_at, ttl).Exec()
	if err != nil {
		return "", broker.ErrRunQuery
	}

	return fmt.Sprintf("%v", id), nil
}

func (ds *DataScylla) RetriveMessage(id string) (broker.Message, error) {
	uuid, err := uuid.Parse(id)
	if err != nil {
		return broker.Message{}, broker.ErrInvalidID
	}

	query := `SELECT body, expiration_duration, expires_at FROM messages WHERE id = ?`

	cqluuid := gocql.UUID(uuid)
	msg := broker.Message{Id: id}
	var expiresAt time.Time
	if err := ds.session.Query(query, cqluuid).Scan(&msg.Body, &msg.Expiration, &expiresAt); err == gocql.ErrNotFound {
		return broker.Message{}, broker.ErrInvalidID
	} else if err != nil {
		return broker.Message{}, broker.ErrRunQuery
	}

	msg.Expiration = time.Duration(msg.Expiration * time.Second)

	if time.Now().After(expiresAt) {
		return broker.Message{}, broker.ErrExpiredID
	}

	return msg, nil
}

func (ds *DataScylla) ClearData() error {
	query := `TRUNCATE messages;`
	err := ds.session.Query(query).Exec()
	if err != nil {
		log.Println(err)
		return broker.ErrClearData
	}
	return nil
}
