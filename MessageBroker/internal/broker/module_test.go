package broker

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"testing"
	datacontrol "therealbroker/internal/data_control"
	"therealbroker/pkg/broker"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

var (
	service broker.Broker
	mainCtx = context.Background()
	letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	data    datacontrol.DataControl
)

func TestMain(m *testing.M) {
	// postgres := datacontrol.NewDataPostgres("localhost", "5432", "postgres", "8764", "TestDB", context.Background())
	// err := postgres.Connect()
	// if err != nil {
	// 	log.Println(err)
	// 	return
	// }
	// defer postgres.Close()
	// err = postgres.ClearData()
	// if err != nil {
	// 	log.Println(err)
	// 	return
	// }
	// data = postgres

	scylla := datacontrol.NewDataScylla("127.0.0.1", "9042", "test_db", time.Duration(10*time.Second))
	err := scylla.Connect()
	if err != nil {
		log.Println(err)
	}
	defer scylla.Close()
	err = scylla.ClearData()
	if err != nil {
		log.Println(err)
		return
	}
	data = scylla

	// data = NewDataMemory()

	service = NewModule(data)
	rand.Seed(time.Now().Unix())
	m.Run()
}

func TestPublishShouldFailOnClosed(t *testing.T) {
	msg := createMessage()

	err := service.Close()
	assert.Nil(t, err)

	_, err = service.Publish(mainCtx, "ali", msg)
	assert.Equal(t, broker.ErrUnavailable, err)
}

func TestSubscribeShouldFailOnClosed(t *testing.T) {
	err := service.Close()
	assert.Nil(t, err)

	_, err = service.Subscribe(mainCtx, "ali")
	assert.Equal(t, broker.ErrUnavailable, err)
}

func TestFetchShouldFailOnClosed(t *testing.T) {
	err := service.Close()
	assert.Nil(t, err)

	_, err = service.Fetch(mainCtx, "ali", fmt.Sprintf("%v", rand.Intn(100)))
	assert.Equal(t, broker.ErrUnavailable, err)
}

func TestPublishShouldNotFail(t *testing.T) {
	msg := createMessage()

	_, err := service.Publish(mainCtx, "ali", msg)

	assert.Equal(t, nil, err)
}

func TestSubscribeShouldNotFail(t *testing.T) {
	sub, err := service.Subscribe(mainCtx, "ali")

	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, sub)
}

func TestPublishShouldSendMessageToSubscribedChan(t *testing.T) {
	msg := createMessage()

	sub, _ := service.Subscribe(mainCtx, "ali")
	_, _ = service.Publish(mainCtx, "ali", msg)
	in := <-sub

	assert.Equal(t, msg, in)
}

func TestPublishShouldSendMessageToSubscribedChans(t *testing.T) {
	msg := createMessage()

	sub1, _ := service.Subscribe(mainCtx, "ali")
	sub2, _ := service.Subscribe(mainCtx, "ali")
	sub3, _ := service.Subscribe(mainCtx, "ali")
	_, _ = service.Publish(mainCtx, "ali", msg)
	in1 := <-sub1
	in2 := <-sub2
	in3 := <-sub3

	assert.Equal(t, msg, in1)
	assert.Equal(t, msg, in2)
	assert.Equal(t, msg, in3)
}

func TestPublishShouldPreserveOrder(t *testing.T) {
	n := 50
	messages := make([]broker.Message, n)
	sub, _ := service.Subscribe(mainCtx, "ali")
	for i := 0; i < n; i++ {
		messages[i] = createMessage()
		_, _ = service.Publish(mainCtx, "ali", messages[i])
	}

	for i := 0; i < n; i++ {
		msg := <-sub
		assert.Equal(t, messages[i], msg)
	}
}

func TestPublishShouldNotSendToOtherSubscriptions(t *testing.T) {
	msg := createMessage()
	ali, _ := service.Subscribe(mainCtx, "ali")
	maryam, _ := service.Subscribe(mainCtx, "maryam")

	_, _ = service.Publish(mainCtx, "ali", msg)
	select {
	case m := <-ali:
		assert.Equal(t, msg, m)
	case <-maryam:
		assert.Fail(t, "Wrong message received")
	}
}

func TestNonExpiredMessageShouldBeFetchable(t *testing.T) {
	msg := createMessageWithExpire(time.Second * 10)
	id, _ := service.Publish(mainCtx, "ali", msg)
	log.Println(id)
	fMsg, _ := service.Fetch(mainCtx, "ali", id)

	assert.Equal(t, msg.Body, fMsg.Body)
}

