# Лабораторна робота №2
**Тема:** Робота з брокерами повідомлень – RabbitMQ

---

## Мета роботи

Розробити розподілений застосунок, розбитий на два мікросервіси, що взаємодіють через брокер повідомлень RabbitMQ. Перший сервіс приймає gRPC-запити від клієнта, валідує дані та публікує події до черги. Другий сервіс споживає повідомлення з черги, зберігає дані у власній базі даних та збирає статистику обробки подій.

---

## Теоретичні відомості

**Брокер повідомлень** — програмний посередник, що забезпечує асинхронну передачу повідомлень між сервісами. Відправник (producer) публікує повідомлення, не чекаючи відповіді; отримувач (consumer) опрацьовує їх у власному темпі. Це забезпечує слабку зв'язаність (loose coupling) і незалежне масштабування сервісів.

**RabbitMQ** — популярний брокер повідомлень, що реалізує протокол AMQP (Advanced Message Queuing Protocol). Основні поняття:

| Концепція | Опис |
|-----------|------|
| **Producer** | Сервіс, що публікує повідомлення |
| **Consumer** | Сервіс, що читає та обробляє повідомлення |
| **Queue** | Буфер, в якому зберігаються повідомлення до обробки |
| **Exchange** | Маршрутизатор повідомлень від producer до queue |
| **Binding** | Зв'язок між exchange і queue |
| **Ack / Nack** | Підтвердження успішної (або невдалої) обробки повідомлення |

**Режими підтвердження:**
- **Auto-ack** — повідомлення видаляється з черги одразу після доставки.
- **Manual ack** — повідомлення видаляється тільки після явного підтвердження споживачем. Якщо обробка завершилась помилкою — видається `nack`, і повідомлення повертається в чергу.

**Durable queue + Persistent messages** — черга та повідомлення зберігаються на диску й виживають після перезапуску RabbitMQ.

**Архітектура даної роботи:**

```
Client
  │ gRPC
  ▼
┌─────────────────────┐       JSON       ┌──────────────────────┐
│    api-service      │ ──────────────▶  │   worker-service     │
│  gRPC :50052        │  device.events   │  RabbitMQ consumer   │
│  SQLite (api.db)    │     (queue)      │  gRPC :50053 (stats) │
│  RabbitMQ publisher │                  │  SQLite (worker.db)  │
└─────────────────────┘                  └──────────────────────┘
```

---

## Завдання

1. Розгорнути RabbitMQ за допомогою Docker;
2. Розробити `api-service` — gRPC-сервер для CRUD операцій над комп'ютерними комплектуючими (Device), що публікує події до черги RabbitMQ та зберігає дані у власній SQLite БД;
3. Розробити `worker-service` — сервіс-споживач черги RabbitMQ, що зберігає дані у власній SQLite БД та збирає статистику подій (created / updated / deleted) у розрізі груп пристроїв;
4. Додати до `worker-service` gRPC API для перегляду статистики та списку пристроїв;
5. Протестувати повний потік даних: gRPC запит → черга → обробка → статистика.

---

## Хід роботи

### 1. Розгортання RabbitMQ через Docker

Для швидкого запуску RabbitMQ використано офіційний образ `rabbitmq:3-management`, що включає веб-інтерфейс управління.

```yaml
# docker-compose.yml (фрагмент)
rabbitmq:
  image: rabbitmq:3-management
  ports:
    - "5672:5672"    # AMQP
    - "15672:15672"  # Management UI
  environment:
    RABBITMQ_DEFAULT_USER: guest
    RABBITMQ_DEFAULT_PASS: guest
  healthcheck:
    test: ["CMD", "rabbitmq-diagnostics", "ping"]
    interval: 10s
    timeout: 5s
    retries: 5
```

Запуск усього стеку:

```bash
$ docker compose up -d

[+] Running 3/3
 ✔ Container lab2-rabbitmq-1        Healthy
 ✔ Container lab2-api-service-1     Started
 ✔ Container lab2-worker-service-1  Started
```

Веб-інтерфейс RabbitMQ Management UI доступний за адресою `http://localhost:15672` (guest / guest).

---

### 2. Архітектура та структура проекту

