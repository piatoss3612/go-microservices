# Section10 Review

## 1. Kubernetes - Minikube Setup

### Minikube 설치 - chocolaty

[minikube start](https://minikube.sigs.k8s.io/docs/start/)

```bash
$ choco install minikube
```

### kubectl 설치 - chocolaty

[Install and Set Up kubectl on Windows](https://kubernetes.io/docs/tasks/tools/install-kubectl-windows/)

```bash
choco install kubernetes-cli
```

### Minikube 실행

2개의 노드로 구성된 minikube 클러스터 실행

```
$ minikube start --nodes 2
```

```
$ minikube status
minikube
type: Control Plane
host: Running
kubelet: Running
apiserver: Running
kubeconfig: Configured

minikube-m02
type: Worker
host: Running
kubelet: Running
```

### Minikube Dashboard 실행

minikube를 사용하면 사전 생성된 대시보드를 사용할 수 있다

```bash
$ minikube dashboard
```

## 2. Kubernetes Object - Deployment and Service

YAML 파일을 사용한 선언적(declarative) 쿠버네티스 오브젝트 생성 방법을 사용

### 1. MongoDB Deployment & Service 생성

`project/k8s/mongo.yaml` 파일 참고

```bash
$ kubectl apply -f k8s/mongo.yaml
deployment.apps/mongo created
service/mongo created
```

### 2. 생성된 mongo Pod 확인

```bash
$ kubectl get pods
NAME                     READY   STATUS              RESTARTS   AGE
mongo-869c89b6bd-5ch76   0/1     ContainerCreating   0          6s
```

### 3. 생성된 mongo Deployment 확인

```bash
$ kubectl get deploy
NAME    READY   UP-TO-DATE   AVAILABLE   AGE
mongo   1/1     1            1           4m17s
```

### 4. 생성된 mongo Service 확인

```bash
$ kubectl get svc
NAME         TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)     AGE
kubernetes   ClusterIP   10.96.0.1       <none>        443/TCP     5h26m
mongo        ClusterIP   10.108.82.194   <none>        27017/TCP   4m36s
```

### 5. RabbitMQ Deployment & Service 생성

`project/k8s/rabbit.yaml` 파일 참고

```bash
$ kubectl apply -f k8s/rabbit.yaml
deployment.apps/rabbitmq created
service/rabbitmq created
```

### 6. Broker Service Deployment & Service 생성

`project/k8s/broker.yaml` 파일 참고

```bash
$ kubectl apply -f k8s/broker.yaml
deployment.apps/broker-service created
service/broker-service created
```

### 7. Mailer Service Deployment & Service 생성

`project/k8s/mailer.yaml` 파일 참고

```bash
$ kubectl apply -f k8s/mailer.yaml
deployment.apps/mailer-service created
service/mailer-service created
```

### 8. Logger Service Deployment & Service 생성

`project/k8s/logger.yaml` 파일 참고

```bash
$ kubectl apply -f k8s/logger.yaml
deployment.apps/logger-service created
service/logger-service created
```

### 9. Listener Service Deployment & Service 생성

`project/k8s/listener.yaml` 파일 참고

```bash
$ kubectl apply -f k8s/listener.yaml
deployment.apps/listener-service created
service/listener-service created
```

### 10. Postgres on remote server

Postgres를 minikube 클러스터에서 실행하지 않고 docker compose로 실행하여

Kubernetes로 배포한 서비스와 다른 서버에서 동작중인 Postgres를 연동하는 상황에 대한 시뮬레이션을 진행한다

`project/postgres.yaml` 파일 참고

```bash
docker compose -f postgres.yaml up -d
```

### 11. Authentication Service Deployment & Service 생성

`project/k8s/authentication.yaml` 파일 참고

```bash
$ kubectl apply -f k8s/authentication.yaml
deployment.apps/authentication-service created
service/authentication-service created
```

### 12. Trouble Shooting: 이전에 사용한 DB가 정상 종료/제거되지 않아 Postgres에 연결할 수 없는 문제

1. Windows 명령 프롬프트에서 5432번 포트를 사용하고 있는 프로세스를 찾는다

```cmd
$ netstat -ano | grep 5432
```

2. 프로세스의 ID에 해당하는 프로그램이 무엇인지 확인한다

```cmd
$ tasklist /FI "PID eq 6980"

이미지 이름                    PID 세션 이름              세션#  메모리 사용
========================= ======== ================ =========== ============
com.docker.backend.exe        6980 Console                    1     46,744 K
```

3. 작업관리자로 들어가 사용하지 않고 있는 프로그램의 PID에 해당하는 프로세스를 삭제한다

4. Postgres와 정상적으로 연결이 되었는지 authentication 서비스의 로그를 확인한다

```bash
$ kubectl logs authentication-service-566bf6689b-zm4w5
2022/07/17 06:19:09 Starting authentication service...
2022/07/17 06:19:09 Connected to Postgres!
```

## 3. Broker 서비스에 Load Balancer 적용하기

### 1. 기존의 broker-service 서비스 제거

```bash
$ kubectl delete svc broker-service 
service "broker-service" deleted
```

### 2. 명령형으로 broker-service 로드 밸런서 생성

```bash
$ kubectl expose deployment broker-service --type=LoadBalancer --port=8080 --target-port=8080       
service/broker-service exposed
```

### 3. 생성된 로드 밸런서 확인

로스 밸런서를 사용하기 위한 외부 IP 주소가 부여되지 않은 것을 확인할 수 있다

```bash
$ kubectl get svc broker-service
NAME             TYPE           CLUSTER-IP    EXTERNAL-IP   PORT(S)          AGE
broker-service   LoadBalancer   10.104.4.87   <pending>     8080:32183/TCP   36s
```

### 4. minikube로 외부 IP 부여

```bash
$ minikube tunnel
✅  Tunnel successfully started

📌  NOTE: Please do not close this terminal as this process must stay alive for the tunnel to be accessible ...

🏃  broker-service 서비스의 터널을 시작하는 중
```

### 5. 로드 밸런서 외부 IP 확인

```bash
$ kubectl get svc broker-service 
NAME             TYPE           CLUSTER-IP    EXTERNAL-IP   PORT(S)          AGE
broker-service   LoadBalancer   10.104.4.87   127.0.0.1     8080:32183/TCP   92s
```

### 6. front-end 서비스를 로컬에서 실행하여 로드 밸런서 테스트

`front-end/cmd/web/main.go` 파일 수정

```go
func render(w http.ResponseWriter, t string) {

    ...

	var data struct {
		BrokerURL string
	}

	// data.BrokerURL = os.Getenv("BROKER_URL")
    data.BrokerURL = "http://localhost:8080"

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
```

```go
$ cd front-end
$ go run 
```

### 7. minikube 터널 종료 & Load Balancer 제거

```bash
$ kubectl delete svc broker-service 
service "broker-service" deleted
```

기존의 ClusterIP 타입의 broker-service 서비스도 살려놓는다

```bash
$ kubectl apply -f k8s/broker.yaml 
deployment.apps/broker-service unchanged
service/broker-service created
```

## 4. Nginx Ingress 적용하기

minikube 클러스터에 배포된 모든 마이크로 서비스는 ClusterIP 타입의 서비스를 통해

클러스터 내부에서는 서로 커뮤니케이션이 가능하지만,

외부에서 접근할 수 있는 방법이 존재하지 않는다

ingress는 Docker Swarm을 사용하여 애플리케이션을 배포했을 때 사용한 Caddy와 비슷한 역할을 한다

ingress는 클라이언트들이 클러스터 외부에서 접근할 수 있는 진입점을 제공한다

### 1. Frontend Service Deployment & Service 생성

`project/k8s/front-end.yaml` 파일 참고

```bash
$ kubectl apply -f k8s/front-end.yaml
deployment.apps/front-end created
service/front-end created
```

### 2. Nginx Ingress 컨트롤러 활성화

```bash
$ minikube addons enable ingress
💡  After the addon is enabled, please run "minikube tunnel" and your ingress resources would be available at "127.0.0.1"
    ▪ Using image k8s.gcr.io/ingress-nginx/controller:v1.2.1
    ▪ Using image k8s.gcr.io/ingress-nginx/kube-webhook-certgen:v1.1.1
    ▪ Using image k8s.gcr.io/ingress-nginx/kube-webhook-certgen:v1.1.1
🔎  Verifying ingress addon...
🌟  'ingress' 애드온이 활성화되었습니다
```

### 3. Nginx Ingress 생성

`project/ingress.yaml` 파일 참고

```bash
$ kubectl apply -f ingress.yaml
ingress.networking.k8s.io/my-ingress created
```
생성된 ingress 확인

```
$ kubectl get ing
NAME         CLASS   HOSTS                                ADDRESS        PORTS   AGE        
my-ingress   nginx   front-end.info,broker-service.info   192.168.49.2   80      71s  
```

### 4. hosts 파일 수정

아래의 내용 추가

```
127.0.0.1   front-end.info broker-service.info
```

### 5. Nginx Ingress 실행

```bash
$ minikube tunnel
🏃  my-ingress 서비스의 터널을 시작하는 중
```

### 6. Trouble Shooting: broker 서비스에 대한 요청이 http://front-end.info/ 로만 가는 문제

업데이트 중...