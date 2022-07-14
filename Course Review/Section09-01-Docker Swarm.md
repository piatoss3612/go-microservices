# Section09 Review

Docker Swarm - container orchestration service

### 1. 마이크로 서비스를 Docker Image로 빌드

#### logger 서비스 이미지 빌드

```cmd
$ docker build -f logger-service.dockerfile -t piatoss3612/logger-service:1.0.0 .
$ docker image ls
```

- `f`: dockerfile 이름
- `t`: 이미지 이름:태그

### 2. 개인 Docker Hub에 저장

Docker Hub 계정으로 로그인 필요!

```cmd
$ docker login
$ docker push piatoss3612/logger-service:1.0.0
```

### 3. Docker Swarm으로 배포할 서비스를 swam.yaml 파일로 작성

```yaml
version: '3'

services:
  broker-service:
    image: piatoss3612/broker-service:1.0.0 # Docker Hub에 저장한 이미지를 불러온다
    ports:
      - "8080:80"
    deploy:
      mode: replicated
      replicas: 1

  listener-service:
    image: piatoss3612/listener-service:1.0.0
    deploy:
      mode: replicated
      replicas: 1

  ...
```

### 4. Docker Swarm으로 로컬에 마이크로 서비스 배포

#### Docker Swarm 초기화 및 배포

```cmd
$ docker swarm init
$ docker stack deploy -c swarm.yaml myapp
```

#### 실행중인 Docker 서비스 확인

```cmd
$ docker service ls
ID             NAME                           MODE         REPLICAS   IMAGE                                      PORTS
ot4xc6sb9spc   myapp_authentication-service   replicated   1/1        piatoss3612/authentication-service:1.0.0   
xz7k758130bm   myapp_broker-service           replicated   1/1        piatoss3612/broker-service:1.0.0           *:8080->80/tcp
z3ps5kl2nri0   myapp_listener-service         replicated   1/1        piatoss3612/listener-service:1.0.0
pthlm6siqtm8   myapp_logger-service           replicated   1/1        piatoss3612/logger-service:1.0.0           *:8081->80/tcp
warburhsp8pc   myapp_mailer-service           replicated   1/1        piatoss3612/mailer-service:1.0.0
usvggjxptalr   myapp_mailhog                  global       1/1        mailhog/mailhog:latest                     *:1025->1025/tcp, *:8025->8025/tcp
5onnzj1vft3y   myapp_mongo                    global       1/1        mongo:4.2.17-bionic                        *:27017->27017/tcp
kvst1sq5nqpb   myapp_postgres                 replicated   1/1        postgres:14.2                              *:5432->5432/tcp
1u05hgmgjkru   myapp_rabbitmq                 global       1/1        rabbitmq:3.10.6-alpine
```

### 5. 서비스 스케일링

#### 1. 실행중인 listener 서비스를 1개에서 3개로 늘리는 경우

```cmd
$ docker service scale myapp_listener-service=3
```

```cmd
$ docker service ls | grep listener
z3ps5kl2nri0   myapp_listener-service         replicated   3/3        piatoss3612/listener-service:1.0.0
```

#### 2. 다시 1개로 줄이는 경우

```cmd
$ docker service scale myapp_listener-service=1
```

```cmd
$ docker service ls | grep listener
z3ps5kl2nri0   myapp_listener-service         replicated   1/1        piatoss3612/listener-service:1.0.0
```

### 6. 서비스 업데이트

1.0.0에서 1.0.1로 업데이트된 logger 서비스의 이미지를 적용하는 경우

#### 1. 서비스 스케일링

업데이트는 순차적(rolling update)으로 진행되므로

업데이트 도중에 서비스가 중단되는 것을 방지하기 위해

logger 서비스를 스케일링하여 컨테이너 수를 증가시킨다

```cmd
$ docker service scale myapp_logger-service=2
```

#### 2. 이미지 업데이트

`docker service update` 명령으로 logger 서비스의 이미지를 1.0.1 버전으로 업데이트한다

```cmd
$ docker service update --image piatoss3612/logger-service:1.0.1 myapp_logger-service
```

#### 3. 다운그레이드

다운그레이드가 필요한 경우, 이미지 태그를 이전 버전으로 지정하고 업데이트 명령을 실행한다

```cmd
$ docker service update --image piatoss3612/logger-service:1.0.0 myapp_logger-service
```

### 7. Docker Swarm 종료

#### 1. 모든 서비스 제거

```cmd
$ docker stack rm myapp
$ docker service ls
ID        NAME      MODE      REPLICAS   IMAGE     PORTS
```

#### 2. Docker Swarm 떠나기

```cmd
$ docker swarm leave --force
```