```
lab2/
├── docker-compose.yml
├── api-service/                  # gRPC API + RabbitMQ publisher
│   ├── proto/devices.proto       # специфікація DeviceService
│   ├── gen/devicepb/             # згенерований код
│   ├── internal/
│   │   ├── model/                # модель Device
│   │   ├── repo/                 # CRUD через GORM
│   │   ├── service/              # бізнес-логіка
│   │   ├── publisher/            # публікація подій до RabbitMQ
│   │   └── grpcserver/           # gRPC хендлери
│   └── cmd/main.go
└── worker-service/               # RabbitMQ consumer + статистика
    ├── proto/stats.proto         # специфікація StatsService
    ├── gen/statspb/              # згенерований код
    ├── internal/
    │   ├── model/                # Device, EventStat, LastProcessed
    │   ├── repo/                 # device_repo, stats_repo
    │   ├── service/              # обробка повідомлень
    │   ├── consumer/             # споживач черги RabbitMQ
    │   └── grpcserver/           # gRPC хендлери (stats)
    └── cmd/main.go
```

---

### 3. Proto-специфікація

**api-service** — `proto/devices.proto`:

```protobuf
service DeviceService {
  rpc CreateDevice (CreateDeviceRequest) returns (DeviceResponse);
  rpc UpdateDevice (UpdateDeviceRequest) returns (DeviceResponse);
  rpc DeleteDevice (DeleteDeviceRequest) returns (DeleteDeviceResponse);
  rpc ListDevices  (ListDevicesRequest)  returns (ListDevicesResponse);
  rpc GetDevice    (GetDeviceRequest)    returns (DeviceResponse);
}

message DeviceType {
  bool            peripheral  = 1;
  int32           power_watts = 2;
  bool            has_cooler  = 3;
  string          group       = 4;  // "io" | "multimedia"
  repeated string ports       = 5;  // "COM", "USB", "LPT"
}

message Device {
  uint64     id          = 1;
  string     name        = 2;
  string     origin      = 3;
  double     price       = 4;
  bool       critical    = 5;
  DeviceType device_type = 6;
}
```

**worker-service** — `proto/stats.proto`:

```protobuf
service StatsService {
  rpc GetStats    (GetStatsRequest)    returns (StatsResponse);
  rpc ListDevices (ListDevicesRequest) returns (ListDevicesResponse);
  rpc GetDevice   (GetDeviceRequest)   returns (Device);
}

message StatsResponse {
  int64             total_created     = 1;
  int64             total_updated     = 2;
  int64             total_deleted     = 3;
  repeated GroupCount by_group        = 4;
  string            last_processed_at = 5;
}
```

Генерація коду виконується скриптами `generate.sh` у кожному сервісі:

```bash
$ cd api-service && bash generate.sh
proto generation done

$ cd ../worker-service && bash generate.sh
proto generation done
```

---

### 4. Формат повідомлень черги

Черга `device.events` оголошена як **durable**, повідомлення — **persistent**. Формат JSON:

```json
{
  "event": "created",
  "device": {
    "ID": 4,
    "Name": "RTX 4090",
    "Origin": "Taiwan",
    "Price": 1599.99,
    "Critical": true,
    "Peripheral": false,
    "PowerWatts": 450,
    "HasCooler": true,
    "Group": "multimedia",
    "Ports": ["PCIe"]
  }
}
```

Поле `event` приймає значення: `"created"`, `"updated"`, `"deleted"`.

---

### 5. Реалізація api-service

При кожній write-операції сервіс:
1. Зберігає зміни у власній SQLite БД (`api.db`);
2. Публікує JSON-повідомлення до черги `device.events`.

```go
// internal/publisher/publisher.go
func (p *Publisher) Publish(ctx context.Context, eventType string, device any) error {
    body, _ := json.Marshal(Event{Event: eventType, Device: device})
    return p.ch.PublishWithContext(ctx, "", queueName, false, false, amqp.Publishing{
        ContentType:  "application/json",
        DeliveryMode: amqp.Persistent,
        Body:         body,
    })
}
```

При старті відбувається спроба підключення до RabbitMQ (5 спроб, інтервал 2 с) та автоматичне заповнення БД тестовими даними:

```
$ go run ./cmd
2025/04/06 12:00:01 seeded 3 devices
2025/04/06 12:00:01 gRPC server listening on :50052
```

---

### 6. Реалізація worker-service

Споживач черги використовує **manual ack**: повідомлення підтверджується лише після успішної обробки. У разі помилки видається `nack`.

