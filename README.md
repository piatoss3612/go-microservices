# Working with Microservices in Go

---

### 1. Broker Service

1. broker service 생성

- 라우터: `github.com/go-chi/chi/v5`
- 미들웨어: `github.com/go-chi/chi/v5/middleware`
- CORS: `github.com/go-chi/cors`
  <br>

2. Docker 이미지 빌드

- `broker-service.dockerfile`
- `docker-compse.yml`
- 도커 서버 실행 & `docker-compose up -d`

3. broker service 테스트

- `localhost:80`을 사용할 수 없었던 문제 -> 아파치 서버를 종료함으로써 해결
- 도커 컨테이너 재실행
- 테스트 코드 작성
- `go run ./cmd/web`

4. JSON 형식의 데이터 처리를 도와주는 helper 함수 추가

- readJSON, writeJSON, errorJSON

5. make, Makefile을 사용하여 컴파일 과정 단순화

- Makefile: 컴파일에 필요한 명령어를 묶어 `make [커스텀 명령어]` 형식으로 실행
- Error: `Makefile: *** missing separator. Stop.`
  - Make 파일에서 tab을 공백\*4로 이해하는 문제 해결 방법
    - vscode -> command palette -> Convert Indentation to Tabs

### Auth Service

> User -요청-> Broker -요청전달-> Auth -> DB -> Auth -응답-> Broker -응답전달-> User

1. authentication service 구현

- 사용자 데이터 모델 추가
- 라우터 추가

[Udemy](https://www.udemy.com/course/working-with-microservices-in-go/)
