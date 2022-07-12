# Section06 Review

### 1. Go RabbitMQ 공식 라이브러리

```cmd
go get github.com/rabbitmq/amqp091-go
```


### 2. docker-compose로 RabbitMQ 실행

- [RabbitMQ Dockerhub](https://hub.docker.com/_/rabbitmq)

```yaml
version: '3'

services:

  rabbitmq:
    image: 'rabbitmq:3.10.6-alpine'
    ports:
      - '5672:5672'
    deploy:
      mode: replicated
      replicas: 1
    volumes:
      - ./db-data/rabbitmq/:/var/lib/rabbitmq/
```

### 3. RabbitMQ 클라이언트 연결

```go
package main

import (
	"fmt"
	"log"
	"math"
	"time"

	"listener/event"

	amqp "github.com/rabbitmq/amqp091-go"
)
```

```go
func connect() (*amqp.Connection, error) {
	var counts int64
	var backOff = 1 * time.Second
	var connection *amqp.Connection

	for {
        // AMQP URI: amqp://guest:guest@rabbitmq
        // RabbitMQ 클라이언트 연결
		c, err := amqp.Dial("amqp://guest:guest@rabbitmq")
		if err != nil {
			fmt.Println("RabbitMQ not yet ready...")
			counts++
		} else {
			log.Println("Connected to RabbitMQ!")
			connection = c
			break
		}

		if counts > 5 {
			fmt.Println(err)
			return nil, err
		}

        // RabbitMQ 클라이언트가 준비될 때까지 backoff 시간동안 대기
		backOff = time.Duration(math.Pow(float64(counts), 2))
		log.Println("backing off...")
		time.Sleep(backOff)
		continue
	}

	return connection, nil
}
```

### 4. RabbitMQ Exchange, Queue 생성

```go
package event

import (
	amqp "github.com/rabbitmq/amqp091-go"
)
```

#### Exchange 생성

```go
func declareExchange(ch *amqp.Channel) error {
	return ch.ExchangeDeclare(
		"logs_topic", // exchange 이름
		"topic", // exchange 타입 
        true, // durable?
		false, // auto-delete?
		false, // internal: true인 경우 외부 메시지를 받지 않는다
		false, // nowait: true인 경우 서버의 확인 응답을 기다리지 않는다
		nil, // 그 외에 추가할 인수
	)
}
```

#### 무작위 임시 Queue 생성

```go
func declareRandomQueue(ch *amqp.Channel) (amqp.Queue, error) {
	return ch.QueueDeclare(
		"", // queue 이름: 명시하지 않을 경우 무작위로 생성
        false, // durable?
		true, // auto-delete?
		true, // exclusive?
		false, // no-wait?
		nil, // extra arguments
	)
}

```

### 5. RabbitMQ로 메시지 발행

```go
package event

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)
```

#### 발행자 객체

```go
type Emitter struct {
	conn *amqp.Connection
}

// 팩토리 함수
func NewEventEmitter(conn *amqp.Connection) (Emitter, error) {
	emitter := Emitter{
		conn: conn,
	}

    // exchange 생성
	err := emitter.setup()
	if err != nil {
		return Emitter{}, err
	}

	return emitter, nil
}
```

```go
func (e *Emitter) setup() error {
	channel, err := e.conn.Channel()
	if err != nil {
		return err
	}

	defer channel.Close()
	return declareExchange(channel)
}
```

#### Exchange로 메시지 발행

```go
func (e *Emitter) Push(event string, severity string) error {
	channel, err := e.conn.Channel() // 채널 연결
	if err != nil {
		return err
	}

	defer channel.Close()

	log.Println("Pushing message to channel")

    // 'logs_topic' exchange로 메시지 발행
	err = channel.Publish(
		"logs_topic",
		severity, // key
		false, // mandatory?
		false, // immediate?
		amqp.Publishing{
			ContentType: "text/plain", // 메시지 타입
			Body:        []byte(event), // 메시지 본문
		},
	)

	if err != nil {
		return err
	}

	return nil
}

```

### 6. RabbitMQ로 메시지 소비

```go
package event

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	amqp "github.com/rabbitmq/amqp091-go"
)
```

#### 소비자 객체

```go
type Consumer struct {
	conn      *amqp.Connection
	queueName string
}

// 팩토리 함수
func NewConsumer(conn *amqp.Connection) (Consumer, error) {
	consumer := Consumer{
		conn: conn,
	}

    // exchange 선언
	err := consumer.setup()
	if err != nil {
		return Consumer{}, err
	}

	return consumer, nil
}

func (consumer *Consumer) setup() error {
	channel, err := consumer.conn.Channel()
	if err != nil {
		return err
	}

	return declareExchange(channel)
}
```

#### 메시지 형식

```go
type Payload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}
```

#### 무작위 Queue를 생성하여 특정 Exchange에 바인딩

```go
func (consumer *Consumer) Listen(topics []string) error {
	ch, err := consumer.conn.Channel()
	if err != nil {
		return err
	}

	defer ch.Close()

    // 무작위 queue 생성
	q, err := declareRandomQueue(ch)
	if err != nil {
		return err
	}

    // topics에 지정된 모든 topic에 대하여 queue와 exchange를 바인딩
	for _, topic := range topics {
		err = ch.QueueBind(
			q.Name, // 무작위로 생성된 큐의 이름
			topic, // 큐에서 받을 메시지의 토픽 지정
			"logs_topic", // 바인딩할 exchange 이름 지정
			false, // no-wait?
			nil, // extra arguments
		)

		if err != nil {
			return err
		}
	}

    // queue로 메시지 전달 받기
	messages, err := ch.Consume(
        q.Name, // 큐의 이름
        "", // 소비자 이름
        true, // auto acknowledge?
        false, // exclusive?
        false, // no-local?
        false, // no-wait?
        nil, // extra arguments
        )

	if err != nil {
		return err
	}

	forever := make(chan bool)

    // goroutine 실행
	go func() {
        // amqp.Delivery 타입의 채널에서 전달받은 메시지를 꺼내 처리
		for msg := range messages {
			var payload Payload
			_ = json.Unmarshal(msg.Body, &payload) // json 형식의 데이터를 디코딩

            // goroutine 내부에서 다른 goroutine 실행
			go handlePayload(payload)
		}
	}()

	fmt.Printf("Waiting for message on [Exchange, Queue] [logs_topic, %s]\n", q.Name)
	<-forever // goroutine이 종료되지 않도록 어떤 값도 들어오지 않는 채널에서 대기

	return nil
}
```

#### AMQP를 통해 요청받은 이벤트 처리

```go
func handlePayload(payload Payload) {
	switch payload.Name {
	case "log", "event":
		// log whatever we get
		err := logEvent(payload)
		if err != nil {
			log.Println(err)
		}
	case "default":
		err := logEvent(payload)
		if err != nil {
			log.Println(err)
		}
	}
}
```

#### logger 마이크로 서비스로 logging 이벤트 요청

```go
func logEvent(entry Payload) error {
	jsonData, _ := json.MarshalIndent(entry, "", "\t")

	logServiceURL := "http://logger-service/log"

	request, err := http.NewRequest("POST", logServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		return err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		return err
	}

	return nil
}
```