# Section08 Review

### 1. Golang으로 gRPC 시작하기

- Broker -> Logger 서비스로 gRPC를 통해 Protocol Buffer로 직렬화된 데이터를 매우 빠르고 간편하게 메시지를 전달할 수 있다
- gRPC를 사용하기 위해서 서버와 클라이언트가 서로 다른 언어도 작성될 수도 있으며
- 컴파일러로 생성된 서버 인터페이스를 구현하여 gRPC 서버에 등록해야 한다


#### protoc-gen-go 플러그인 설치

- proto2, proto3 버전의 프로토콜 버퍼 언어를 Go 코드로 변환해주는 플러그인

```cmd
$ go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
```

#### protoc-gen-go-grpc 플러그인 설치

- 프로토콜 버퍼로 정의된 파일에서 Go 언어로 바인딩된 service를 생성하는 플러그인

```cmd
$ go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### 2. logs.proto 파일 작성

```go
syntax = "proto3"; // protocol buffer v3 사용

package logs; // 패키지명

option go_package = "/logs"; // go 모듈을 기준으로 go 패키지를 생성할 경로

// Log: name, data 순으로 정해진 메시지 타입
message Log {
    string name = 1; // 1번째로 들어가는 인수의 이름과 타입
    string data = 2; // 2번째로 들어가는 인수의 이름과 타입
}

// LogRequest: Log 메시지 타입을 포함하는 gRPC 요청 메시지
message LogRequest {
    Log logEntry = 1;
}

// LogResponse: result 문자열을 포함하는 gRPC 응답 메시지
message LogResponse {
    string result = 1;
}

// 서비스 정의
service LogService {
    // WriteLog 메서드는 클라이언트로부터 LogRequest를 받아 응답으로 LogResponse를 보낸다
    rpc WriteLog(LogRequest) returns (LogResponse);
}
```

### 3. Protocol Buffer 컴파일러 설치

- [Github](https://github.com/protocolbuffers/protobuf/releases/tag/v21.2)
- 설치한 컴파일러(protoc)를 GOPATH로 복사하여 간편하게 사용하기

```cmd
$ cp protoc ~/go/bin/
$ protoc --version
libprotoc 3.20.1
```

### 4. logs.proto 파일 컴파일하기

```cmd
$ cd logger-service/logs
$ protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative logs.proto
```

- `logs_grpc.pb.go`, `logs.pb.go` 파일 생성

### 5. gRPC 패키지 설치

```cmd
$ go get google.golang.org/grpc
```
```cmd
$ go get google.golang.org/protobuf
```

### 6. gRPC 서버 설정

```go
package main

import (
	"context"
	"fmt"
	"log"
	"logger-service/data"
	"logger-service/logs" // proto 파일을 컴파일한 패키지
	"net"

	"google.golang.org/grpc"
)
```

#### Logger 서비스: gRPC 서버 객체와 메서드

```go
type LogServer struct {
    // 컴파일러로 생성된 `logs_grpc.pb.go` 파일에 정의된 서버 객체를 임베딩
    // UnimplementedLogServiceServer 객체는 LogServiceServer 인터페이스를 구현
	logs.UnimplementedLogServiceServer
	Models data.Models
}

// UnimplementedLogServiceServer 객체의 WriteLog 메서드를 오버라이딩
func (l *LogServer) WriteLog(ctx context.Context, req *logs.LogRequest) (*logs.LogResponse, error) {
	input := req.GetLogEntry() // 요청 메시지에서 LogEntry를 불러온다

	logEntry := data.LogEntry{
		Name: input.Name,
		Data: input.Data,
	}

    // MongoDB에 logEntry 데이터 삽입
	err := l.Models.LogEntry.Insert(logEntry)
	if err != nil {
        // 데이터 삽입 실패 응답
		resp := &logs.LogResponse{
			Result: "failed",
		}
		return resp, err
	}

    // 데이터 삽입 성공 응답
	resp := &logs.LogResponse{
		Result: "logged",
	}

	return resp, nil
}
```

#### Logger 서비스: gRPC 서버 실행 및 요청 수신 대기

```go
func (app *Config) gRPCListen() {
    // 50001번 포트로 들어오는 tcp 연결에 대한 리스너 선언
	listen, err := net.Listen("tcp", fmt.Sprintf(":%s", gRPCPort))
	if err != nil {
		log.Fatalf("failed to listen for gRPC: %v\n", err)
	}

	server := grpc.NewServer() // gRPC 서버 생성

    // gRPC 서버에 LogServer 서버 객체 등록
	logs.RegisterLogServiceServer(server, &LogServer{
		Models: app.Models,
	})

	log.Printf("gRPC Server started on port %s\n", gRPCPort)

    // gRPC 서버에서 리스너로 들어오는 요청 수신
	if err := server.Serve(listen); err != nil {
		log.Fatalf("failed to listen for gRPC: %v\n", err)
	}
}
```

```go
package main

const (
	gRPCPort = "50001"
)

func main() {
	// ...

    // main 함수에서 고루틴으로 실행
	go app.gRPCListen()

	// ...
}
```

### 7. gRPC 클라이언트 설정

```go
package main

import (
	"broker/logs"
	"context"
	"net/http"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)
```

#### Broker 서비스: gRPC 클라이언트에서 서버로 서비스 요청

```go
func (app *Config) logViagRPC(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

    // Logger 서비스의 50001번 포트로 gRPC 클라이언트 연결 시도
	conn, err := grpc.Dial("logger-service:50001",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	defer conn.Close()

    // gRPC 클라이언트를 사용해 LogServiceClient 생성
	client := logs.NewLogServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second) // gRPC 연결은 매우 빠르 시간 안에 처리된다
	defer cancel()

    // 클라이언트에서 서버로 WriteLog 서비스 요청
	_, err = client.WriteLog(ctx, &logs.LogRequest{
		LogEntry: &logs.Log{
			Name: requestPayload.Log.Name,
			Data: requestPayload.Log.Data,
		},
	})
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "logged"

	app.writeJSON(w, http.StatusAccepted, payload)
}
```