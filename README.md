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

[Udemy](https://www.udemy.com/course/working-with-microservices-in-go/)
