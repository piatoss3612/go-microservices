# Section04 Review

### 1. MongoDB 공식 드라이버

```cmd
go get go.mongodb.org/mongo-driver/mongo
```

### 2. MongoDB CRUD Operations in Go

```go
package data

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)
```

#### Models 객체 및 팩토리 함수

```go
var client *mongo.Client // MongoDB 클라이언트 전역 변수 선언

// Models 객체 생성자는 마이크로서비스가 실행될 때 최초로 1회 실행되어
// 전역 변수 client를 초기화하고 Models 객체를 생성하여 반환한다
func New(mongo *mongo.Client) Models {
	client = mongo

	return Models{
		LogEntry: LogEntry{},
	}
}

type Models struct {
	LogEntry LogEntry
}
```

#### LogEntry 구조체

- MongoDB에 저장하거나 불러오는 로그 데이터의 형식을 정의
- bson 형식은 DB, json 형식은 front-end와 통신할 때 사용한다
<br>

> MongoDB는 bson(binary json) 형태로 데이터를 저장한다

```go
type LogEntry struct {
	ID        string    `bson:"_id,omitempty" json:"id,omitempty"`
	Name      string    `bson:"name" json:"name"`
	Data      string    `bson:"data" json:"data"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}
```

#### Document 삽입

```go
func (l *LogEntry) Insert(entry LogEntry) error {
    // MongoDB -> 'logs' Database -> 'logs' collection이 이미 존재하면 불러오고 없으면 생성한다
	collection := client.Database("logs").Collection("logs")

    // collection에 단일 document 삽입
	_, err := collection.InsertOne(context.TODO(), LogEntry{
		Name:      entry.Name,
		Data:      entry.Data,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	if err != nil {
		log.Println("Error inserting into logs: ", err)
		return err
	}

	return nil
}
```

#### Document 조회

- 모든 document 조회

```go
func (l *LogEntry) All() ([]*LogEntry, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	collection := client.Database("logs").Collection("logs")

    // Find 메서드의 옵션 객체 생성
	opts := options.Find()
    // created_at 값을 기준으로 정렬 옵션 설정
	opts.SetSort(bson.D{{Key: "created_at", Value: -1}})

    // cursor 객체는 collection에 포함된 모든 document를 가리키는 객체
	cursor, err := collection.Find(context.TODO(), bson.D{}, opts)
	if err != nil {
		log.Println("Finding all docs error: ", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var logs []*LogEntry

    // cursor가 다음 document를 가리키고 있는 경우
	for cursor.Next(ctx) {
		var item LogEntry

        // bson 형식의 document를 LogEntry 객체로 디코딩
		err := cursor.Decode(&item)
		if err != nil {
			log.Println("Error decoding log into slice: ", err)
			return nil, err
		}

		logs = append(logs, &item)
	}

	return logs, nil
}
```

- id로 단일 document 조회

```go
func (l *LogEntry) GetOne(id string) (*LogEntry, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	collection := client.Database("logs").Collection("logs")

    // 유효한 16진수 문자열을 MongoDB의 ObjectId로 변환
	docID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var entry LogEntry
    // 앞서 변환한 ObjectId와 일치하는 단일 document를 찾아 LogEntry 객체로 디코딩
	err = collection.FindOne(ctx, bson.M{"_id": docID}).Decode(&entry)
	if err != nil {
		return nil, err
	}

	return &entry, nil
}
```

#### Collection 삭제

```go
func (l *LogEntry) DropCollection() error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	collection := client.Database("logs").Collection("logs")

    // 'logs' collection 삭제
	if err := collection.Drop(ctx); err != nil {
		return err
	}

	return nil
}
```

#### Document 업데이트

```go
func (l *LogEntry) Update() (*mongo.UpdateResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	collection := client.Database("logs").Collection("logs")

    // 유효한 16진수 문자열을 MongoDB의 ObjectId로 변환
	docID, err := primitive.ObjectIDFromHex(l.ID)
	if err != nil {
		return nil, err
	}

	result, err := collection.UpdateOne(
		ctx,
		bson.M{"_id": docID}, // 앞서 변환한 ObjectId에 해당하는 document를 찾는 필터
		bson.D{ // 업데이트 연산자 '$set'과 업데이트할 값을 key:value 형태로 전달
			{Key: "$set", Value: bson.D{
				{Key: "name", Value: l.Name},
				{Key: "data", Value: l.Data},
				{Key: "updated_at", Value: time.Now()},
			}},
		},
	)

	if err != nil {
		return nil, err
	}

	return result, nil
}
```

### 3. MongoDB 연결

```go
import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	mongoURL = "mongodb://mongo:27017"
)

func connectToMongo() (*mongo.Client, error) {
    // MongoDB DSN을 파싱하여 MongoDB 클라이언트 옵션에 저장
	clientOptions := options.Client().ApplyURI(mongoURL)
    // 클라이언트 인증 정보를 옵션에 저장
	clientOptions.SetAuth(options.Credential{
		Username: "admin",
		Password: "password",
	})

    // 클라이언트 옵션으로 MongoDB 연결
	conn, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Println("Error connecting: ", err)
		return nil, err
	}

	log.Println("Connected to mongo")

	return conn, nil
}
```

### 4. docker-compose로 MongoDB 실행

```yaml
version: '3'

services:

  mongo:
    image: 'mongo:4.2.16-bionic' # MongoDB 이미지
    ports:
      - "27017:27017"
    environment: # 환경변수로 루트 사용자 설정
      MONGO_INITDB_DATABASE: logs
      MONGO_INITDB_ROOT_USERNAME: admin
      MONGO_INITDB_ROOT_PASSWORD: password
    volumes: # MongoDB 데이터 저장 경로
      - ./db-data/mongo/:/data/db
```