# Section09 Review

## 1. Linode

가상 서버(Virtual Private Server)를 제공하는 클라우딩 플랫폼

DigitalOcean, Vultr 등의 다른 옵션도 있지만 본 강의에서는 Linode를 사용한다

## 2. Linode 서버 생성

### 동일한 옵션의 Linode(Linux Server)를 2개 생성

##### manager node: node-1, worker node: node-2

Distribution: Ubuntu 22.04 LTS

Region: Singapore

Linode Plan: Shared CPU - Lonide 2GB

SSH Keys: 로컬에서 `ssh-keygen -t rsa` 명령어로 생성한 ssh 공개키 복사 붙여넣기

## 3. Linode 서버 설정

### 1. 로컬 환경에서 Linode 서버 ssh 접속

```bash
$ ssh root@[Linode 서버 IP 주소]
```
### 2. 새로운 사용자 추가 및 root 권한 부여

```bash
$ adduser piatoss
$ usermod -aG sudo piatoss
```

### 3. 방화벽 설정1: ssh, http, https

```bash
$ ufw allow ssh
$ ufw allow http
$ ufw allow https
```

### 4. 방화벽 설정2: Docker Swarm을 사용하기 위한 설정

```bash
$ ufw allow 2377/tcp
$ ufw allow 7946/tcp
$ ufw allow 7946/udp
$ ufw allow 4789/udp
$ ufw allow 8025/tcp
```

### 5. 방화벽 설정 적용

앞서 ssh 접속을 허용하지 않고 아래의 명령을 실행하면

서버에 접근할 수 없는 불상사가 발생할 수 있으므로 주의

```bash
$ ufw enable
```

### 6. 방화벽 상태 확인

```bash
$ ufw status
Status: active

To                         Action      From
--                         ------      ----
22/tcp                     ALLOW       Anywhere
80/tcp                     ALLOW       Anywhere
443                        ALLOW       Anywhere
2377/tcp                   ALLOW       Anywhere
7946/tcp                   ALLOW       Anywhere
7946/udp                   ALLOW       Anywhere
4789/udp                   ALLOW       Anywhere
8025/tcp                   ALLOW       Anywhere
22/tcp (v6)                ALLOW       Anywhere (v6)
80/tcp (v6)                ALLOW       Anywhere (v6)
443 (v6)                   ALLOW       Anywhere (v6)
2377/tcp (v6)              ALLOW       Anywhere (v6)
7946/tcp (v6)              ALLOW       Anywhere (v6)
7946/udp (v6)              ALLOW       Anywhere (v6)
4789/udp (v6)              ALLOW       Anywhere (v6)
8025/tcp (v6)              ALLOW       Anywhere (v6)
```

### 7. root 권한이 부여된 사용자로 접속

```bash
$ ssh piatoss@[Linode 서버 IP 주소]
```

### 8. Docker 설치

