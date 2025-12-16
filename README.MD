# 🧩 ScoreHub — Credit Event Notification System

**ScoreHub** is a demonstration microservice platform that models a simplified credit monitoring system.  
It showcases a modern Go backend architecture featuring **gRPC communication**, **Kafka event streams**, and **PostgreSQL persistence**.

Users receive notifications about credit score changes or related credit report activities.  
The system is designed to illustrate production-ready service interactions, event-driven processing, and observability patterns.

---

## 🚀 Features

- **Microservice architecture** (4 services)
- **gRPC communication** with optional REST gateway
- **Event-driven flow** via Apache Kafka
- **PostgreSQL** persistence for users and notifications
- **Docker Compose / Kubernetes** deployment support
- **Prometheus metrics** and **structured logging**
- Clean, idiomatic Go layout (`cmd`, `pkg`, `deploy`, `docs`)

---

## 🏗 Architecture Overview

### 1️⃣ `user-service`
Handles user management and exposes CRUD operations.

- gRPC API + REST gateway
- PostgreSQL `users` table
- Example RPCs: `CreateUser`, `GetUser`, `ListUsers`

---

### 2️⃣ `event-service`
Simulates external credit events and publishes them to Kafka.

- Publishes JSON messages to topic `score_events`
- Optional REST auth via header `X-API-Key` (set `API_KEY` env var on the service); omit the header if auth is disabled
- Example payload:
  ```json
  {
    "user_id": 42,
    "new_score": 730,
    "change": 15
  }
  ```
- Example RPC:
  ```proto
  service EventService {
    rpc SendScoreEvent(ScoreEventRequest) returns (EventAck);
  }
  ```

---

### 3️⃣ `notification-service`
Consumes score updates and generates user notifications.

- Subscribes to Kafka topic `score_events`
- Calls `user-service` via gRPC to enrich events
- Stores results in `notifications` table (PostgreSQL)
- Business logic:
    - If score change `> +10` → create a “congratulation” message
    - Otherwise → record a silent update

Example schema:
```sql
CREATE TABLE notifications (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    message TEXT NOT NULL,
    score_change INT,
    created_at TIMESTAMP DEFAULT now()
);
```

---

### 4️⃣ `email-service` *(optional)*
Consumes Kafka topic `notifications` and simulates sending emails.

- Logs messages to stdout (mock SMTP)
- Demonstrates asynchronous fan-out and background processing
- Example output:
  ```
  [2025-11-02 10:32:14] Sending congratulation email to user 42: "Your score increased by 15 points!"
  ```

---

## 🧩 High-Level Diagram

```
             +----------------+
             |  event-service |
             | (score events) |
             +-------+--------+
                     |
                     | Kafka: score_events
                     v
             +---------------------+
             | notification-service|
             | consumes events     |
             +----+-----------+----+
                  |           |
                  | gRPC      | Kafka: notifications
                  v           v
        +----------------+   +----------------+
        |  user-service   |   |  email-service |
        | PostgreSQL (users)| | mock SMTP/logs |
        +----------------+   +----------------+
```

---

## 📁 Repository Structure

```
/scorehub
  /cmd
    /user-service
    /event-service
    /notification-service
    /email-service
  /pkg
    /user
    /event
    /notification
    /email
    /common
  /deploy
    docker-compose.yml
    /k8s
      user-deployment.yaml
      event-deployment.yaml
      notification-deployment.yaml
      email-deployment.yaml
      postgres-statefulset.yaml
      kafka-deployment.yaml
  /docs
    architecture-diagram.png
    README.md
  Makefile
  go.mod
  go.sum
```

---

## 🧠 Technologies

| Category | Tools |
|-----------|-------|
| Language | Go 1.23+ |
| RPC | gRPC + grpc-gateway |
| Messaging | Apache Kafka |
| Database | PostgreSQL |
| Infrastructure | Docker Compose, Kubernetes |
| Observability | Prometheus, Grafana |
| CI/CD | GitHub Actions |
| Config | `cleanenv` (ENV-based configuration) |
| Logging | `zap` (structured logging) |

---

## ⚙️ Local Development

### 1️⃣ Prerequisites
- Docker & Docker Compose
- Go 1.23+
- Makefile support (`make` command)

### 2️⃣ Run all services
```bash
make up
```
This command runs PostgreSQL, Kafka, and all four microservices via Docker Compose.

### 3️⃣ Generate protobufs
```bash
make proto
```

### 4️⃣ Run tests
```bash
make test
```

---

## 🌍 Environment Variables

Each service loads configuration via environment variables using [`cleanenv`](https://github.com/ilyakaznacheev/cleanenv).

Example for `user-service`:
```bash
HTTP_PORT=8080
GRPC_PORT=9090
DB_DSN=postgres://scorehub:password@postgres:5432/scorehub?sslmode=disable
LOG_LEVEL=debug
```

`event-service` REST auth:
- Set `API_KEY=<secret>` to require clients to send `X-API-Key` with that value on `/api/v1/score-events`.
- Leave `API_KEY` unset/empty to allow requests without the header (useful for local/dev).

---

## 📊 Metrics & Observability

- Each service exposes `/metrics` endpoint for Prometheus.
- Use Grafana dashboards from `/deploy/grafana/` to visualize key metrics (latency, Kafka lag, DB performance).
- Logging follows structured JSON format with correlation IDs for traceability.

---

## 🧩 Makefile Commands

| Command | Description |
|----------|-------------|
| `make up` | Start all services via Docker Compose |
| `make down` | Stop and clean up |
| `make proto` | Generate protobuf files |
| `make test` | Run tests |
| `make lint` | Run linters |
| `make build` | Build binaries for all services |

---

## 🧭 Future Improvements

- Add OpenTelemetry tracing across services
- Add authentication/authorization service
- Implement retry and dead-letter queue for Kafka
- Integrate CI/CD deployment to Kubernetes
- Enhance unit/integration test coverage with `testcontainers-go`

---

## 🧑‍💻 Author

**Evgenii Morenkov** — Senior Golang Engineer  
[LinkedIn](https://www.linkedin.com/in/emorenkov/) · [GitHub](https://github.com/emorenkov) · [Email](mailto:yarad8613@gmail.com)

---

## ⚖️ License
MIT License — feel free to fork and adapt.
