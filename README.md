# Working with Microservices in Go

---

## Microservices

### 1. Broker Service

### 2. Authentication Service

### 3. Logger Service

### 4. Mail Service

### 5. Listener Service: AMQP with RabbitMQ

### 6. RPC(Remote Procedure Call)

### 7. gRPC

---

## Deploying Application to Cloud Platform 

- Linode, Hostinger Docker Swarm
- Hosting URL: https://swarm.piatoss.tech

---

## Deploying Application to Kubernetes

- Minikube, Docker Compose

---

## Testing

### Authentication Service

```cmd
$ go test -v .
=== RUN   Test_Authenticate
--- PASS: Test_Authenticate (0.00s)
=== RUN   Test_routes_exist
--- PASS: Test_routes_exist (0.00s)
PASS
ok      authentication/cmd/api  0.310s
```

### Reference

[Udemy: Working with Microservices in Go](https://www.udemy.com/course/working-with-microservices-in-go/)
