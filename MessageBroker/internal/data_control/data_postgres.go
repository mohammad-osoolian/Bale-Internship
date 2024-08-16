package datacontrol

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"therealbroker/pkg/broker"
	"time"

	"github.com/jackc/pgx/pgtype"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DataPostgres struct {
	DataControl
	host     string
	port     string
	username string
	password string
	dbName   string
	db       *pgxpool.Pool
	batch    *PublishBatch
	ctx      context.Context
}

func NewDataPostgres(host, port, username, password, dbName string, ctx context.Context) *DataPostgres {
	return &DataPostgres{
		host:     host,
		port:     port,
		username: username,
		password: password,
		dbName:   dbName,
		db:       nil,
		batch:    nil,
		ctx:      ctx,
	}
}

type PublishBatch struct {
	lock          sync.Mutex
	msgs          []broker.Message
	responses     []chan string
	db            *pgxpool.Pool
	ctx           context.Context
	flushInterval time.Duration
	stopChan      chan bool
}

func NewPublishBatch(db *pgxpool.Pool, ctx context.Context) *PublishBatch {
	batch := PublishBatch{
		lock:          sync.Mutex{},
		msgs:          make([]broker.Message, 0),
		responses:     make([]chan string, 0),
		db:            db,
		ctx:           ctx,
		flushInterval: 500 * time.Millisecond,
		stopChan:      make(chan bool),
	}
	batch.StartExecuter()
	return &batch
}

func (b *PublishBatch) Query() string {
	var builder strings.Builder
	builder.WriteString("INSERT INTO messages (body, expiration_duration)\nVALUES\n")
	for i, msg := range b.msgs {
		if i > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString(fmt.Sprintf("('%s', '%v')", msg.Body, msg.Expiration.Seconds()))
	}
	builder.WriteString("\n RETURNING id")
	return builder.String()
}

func (b *PublishBatch) AddtoQueue(msg broker.Message) chan string {
	b.lock.Lock()
	b.msgs = append(b.msgs, msg)
	newresp := make(chan string, 1)
	b.responses = append(b.responses, newresp)
	b.lock.Unlock()
	return newresp
}

func (b *PublishBatch) Execute() {
	b.lock.Lock()
	defer b.lock.Unlock()
	if len(b.responses) == 0 {
		return
	}
	rows, err := b.db.Query(b.ctx, b.Query())
	if err != nil {
		log.Println(err)
		return
	}
	defer rows.Close()
	i := 0
	for rows.Next() {
		var id string
		err := rows.Scan(&id)
		if err != nil {
			log.Println(err)
			return
		}
		b.responses[i] <- id
		i++
	}
	b.responses = make([]chan string, 0)
	b.msgs = make([]broker.Message, 0)
}

func (b *PublishBatch) StartExecuter() {
	ticker := time.NewTicker(b.flushInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				b.Execute()
			case <-b.stopChan:
				ticker.Stop()
				return
			}
		}
	}()
}

func (b *PublishBatch) StopExecuter() {
	b.stopChan <- true
}

func (dp *DataPostgres) ClearData() error {
	query := `DELETE FROM messages`
	_, err := dp.db.Exec(dp.ctx, query)
	if err != nil {
		return broker.ErrClearData
	}
	return nil
}

func (dp *DataPostgres) Connect() error {
	connString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dp.host, dp.port, dp.username, dp.password, dp.dbName)

	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return errors.New("unable to parse database configuration")
	}

	config.MaxConns = 50
	config.MinConns = 30

	dp.db, err = pgxpool.NewWithConfig(dp.ctx, config)
	if err != nil {
		return errors.New("unable to create connection pool")
	}

	if !dp.TestConnection() {
		return errors.New("failed to ping the database")
	}

	dp.batch = NewPublishBatch(dp.db, dp.ctx)

	return nil
}

func (dp *DataPostgres) Close() error {
	dp.db.Close()
	dp.batch.StopExecuter()
	return nil
}

func (dp *DataPostgres) TestConnection() bool {
	return dp.db.Ping(dp.ctx) == nil
}

func (dp *DataPostgres) SaveMessage(msg broker.Message) (string, error) {
	resp := dp.batch.AddtoQueue(msg)
	id := <-resp
	return id, nil
}

// func (dp *DataPostgres) SaveMessage(msg broker.Message) (int, error) {
// 	query := `
//         INSERT INTO messages (body, expiration_duration)
//         VALUES ($1, $2)
//         RETURNING id
//     `

// 	row := dp.db.QueryRow(dp.ctx, query, msg.Body, msg.Expiration)
// 	var id int
// 	err := row.Scan(&id)
// 	if err != nil {
// 		log.Println(err)
// 		return 0, broker.ErrRunQuery
// 	}
// 	return id, nil
// }

func (dp *DataPostgres) RetriveMessage(id string) (broker.Message, error) {
	query := `
        SELECT id, body, expiration_duration, expires_at
        FROM messages 
        WHERE id=$1
    `
	intid, err := strconv.Atoi(id)
	if err != nil {
		return broker.Message{}, broker.ErrInvalidID
	}

	row := dp.db.QueryRow(dp.ctx, query, intid)

	msg := broker.Message{}
	var expiration pgtype.Text
	var expiresAt pgtype.Timestamptz
	err = row.Scan(&msg.Id, &msg.Body, &expiration, &expiresAt)
	if err == pgx.ErrNoRows {
		return broker.Message{}, broker.ErrInvalidID
	} else if err != nil {
		return broker.Message{}, broker.ErrRunQuery
	}

	msg.Expiration = timeStringToDuration(expiration.String)

	if time.Now().After(expiresAt.Time) {
		return broker.Message{}, broker.ErrExpiredID
	}

	return msg, nil
}

func (dp *DataPostgres) IdExists(id string) bool {
	query := `
        SELECT id
        FROM messages 
        WHERE id=$1
    `
	row := dp.db.QueryRow(dp.ctx, query, id)

	var msgId int
	err := row.Scan(&msgId)
	return err != pgx.ErrNoRows
}

func timeStringToDuration(timeStr string) time.Duration {
	// Split the time string by colon
	parts := strings.Split(timeStr, ":")
	hours, _ := strconv.Atoi(parts[0])
	minutes, _ := strconv.Atoi(parts[1])
	seconds, _ := strconv.Atoi(parts[2])

	duration := time.Duration(hours)*time.Hour +
		time.Duration(minutes)*time.Minute +
		time.Duration(seconds)*time.Second

	return duration
}
