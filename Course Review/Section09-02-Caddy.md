# Section09 Review

### 1. Caddy: Reverse Proxy 적용하기

#### front end와 broker 서비스의 외부 진입점 제거

front end와 broker 서비스를 외부로 노출시키는 포트를 설정하지 않으면

컨테이너 끼리는 소통이 가능하나, 외부 사용자는 서비스에 접근할 수 없게 된다

```yaml
version: '3'

services:
  front-end:
    image: piatoss3612/front-end:1.0.0
    deploy:
      mode: replicated
      replicas: 1

  broker-service:
    image: piatoss3612/broker-service:1.0.1
    deploy:
      mode: replicated
      replicas: 1
```

#### 리버스 프록시 서버?

리버스 프록시 서버는 클라이언트로부터 들어오는 요청을 받는 단일 진입점으로써

프록시 서버에 연결된 서비스로 요청을 라우팅해주는 로드 밸런서와 같은 역할을 한다


리버스 프록시는 연결된 서비스의 IP를 외부로 노출시키지 않고도

외부 요청을 받아 처리할 수 있으므로 외부의 공격으로부터 보안을 강화하기 위한 좋은 방법이다


Caddy는 Go언어로 작성된 서버에 대한 서버(server of servers)


본 강의에서는 Caddy 설정파일을 Docker 이미지로 빌드하고

다른 마이크로 서비스들에 대한 리버스 프록시로 사용한다

[Caddy: Reverse Proxy](https://caddyserver.com/docs/caddyfile/patterns#reverse-proxy)

#### Caddy 설정 파일: Caddyfile 

```
{
    email   piatoss3612@example.com
}

# 정적 파일 캐시 설정

(static) {
	@static {
		file
		path *.ico *.css *.js *.gif *.jpg *.jpeg *.png *.svg *.woff *.json
	}
	header @static Cache-Control max-age=5184000
}

# https 연결시 보안 설정

(security) {
	header {
		# enable HSTS
		Strict-Transport-Security max-age=31536000;
		# disable clients from sniffing the media type
		X-Content-Type-Options nosniff
		# keep referrer data off of HTTP connections
		Referrer-Policy no-referrer-when-downgrade
	}
}

# 리버스 프록시 설정

localhost:80 {
	encode zstd gzip
	import static

	reverse_proxy  http://front-end:8081
}

backend:80 {
	reverse_proxy http://broker-service:8080
}
```

#### Caddyfile과 Caddy 이미지로 Docker 이미지 빌드

```dockerfile
FROM caddy:2.4.6-alpine

COPY Caddyfile /etc/caddy/Caddyfile
```

```cmd
$ docker build -f caddy.dockerfile -t piatoss3612/micro-caddy:1.0.0 .
```

#### Docker Hub에 저장

```cmd
$ docker push piatoss3612/micro-caddy:1.0.0
```

#### hosts 파일 수정

로컬 환경에서 `localhost:80`, `backend:80` 2개의 경로에 대한 요청을

리버스 프록시 서버가 정상적으로 처리하는지 테스트하기 위해 호스트 파일을 수정해야 한다 

`127.0.0.1       localhost backend` 추가

[windows](https://www.thewindowsclub.com/hosts-file-in-windows)


#### swarm.yaml 파일에 Caddy 추가

```yaml
version: '3'

services:
  caddy:
    image: piatoss3612/micro-caddy:1.0.0
    deploy:
      mode: replicated
      replicas: 1
    ports:
      - "80:80" # http
      - "443:443" # https
      
    volumes: # 인증서나 설정 파일이 저장될 볼륨을 마운트한다
      - caddy_data:/data
      - caddy_config:/config

volumes:
  caddy_data:
    external: true
  caddy_config:
    external: true
```

#### Docker Swarm으로 Caddy를 포함하여 로컬 환경에 배포

```cmd
$ docker swarm init
$ docker stack deploy -c swarm.yaml myapp
```