```go
// internal/consumer/consumer.go (фрагмент)
for msg := range msgs {
    if err := w.svc.Process(msg.Body); err != nil {
        log.Printf("process error: %v", err)
        msg.Nack(false, true)
        continue
    }
    msg.Ack(false)
}
```

Обробка повідомлення у `worker_service.go`:
- `created` / `updated` → upsert пристрою в БД + інкремент статистики;
- `deleted` → видалення пристрою + інкремент статистики.

Статистика зберігається в таблиці `event_stats` з ключем `(event_type, group)`:

```
event_type | group      | count
-----------+------------+------
created    | io         |     2
created    | multimedia |     1
updated    | multimedia |     1
deleted    | io         |     1
```

---

### 7. Тестування повного потоку

Для тестування використано `grpcurl`. Скрипт `test_grpc.sh` автоматизує весь флоу.

**Крок 1 — Список пристроїв (заповнено при старті):**

```bash
$ grpcurl -plaintext -d '{}' localhost:50052 devicepb.DeviceService/ListDevices

{
  "devices": [
    { "id": "1", "name": "Logitech MX Keys", "origin": "USA", "price": 109.99,
      "deviceType": { "peripheral": true, "powerWatts": 5, "group": "io", "ports": ["USB"] } },
    { "id": "2", "name": "Corsair RM850x", "origin": "USA", "price": 139.99, "critical": true,
      "deviceType": { "powerWatts": 850, "group": "io", "ports": ["COM"] } },
    { "id": "3", "name": "Razer Kraken", "origin": "China", "price": 79.99,
      "deviceType": { "peripheral": true, "powerWatts": 3, "group": "multimedia", "ports": ["USB","COM"] } }
  ]
}
```

**Крок 2 — Створення нового пристрою:**

```bash
$ grpcurl -plaintext -d '{
  "name": "RTX 4090", "origin": "Taiwan", "price": 1599.99, "critical": true,
  "device_type": { "power_watts": 450, "has_cooler": true, "group": "multimedia", "ports": ["PCIe"] }
}' localhost:50052 devicepb.DeviceService/CreateDevice

{
  "device": {
    "id": "4", "name": "RTX 4090", "origin": "Taiwan", "price": 1599.99, "critical": true,
    "deviceType": { "powerWatts": 450, "hasCooler": true, "group": "multimedia", "ports": ["PCIe"] }
  }
}
```

**Крок 3 — Оновлення пристрою:**

```bash
$ grpcurl -plaintext -d '{
  "id": 4, "name": "RTX 4090 Ti", "origin": "Taiwan", "price": 1799.99, "critical": true,
  "device_type": { "power_watts": 500, "has_cooler": true, "group": "multimedia", "ports": ["PCIe"] }
}' localhost:50052 devicepb.DeviceService/UpdateDevice

{ "device": { "id": "4", "name": "RTX 4090 Ti", "price": 1799.99, ... } }
```

**Крок 4 — Видалення пристрою:**

```bash
$ grpcurl -plaintext -d '{"id": 4}' localhost:50052 devicepb.DeviceService/DeleteDevice

{ "success": true }
```

**Крок 5 — Перегляд статистики через worker-service:**

```bash
$ grpcurl -plaintext -d '{}' localhost:50053 statspb.StatsService/GetStats

{
  "totalCreated": "4",
  "totalUpdated": "1",
  "totalDeleted": "1",
  "byGroup": [
    { "group": "io",         "count": "3" },
    { "group": "multimedia", "count": "3" }
  ],
  "lastProcessedAt": "2025-04-06T12:00:15Z"
}
```

**Крок 6 — Список пристроїв у БД worker-service:**

```bash
$ grpcurl -plaintext -d '{}' localhost:50053 statspb.StatsService/ListDevices

{
  "devices": [
    { "id": "1", "name": "Logitech MX Keys", ... },
    { "id": "2", "name": "Corsair RM850x",   ... },
    { "id": "3", "name": "Razer Kraken",     ... }
  ]
}
```

---

## Лістинг програми