[Install Docker Engine on Ubuntu](https://docs.docker.com/engine/install/ubuntu/#install-using-the-repository)

#### 리포지토리 설정

```bash
$ sudo apt-get update
$ sudo apt-get install \
    ca-certificates \
    curl \
    gnupg \
    lsb-release
```

```bash
$ sudo mkdir -p /etc/apt/keyrings
$ curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
```

```bash
$ echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
```

#### Docker 엔진 설치

```bash
$ sudo apt update
$ sudo apt-get install docker-ce docker-ce-cli containerd.io docker-compose-plugin
```

#### Docker 설치 확인

```bash
$ which docker
/usr/bin/docker
```

## 4. Linode 서버 hostname 설정

### 첫 번째 서버

```bash
$ sudo hostnamectl set-hostname node-1
```

### 두 번째 서버

```bash
$ sudo hostnamectl set-hostname node-2
```

### hosts 파일 수정

```bash
$ sudo vi /etc/hosts
```

추가할 내용은 아래와 같다

```
[첫 번째 Linode의 IP 주소]  node-1
[두 번째 Linode의 IP 주소]  node-2
```

## 5. Hostinger

호스팅 제공자

1년간 사용할 수 있는 `piatoss.tech` 도메인을 구매하였다

### 1. DNS 설정

> A record: domain name과 IP 주소를 맵핑

<br>

<table>
<thead>
<tr>
<th>Type</th>
<th>Name</th>
<th>Points to</th>
<th>TTL</th>
</tr>
</thead>
<tbody>
<tr>
<td>A</td>
<td>node-1</td>
<td>첫 번째 Linode의 IP주소</td>
<td>default</td>
</tr>
<tr>
<td>A</td>
<td>node-2</td>
<td>두 번째 Linode의 IP주소</td>
<td>default</td>
</tr>
<tr>
<td>A</td>
<td>swarm</td>
<td>첫 번째 Linode의 IP주소</td>
<td>default</td>
</tr>
<tr>
<td>A</td>
<td>swarm</td>
<td>두 번째 Linode의 IP주소</td>
<td>default</td>
</tr>
</tbody>
</table>

#### swarm 이라는 이름의 A 레코드를 2개 생성한 이유

Docker Swarm을 사용해 Linode 서버에 배포한 애플리케이션을 어느 노드에서든 접근할 수 있도록

`swarm.piatoss.tech` 도메인을 단일 진입점으로 사용할 수 있다 

### 2. ping 체크

```bash
$ ping swarm.piatoss.tech

Pinging swarm.piatoss.tech [172.104.191.116] with 32 bytes of data:
Reply from 172.104.191.116: bytes=32 time=73ms TTL=52
Reply from 172.104.191.116: bytes=32 time=73ms TTL=52
Reply from 172.104.191.116: bytes=32 time=75ms TTL=52
Reply from 172.104.191.116: bytes=32 time=73ms TTL=52

Ping statistics for 172.104.191.116:
    Packets: Sent = 4, Received = 4, Lost = 0 (0% loss),
Approximate round trip times in milli-seconds:
    Minimum = 73ms, Maximum = 75ms, Average = 73ms
```

## 6. broker 서비스의 DNS 추가

> CNAME record: hostname과 hostname을 맵핑

<br>

<table>
<thead>
<tr>
<th>Type</th>
<th>Name</th>
<th>Points to</th>
<th>TTL</th>
</tr>
</thead>
<tbody>
<tr>
<td>CNAME</td>
<td>broker</td>
<td>swarm.piatoss.tech</td>
<td>default</td>
</tr>
</tbody>
</table>

## 7. Docker Swarm 실행

### 1. manager, worker node 설정

#### node-1: manager node 설정

```cmd
$ sudo docker swarm init --advertise-addr [첫 번째 Linode 서버의 IP 주소]
Swarm initialized: current node is now a manager.
```

#### node-2: worker node 설정

```cmd
$ sudo docker swarm join --token [manager node를 초기화할 때 생성된 토큰]
This node joined a swarm as a worker.
```

### 2. Caddy 설정

#### Caddyfile.production: 로컬 환경 -> 실제 배포 도메인으로 변경

```
swarm.piatoss.tech:80 {
	encode zstd gzip
	import static

	reverse_proxy  http://front-end:8081
}

broker.piatoss.tech:80 {
	reverse_proxy http://broker-service:8080
}
```

#### caddy.production.dockerfile

```dockerfile
FROM caddy:2.4.6-alpine

COPY Caddyfile.production /etc/caddy/Caddyfile
```

#### micro-caddy-production 이미지 빌드

```bash
$ docker build -f caddy.production.dockerfile -t piatoss3612/micro-caddy-production:1.0.0 .
```

#### Docker Hub에 이미지 저장

```bash
$ docker push piatoss3612/micro-caddy-production:1.0.0
```

#### swarm.production.yaml

```yaml
version: '3'

services:
  caddy:
    image: piatoss3612/micro-caddy-production:1.0.0 # 배포 버전으로 이미지 변경
    deploy:
      mode: replicated
      replicas: 1
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - caddy_data:/data
      - caddy_config:/config

  front-end:
    image: piatoss3612/front-end:1.0.1
    deploy:
      mode: replicated
      replicas: 1
    environment:
      BROKER_URL: "http://broker.piatoss.tech" # broker 서비스 도메인 변경
```

### 3. manager node: Docker Swarm 실행 준비

#### root 경로로 이동

```bash
$ cd /
```

#### swarm 디렉토리 생성

```bash
$ sudo mkdir swarm
```

#### 현재 사용자에게 swarm 디렉토리 소유 권한 부여

```bash
$ sudo chown piatoss:piatoss swarm/
```

#### volume으로 사용할 디렉토리 생성

```bash
$ cd swarm
$ mkdir caddy_data
$ mkdir caddy_config
$ mkdir db-data
$ mkdir db-data/mongo
$ mkdir db-data/postgres
```

#### swarm.yaml 파일 생성 및 swarm.production.yaml의 내용 붙여넣기

```bash
$ vim swarm.yaml
```

소문자 i (붙여넣기 모드) -> shift + insert 키로 붙여넣기 -> ctrl + c -> :wq + enter

### 4. Docker Swarm으로 배포

```bash
$ sudo docker stack deploy -c swarm.yaml myapp
Creating network myapp_default
Creating service myapp_logger-service
Creating service myapp_postgres
Creating service myapp_mongo
Creating service myapp_mailhog
Creating service myapp_listener-service
Creating service myapp_caddy
Creating service myapp_front-end
Creating service myapp_authentication-service
Creating service myapp_rabbitmq
Creating service myapp_mailer-service
Creating service myapp_broker-service
```

## 8. HTTPS 접속 설정

### 1. swarm.yaml 수정

```yaml
version: '3'

services:
  caddy:
    image: piatoss3612/micro-caddy-production:1.0.1 # https 연결을 사용하는 버전의 이미지
    deploy:
      mode: replicated
      replicas: 1
      placement:
        constraints:
          - node.hostname == node-1
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - caddy_data:/data
      - caddy_config:/config

  front-end:
    image: piatoss3612/front-end:1.0.1
    deploy:
      mode: replicated
      replicas: 1
    environment:
      BROKER_URL: "https://broker.piatoss.tech" # https 접속

    ...
```

### 2. Caddyfile.production 수정

```swarm.piatoss.tech {
	encode zstd gzip
	import static
    import securty # https 보안 설정 임포트

	reverse_proxy  http://front-end:8081
}

broker.piatoss.tech {
	reverse_proxy http://broker-service:8080
}
```

### 3. 이미지 빌드 및 Docker Hub에 저장

```bash
$ docker build -f caddy.production.dockerfile -t piatoss3612/micro-caddy-production:1.0.1 .
$ docker push piatoss3612/micro-caddy-production:1.0.1
```

### 4. 새로운 버전 배포

```bash
$ docker stack rm myapp
$ docker stack deploy -c swarm.yaml myapp
```

