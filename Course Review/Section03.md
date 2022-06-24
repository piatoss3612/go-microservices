# Section03 Review

### 1. Postgres 데이터베이스 드라이버

```go
package main

import (
	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)
```

### 2. Postgres 데이터베이스 연결

```go
var counts int64

func connectToDB() *sql.DB {
	dsn := os.Getenv("DSN")

    // 데이터베이스가 준비가 되어있지 않아 연결이 되지 않을 수 있으므로
    // for 루프 내에서 전역변수 counts를 늘려가며 2초 쉬고 연결을 재시도하는 작업을 반복한다
    // 만약 counts가 10보다 커지면 다른 문제가 발생한 것이므로 시도를 종료한다
	for {
		conn, err := openDB(dsn)
		if err != nil {
			log.Println("Postgres not yet ready...")
			counts++
		} else {
			log.Println("Connected to Postgres!")
			return conn
		}

		if counts > 10 {
			log.Println(err)
			return nil
		}

		log.Println("Backing off for two seconds...")
		time.Sleep(2 * time.Second)
		continue
	}
}

func openDB(dsn string) (*sql.DB, error) {
    // pgx 드라이버와 데이터베이스 dsn을 사용해 PostgresDB를 연결한다
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

    // DB와 연결이 살아있는지 확인
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}
```

### 3. docker-compose로 PostgresDB 실행

```yaml
version: "3"

services:
  postgres:
    image: "postgres:14.2"
    ports:
      - "5432:5432"
    restart: always
    deploy:
      mode: replicated
      replicas: 1
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
      POSTGRES_DB: users
    volumes:
      - ./db-data/postgres/:/var/lib/postgresql/data/
```

### 4. broker service에서 요청을 받아 authentication service로 전달

```go
// 브로커 서비스 요청 페이로드
type RequestPayload struct {
	Action string      `json:"action"`
	Auth   AuthPayload `json:"auth,omitempty"`
    // 다른 액션 값이 추가됨에 따라 상응하는 페이로드 타입이 추가된다
}

// 인증요청 페이로드
type AuthPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// 요청 처리
func (app *Config) HandleSubmission(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload

	err := app.readJSON(w, r, &requestPayload) // 클라이언트로부터 받은 요청 디코딩
	if err != nil {
		app.errorJSON(w, err)
		return
	}

    // 액션 값에 따른 분기 처리
	switch requestPayload.Action {
	case "auth": // 인증 요청
		app.authenticate(w, requestPayload.Auth)
	default:
		app.errorJSON(w, errors.New("unkown action"))
	}
}

func (app *Config) authenticate(w http.ResponseWriter, a AuthPayload) {
	jsonData, err := json.MarshalIndent(a, "", "\t") // json 포매팅
	if err != nil {
		app.errorJSON(w, err)
		return
	}

    // authentication service로 보낼 요청 생성
	request, err := http.NewRequest("POST", "http://authentication-service/authenticate", bytes.NewBuffer(jsonData))
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	client := &http.Client{}
	response, err := client.Do(request) // 요청 실행
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	defer response.Body.Close()

    // 응답의 상태코드 확인
	if response.StatusCode == http.StatusUnauthorized {
		app.errorJSON(w, errors.New("invalid credentials"))
		return
	} else if response.StatusCode != http.StatusAccepted {
		app.errorJSON(w, errors.New("error calling authentication service"))
		return
	}

	var jsonFromService jsonResponse

    // 응답의 body를 디코딩
	err = json.NewDecoder(response.Body).Decode(&jsonFromService)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

    // 응답에 오류가 포함되어 있는지 다시 확인
	if jsonFromService.Error {
		app.errorJSON(w, err, http.StatusUnauthorized)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "Authenticated"
	payload.Data = jsonFromService.Data

    // 요청이 성공했으므로 클라이언트에게 응답
	app.writeJSON(w, http.StatusAccepted, payload)
}
```