### api-service/proto/devices.proto
```protobuf
syntax = "proto3";
package devicepb;
option go_package = "github.com/necutya/decentrilized_apps/lab2/api-service/gen/devicepb";

service DeviceService {
  rpc CreateDevice (CreateDeviceRequest) returns (DeviceResponse);
  rpc GetDevice    (GetDeviceRequest)    returns (DeviceResponse);
  rpc UpdateDevice (UpdateDeviceRequest) returns (DeviceResponse);
  rpc DeleteDevice (DeleteDeviceRequest) returns (DeleteDeviceResponse);
  rpc ListDevices  (ListDevicesRequest)  returns (ListDevicesResponse);
}

message DeviceType {
  bool peripheral = 1; int32 power_watts = 2;
  bool has_cooler = 3; string group = 4;
  repeated string ports = 5;
}
message Device {
  uint64 id = 1; string name = 2; string origin = 3;
  double price = 4; bool critical = 5; DeviceType device_type = 6;
}
message CreateDeviceRequest {
  string name = 1; string origin = 2; double price = 3;
  bool critical = 4; DeviceType device_type = 5;
}
message UpdateDeviceRequest {
  uint64 id = 1; string name = 2; string origin = 3;
  double price = 4; bool critical = 5; DeviceType device_type = 6;
}
message GetDeviceRequest    { uint64 id = 1; }
message DeleteDeviceRequest { uint64 id = 1; }
message DeviceResponse      { Device device = 1; }
message DeleteDeviceResponse { bool success = 1; }
message ListDevicesRequest  {}
message ListDevicesResponse { repeated Device devices = 1; }
```

### api-service/internal/publisher/publisher.go
```go
package publisher

import (
    "context"
    "encoding/json"
    amqp "github.com/rabbitmq/amqp091-go"
)

const queueName = "device.events"

type Publisher struct {
    conn *amqp.Connection
    ch   *amqp.Channel
}

func New(url string) (*Publisher, error) {
    conn, err := amqp.Dial(url)
    if err != nil { return nil, err }
    ch, err := conn.Channel()
    if err != nil { conn.Close(); return nil, err }
    _, err = ch.QueueDeclare(queueName, true, false, false, false, nil)
    if err != nil { ch.Close(); conn.Close(); return nil, err }
    return &Publisher{conn: conn, ch: ch}, nil
}

func (p *Publisher) Publish(ctx context.Context, eventType string, device any) error {
    body, _ := json.Marshal(struct {
        Event  string `json:"event"`
        Device any    `json:"device"`
    }{Event: eventType, Device: device})
    return p.ch.PublishWithContext(ctx, "", queueName, false, false, amqp.Publishing{
        ContentType:  "application/json",
        DeliveryMode: amqp.Persistent,
        Body:         body,
    })
}

func (p *Publisher) Close() { p.ch.Close(); p.conn.Close() }
```

### worker-service/internal/consumer/consumer.go
```go
package consumer

import (
    "log"
    amqp "github.com/rabbitmq/amqp091-go"
    "github.com/necutya/decentrilized_apps/lab2/worker-service/internal/service"
)

const queueName = "device.events"

type Consumer struct {
    conn *amqp.Connection
    ch   *amqp.Channel
    svc  *service.WorkerService
}

func New(url string, svc *service.WorkerService) (*Consumer, error) {
    conn, err := amqp.Dial(url)
    if err != nil { return nil, err }
    ch, err := conn.Channel()
    if err != nil { conn.Close(); return nil, err }
    _, err = ch.QueueDeclare(queueName, true, false, false, false, nil)
    if err != nil { ch.Close(); conn.Close(); return nil, err }
    return &Consumer{conn: conn, ch: ch, svc: svc}, nil
}

func (c *Consumer) Consume() error {
    msgs, err := c.ch.Consume(queueName, "", false, false, false, false, nil)
    if err != nil { return err }
    log.Printf("consuming queue: %s", queueName)
    for msg := range msgs {
        if err := c.svc.Process(msg.Body); err != nil {
            log.Printf("process error: %v", err)
            msg.Nack(false, true)
            continue
        }
        msg.Ack(false)
    }
    return nil
}

func (c *Consumer) Close() { c.ch.Close(); c.conn.Close() }
```