func TestExpiredMessageShouldNotBeFetchable(t *testing.T) {
	msg := createMessageWithExpire(time.Millisecond * 500)
	id, _ := service.Publish(mainCtx, "ali", msg)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	<-ticker.C
	fMsg, err := service.Fetch(mainCtx, "ali", id)
	assert.Equal(t, broker.ErrExpiredID, err)
	assert.Equal(t, broker.Message{}, fMsg)
}

func TestNewSubscriptionShouldNotGetPreviousMessages(t *testing.T) {
	msg := createMessage()
	_, _ = service.Publish(mainCtx, "ali", msg)
	sub, _ := service.Subscribe(mainCtx, "ali")

	select {
	case <-sub:
		assert.Fail(t, "Got previous message")
	default:
	}
}

func TestConcurrentSubscribesOnOneSubjectShouldNotFail(t *testing.T) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	var wg sync.WaitGroup

	for {
		select {
		case <-ticker.C:
			wg.Wait()
			return

		default:
			wg.Add(1)
			go func() {
				defer wg.Done()

				_, err := service.Subscribe(mainCtx, "ali")
				assert.Nil(t, err)
			}()
		}
	}
}

func TestConcurrentSubscribesShouldNotFail(t *testing.T) {
	ticker := time.NewTicker(2000 * time.Millisecond)
	defer ticker.Stop()
	var wg sync.WaitGroup

	for {
		select {
		case <-ticker.C:
			wg.Wait()
			return

		default:
			wg.Add(1)
			go func() {
				defer wg.Done()

				_, err := service.Subscribe(mainCtx, randomString(4))
				assert.Nil(t, err)
			}()
		}
	}
}

func TestConcurrentPublishOnOneSubjectShouldNotFail(t *testing.T) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	var wg sync.WaitGroup

	i := 1
	for {
		select {
		case <-ticker.C:
			wg.Wait()
			return

		default:
			wg.Add(1)
			go func(i int) {
				defer wg.Done()

				msg := createUniqueMessage(fmt.Sprintf("%v", i))
				_, err := service.Publish(mainCtx, "ali", msg)
				assert.Nil(t, err)
			}(i)
		}
		i++
	}
}

func TestConcurrentPublishShouldNotFail(t *testing.T) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	var wg sync.WaitGroup

	i := 1
	for {
		select {
		case <-ticker.C:
			wg.Wait()
			return

		default:
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				msg := createUniqueMessage(fmt.Sprintf("%v", i))
				_, err := service.Publish(mainCtx, randomString(4), msg)
				assert.Nil(t, err)
			}(i)
		}
		i++
	}
}

func TestDataRace(t *testing.T) {
	duration := 600 * time.Millisecond
	ticker1 := time.NewTicker(duration)
	ticker2 := time.NewTicker(duration)
	ticker3 := time.NewTicker(duration)

	defer ticker1.Stop()
	defer ticker2.Stop()
	defer ticker3.Stop()

	var wg sync.WaitGroup

	ids := make(chan string, 100000)

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer log.Println("FUNC1 DONE")
		i := 0
		for {
			select {
			case <-ticker1.C:
				return

			default:
				id, err := service.Publish(mainCtx, "ali", createUniqueMessageWithExpire(duration, fmt.Sprintf("%v", i)))
				ids <- id
				assert.Nil(t, err)
			}
			i++
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer log.Println("FUNC2 DONE")

		for {
			select {
			case <-ticker2.C:
				return

			default:
				_, err := service.Subscribe(mainCtx, "ali")
				assert.Nil(t, err)
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer log.Println("FUNC3 DONE")

		for {
			select {
			case <-ticker3.C:
				return

			case id := <-ids:
				_, err := service.Fetch(mainCtx, "ali", id)
				assert.Nil(t, err)
			}
		}
	}()
	wg.Wait()
}

func BenchmarkPublish(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := service.Publish(mainCtx, randomString(2), createMessage())
		assert.Nil(b, err)
	}
}

func BenchmarkSubscribe(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := service.Subscribe(mainCtx, randomString(2))
		assert.Nil(b, err)
	}
}

func randomString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func createUniqueMessage(id string) broker.Message {
	body := randomString(16)

	return broker.Message{
		Body:       body,
		Expiration: 0,
		Id:         id,
	}
}

func createMessage() broker.Message {
	body := randomString(16)

	return broker.Message{
		Body:       body,
		Expiration: 0,
	}
}

func createUniqueMessageWithExpire(duration time.Duration, id string) broker.Message {
	body := randomString(16)

	return broker.Message{
		Body:       body,
		Expiration: duration,
		Id:         id,
	}
}

func createMessageWithExpire(duration time.Duration) broker.Message {
	body := randomString(16)

	return broker.Message{
		Body:       body,
		Expiration: duration,
	}
}
