# Section07 Review

### 1. RPC 서버 설정

- Broker -> Logger 서비스로 RPC(Remote Procedure Call)를 통해 JSON 형식으로 인코딩/디코딩할 필요없이 빠르고 간편하게 메시지를 전달할 수 있다
- RPC를 사용하기 위해서 서버와 클라이언트는 동일한 언어로 작성되어야 하며 RPC 서버에 등록할 서버 객체와 메서드를 정의해야 한다

```go
package main

import (
	"context"
	"log"
	"logger-service/data"
	"time"
)
```

#### Logger 서비스: RPC 서버 객체 및 메서드 정의

```go
// 서버 객체
type RPCServer struct {
}

// 메시지 형식
type RPCPayload struct {
	Name string
	Data string
}

// 서버 객체의 메서드
func (r *RPCServer) LogInfo(payload RPCPayload, resp *string) error {
	collection := client.Database("logs").Collection("logs")

	_, err := collection.InsertOne(context.TODO(), data.LogEntry{
		Name:      payload.Name,
		Data:      payload.Data,
		CreatedAt: time.Now(),
	})
	if err != nil {
		log.Println("error writing to MongoDB:", err)
		return err
	}

    // 포인터로 받은 문자열의 실제값으로 메서드 실행 결과 응답
	*resp = "Processed payload via RPC:" + payload.Name
	return nil
}

```


### 2. RPC 서버 객체 등록 및 RPC 서버 실행

#### Logger 서비스: RPC 서버 객체 및 메서드 등록

```go
import (
	"fmt"
	"log"
	"net"
	"net/rpc" // RPC 서버 연결 및 실행은 표준 라이브러리를 사용
)

const (
	rpcPort  = "5001"
)

func main() {
    //...

	// RPC 서버 객체인 RPCServer와 메서드를 RPC 서버에 등록
	err := rpc.Register(new(RPCServer))
	if err != nil {
		log.Panic(err)
	}

    // 고루틴으로 RPC 서버 리스너 실행
	go app.rpcListen()

    // ...
}
```

#### Logger 서비스: RPC 서버 리스너 실행

```go
func (app *Config) rpcListen() error {
	log.Println("Starting RPC server on port", rpcPort)

    // 모든 ip의 5001번 포트로 들어오는 tcp 연결에 대한 리스너 선언
	listen, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", rpcPort))
	if err != nil {
		return err
	}
	defer listen.Close()

	for {
        // 리스너로 들어오는 새로운 RPC 연결 확인
		rpcConn, err := listen.Accept()
        // 연결을 확인할 수 없는 경우 무한 반복
		if err != nil {
			continue
		}

        // RPC 연결이 확인되면 고루틴으로 RPC 서버 실행
		go rpc.ServeConn(rpcConn)
	}
}
```

### 3. RPC 서버를 통해 메시지 전달

#### Broker 서비스: Logger 서비스로 메시지 전달

```go
package main

import (
	"net/http"
	"net/rpc"
)

// 서버를 정의할 때와 동일한 메시지 형식
type RPCPayload struct {
	Name string
	Data string
}

func (app *Config) logItemViaRPC(w http.ResponseWriter, entry LogPayload) {
    // RPC 서버를 통해 Logger 서비스의 5001번 포트로 tcp 연결
	client, err := rpc.Dial("tcp", "logger-service:5001")
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	rpcPayload := RPCPayload(entry)

	var result string

    // RPC 서버에 등록된 RPCServer 객체의 LogInfo 메서드 호출
    // JSON으로 인코딩/디코딩을 거치지 않고 평문 형태인 rpcPayload를 그대로 전달
    // string 타입인 result의 포인터를 넘겨주고 응답을 받아온다
	err = client.Call("RPCServer.LogInfo", rpcPayload, &result)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: result,
	}

	app.writeJSON(w, http.StatusAccepted, payload)
}
```