### worker-service/internal/service/worker_service.go
```go
package service

import (
    "encoding/json"
    "log"
    "time"

    "github.com/necutya/decentrilized_apps/lab2/worker-service/internal/model"
    "github.com/necutya/decentrilized_apps/lab2/worker-service/internal/repo"
)

type WorkerService struct {
    deviceRepo *repo.DeviceRepo
    statsRepo  *repo.StatsRepo
}

func New(dr *repo.DeviceRepo, sr *repo.StatsRepo) *WorkerService {
    return &WorkerService{deviceRepo: dr, statsRepo: sr}
}

func (w *WorkerService) Process(body []byte) error {
    var msg model.DeviceMessage
    if err := json.Unmarshal(body, &msg); err != nil { return err }

    p := msg.Device
    d := &model.Device{
        ID: p.ID, Name: p.Name, Origin: p.Origin, Price: p.Price,
        Critical: p.Critical, Peripheral: p.Peripheral, PowerWatts: p.PowerWatts,
        HasCooler: p.HasCooler, Group: p.Group, Ports: p.Ports,
    }

    switch msg.Event {
    case "created", "updated":
        if err := w.deviceRepo.Upsert(d); err != nil { return err }
    case "deleted":
        if err := w.deviceRepo.Delete(d.ID); err != nil { return err }
    default:
        log.Printf("unknown event: %s", msg.Event)
        return nil
    }

    w.statsRepo.IncrementStat(msg.Event, d.Group)
    w.statsRepo.UpdateLastProcessed(time.Now().UTC())
    log.Printf("processed event=%s device_id=%d", msg.Event, d.ID)
    return nil
}
```

### worker-service/internal/repo/stats_repo.go
```go
package repo

import (
    "time"
    "github.com/necutya/decentrilized_apps/lab2/worker-service/internal/model"
    "gorm.io/gorm"
    "gorm.io/gorm/clause"
)

type StatsRepo struct{ db *gorm.DB }
func NewStatsRepo(db *gorm.DB) *StatsRepo { return &StatsRepo{db: db} }

func (r *StatsRepo) IncrementStat(eventType, group string) error {
    return r.db.Transaction(func(tx *gorm.DB) error {
        var stat model.EventStat
        err := tx.Where("event_type = ? AND `group` = ?", eventType, group).First(&stat).Error
        if err == gorm.ErrRecordNotFound {
            return tx.Create(&model.EventStat{EventType: eventType, Group: group, Count: 1}).Error
        } else if err != nil { return err }
        return tx.Model(&stat).Update("count", stat.Count+1).Error
    })
}

func (r *StatsRepo) UpdateLastProcessed(t time.Time) error {
    return r.db.Clauses(clause.OnConflict{UpdateAll: true}).
        Create(&model.LastProcessed{ID: 1, ProcessedAt: t}).Error
}

type Summary struct {
    TotalCreated, TotalUpdated, TotalDeleted int64
    ByGroup         map[string]int64
    LastProcessedAt time.Time
}

func (r *StatsRepo) GetSummary() (*Summary, error) {
    var stats []model.EventStat
    if err := r.db.Find(&stats).Error; err != nil { return nil, err }
    s := &Summary{ByGroup: make(map[string]int64)}
    for _, st := range stats {
        switch st.EventType {
        case "created": s.TotalCreated += st.Count
        case "updated": s.TotalUpdated += st.Count
        case "deleted": s.TotalDeleted += st.Count
        }
        if st.Group != "" { s.ByGroup[st.Group] += st.Count }
    }
    var lp model.LastProcessed
    if r.db.First(&lp, 1).Error == nil { s.LastProcessedAt = lp.ProcessedAt }
    return s, nil
}
```

---

## Висновки

У ході виконання лабораторної роботи розроблено розподілений застосунок для управління каталогом комп'ютерних комплектуючих (Device) з використанням брокера повідомлень RabbitMQ.

Застосунок складається з двох незалежних мікросервісів. `api-service` виступає точкою входу: приймає gRPC-запити, валідує дані, зберігає їх у власній SQLite БД та публікує асинхронні JSON-події до черги `device.events`. `worker-service` споживає ці події, реплікує дані у власну БД та накопичує статистику — кількість операцій створення, оновлення та видалення у розрізі груп пристроїв. Крім того, `worker-service` надає власний gRPC API для запиту статистики та перегляду пристроїв.

Використання брокера повідомлень забезпечує **слабку зв'язаність** сервісів: `api-service` не залежить від доступності `worker-service` — у разі його недоступності повідомлення накопичуються в черзі й обробляються після відновлення. Механізм **manual ack** гарантує, що жодне повідомлення не буде втрачено через збій під час обробки.

Порівняно з синхронним підходом (наприклад, прямий gRPC-виклик між сервісами), брокер повідомлень забезпечує більшу стійкість до відмов, природне горизонтальне масштабування споживачів та ізоляцію відповідальності між сервісами.
