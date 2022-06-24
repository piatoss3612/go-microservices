# Section02 Review

### 1. HTTP 서버 설정

- `github.com/go-chi/chi/v5`: 라우터
- `github.com/go-chi/chi/v5/middleware`: 미들웨어
- `github.com/go-chi/cors`: CORS

```go
func (app *Config) routes() http.Handler {
	mux := chi.NewRouter() // 라우터 생성

	// cors 미들웨어 설정
	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "DELETE", "PUT", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// 서버가 살아있는지 확인할 수 있는 미들웨어 설정
	mux.Use(middleware.Heartbeat("/ping"))

	// POST 요청 경로와 핸들러 설정
	mux.Post("/", app.Broker)

	return mux
}
```

### 2. HTTP 서버 실행

```go
package main

import (
	"fmt"
	"log"
	"net/http"
)

const webPort = "80"

type Config struct{}

func main() {
	app := Config{}

	log.Printf("Starting broker service on port: %v\n", webPort)

	// HTTP 서버 정의
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	// 서버 실행
	if err := srv.ListenAndServe(); err != nil {
		log.Panic(err)
	}
}
```

### 3. docker 컨테이너 실행 방법

#### 첫번째 방법

1. docker로 실행하는 과정에서 go 언어 이미지를 불러온다
2. 현재 디렉토리를 컨테이너로 복사
3. go 언어 이미지를 사용해 컨테이너 내부에서 마이크로서비스 빌드
4. 빌드된 마이크로서비스 실행

```dockerfile
# base go image
FROM golang:1.18-alpine as builder

RUN mkdir /app

COPY . /app

WORKDIR /app

RUN CGO_ENABLED=0 go build -o brokerApp ./cmd/api

RUN chmod +x /app/brokerApp

#build a tiny docker image
FROM alpine:latest

RUN mkdir /app

COPY --from=builder /app/brokerApp /app

CMD ["/app/brokerApp"]
```

#### 두번째 방법

1. 미리 빌드해 놓은 마이크로서비스를 컨테이너로 복사
2. 마이크로서비스 실행

```dockerfile
FROM alpine:latest

RUN mkdir /app

COPY brokerApp /app

CMD ["/app/brokerApp"]
```

> 2번이 훨씬 빠르다

### 4. Helper function

- HTTP 요청을 읽어오고 응답을 작성하는 작업은 수없이 반복되는 작업들이다
- 따라서 이러한 작업들을 별도의 함수로 정의하고 재사용하는 것이 효율적

```go
package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type jsonResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func (app *Config) readJSON(w http.ResponseWriter, r *http.Request, data any) error {
	maxBytes := 1048576 // 1 megabyte

	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes)) // 요청의 body 용량을 제한

	dec := json.NewDecoder(r.Body)
	err := dec.Decode(data)
	if err != nil {
		return err
	}

	// body에서 데이터를 읽어오고도 아직 뭔가가 남아있는지 확인
	err = dec.Decode(&struct{}{})
	// 에러가 End Of File 에러가 아니라면
	if err != io.EOF {
		return errors.New("body must have only a single json value")
	}

	return nil
}

func (app *Config) writeJSON(w http.ResponseWriter, status int, data any, headers ...http.Header) error {
	out, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if len(headers) > 0 {
		for key, value := range headers[0] {
			w.Header()[key] = value
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_, err = w.Write(out)
	if err != nil {
		return err
	}

	return nil
}

func (app *Config) errorJSON(w http.ResponseWriter, err error, status ...int) error {
	statusCode := http.StatusBadRequest

	if len(status) > 0 {
		statusCode = status[0]
	}

	var payload jsonResponse
	payload.Error = true
	payload.Message = err.Error()

	return app.writeJSON(w, statusCode, payload)
}
```

### 5. Makefile은 들여쓰기로 탭을 사용해야 한다

- 스페이스를 사용하여 들여쓰기를 하면 `Makefile:7: *** missing separator` 이와 같은 오류가 발생한다
- 어디서 스페이스가 사용되었는지 확인하기 어렵다면, VS Code에서 Ctrl + Shift + p를 누르고
- `들여쓰기를 탭으로 변환(Convert indentation to tabs)`을 검색하고 Makefile에 적용해준다
