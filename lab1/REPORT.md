# Лабораторна робота №1
**Тема:** Технології віддаленого виклику процедур (RPC) – gRPC

---

## Мета роботи

Розробити клієнт-серверний застосунок з використанням технології gRPC: описати контракт взаємодії у форматі Protocol Buffers, згенерувати серверний та клієнтський код, реалізувати сервіс бронювання квитків на події з автентифікацією на основі JWT.

---

## Теоретичні відомості

**gRPC** (Google Remote Procedure Call) — високопродуктивний відкритий RPC-фреймворк, розроблений компанією Google. Він дозволяє клієнтам викликати методи серверного застосунку так само, як якби вони були локальними функціями.

Основні компоненти gRPC:

- **Protocol Buffers (protobuf)** — мова опису інтерфейсу (IDL) та бінарний формат серіалізації. Контракт описується у `.proto`-файлі, з якого генерується код для обраної мови.
- **Сервіс** — набір RPC-методів, описаних у `.proto`. Кожен метод приймає одне повідомлення і повертає одне повідомлення (або потік у стрімінгових режимах).
- **Канал** — з'єднання між клієнтом і сервером поверх HTTP/2. HTTP/2 забезпечує мультиплексування запитів, стиснення заголовків і двонаправлені потоки.
- **Stub (заглушка)** — згенерований клієнтський код, який приховує мережеву взаємодію за звичайними викликами методів.
- **Metadata** — аналог HTTP-заголовків у gRPC; використовується для передачі токенів автентифікації та іншої службової інформації.
- **gRPC Reflection** — опціональний серверний сервіс, що дозволяє клієнтам (наприклад, `grpcurl`) динамічно отримувати схему API без `.proto`-файлу.

Типи взаємодії у gRPC:

| Тип | Опис |
|-----|------|
| Unary RPC | Один запит → одна відповідь (використано в даній роботі) |
| Server streaming | Один запит → потік відповідей |
| Client streaming | Потік запитів → одна відповідь |
| Bidirectional streaming | Потік запитів ↔ потік відповідей |

---

## Завдання

1. Встановити компілятор `protoc` та Go-плагіни для генерації коду;
2. Описати контракт сервісу у `.proto`-файлі: сервіси автентифікації (`AuthService`) та роботи з квитками (`TicketService`);
3. Згенерувати Go-код із `.proto`-специфікації;
4. Реалізувати gRPC-сервер з підтримкою JWT-автентифікації та збереженням даних у SQLite;
5. Зареєструвати сервіс Reflection для можливості тестування через `grpcurl`;
6. Протестувати повний потік взаємодії: реєстрація → логін → перегляд подій → бронювання → скасування.

---

## Хід роботи

### 1. Встановлення protoc

Завантажено бінарний дистрибутив `protoc` версії 25.3 для Linux з офіційного репозиторію:

```bash
wget https://github.com/protocolbuffers/protobuf/releases/download/v25.3/protoc-25.3-linux-x86_64.zip
unzip protoc-25.3-linux-x86_64.zip -d $HOME/.local
export PATH="$HOME/.local/bin:$PATH"
```

Перевірка встановлення:

```
$ protoc --version
libprotoc 25.3
```

Встановлено Go-плагіни для генерації коду:

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.34.2
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.4.0
```

---

### 2. Опис специфікації (proto-файл)

Створено файл `proto/tickets.proto`, що описує два сервіси та всі необхідні типи повідомлень.

**`AuthService`** — реєстрація та логін користувача:

```protobuf
service AuthService {
  rpc Register (RegisterRequest) returns (AuthResponse);
  rpc Login    (LoginRequest)    returns (AuthResponse);
}
```

**`TicketService`** — робота з подіями та бронюваннями (потребує JWT-токена в metadata):

```protobuf
service TicketService {
  rpc ListEvents     (ListEventsRequest)    returns (ListEventsResponse);
  rpc GetEvent       (GetEventRequest)      returns (Event);
  rpc BookTicket     (BookTicketRequest)    returns (Booking);
  rpc ListMyBookings (ListBookingsRequest)  returns (ListBookingsResponse);
  rpc CancelBooking  (CancelBookingRequest) returns (CancelBookingResponse);
}
```

Повна специфікація знаходиться у файлі `proto/tickets.proto`.

---

### 3. Генерація Go-коду

Створено скрипт `generate.sh`, що запускає `protoc` з двома плагінами:

```bash
protoc \
  --go_out=gen --go_opt=paths=source_relative \
  --go-grpc_out=gen --go-grpc_opt=paths=source_relative \
  --proto_path=proto \
  proto/tickets.proto
```

Результат генерації — два файли у `gen/ticketpb/`:

- `tickets.pb.go` — структури повідомлень (серіалізація/десеріалізація);
- `tickets_grpc.pb.go` — інтерфейси сервера та заглушки клієнта.

```
$ bash generate.sh
proto generation done

$ ls gen/ticketpb/
tickets.pb.go  tickets_grpc.pb.go
```

---

### 4. Реалізація сервера

Проект структурований за принципом розподілу відповідальностей:

```
lab1/
├── cmd/main.go                  # точка входу, збірка залежностей
├── proto/tickets.proto          # специфікація
├── gen/ticketpb/                # згенерований код
├── internal/
│   ├── grpcserver/
│   │   ├── auth_server.go       # реалізація AuthService
│   │   └── ticket_server.go     # реалізація TicketService
│   ├── service/
│   │   ├── auth_service.go      # бізнес-логіка: JWT, bcrypt
│   │   └── ticket_service.go    # бізнес-логіка: події, бронювання
│   ├── repo/                    # шар роботи з БД (GORM + SQLite)
│   └── model/                   # моделі даних
└── generate.sh
```

**Запуск сервера** (`cmd/main.go`):

```go
grpcSrv := grpc.NewServer()
pb.RegisterAuthServiceServer(grpcSrv, grpcserver.NewAuthServer(authSvc))
pb.RegisterTicketServiceServer(grpcSrv, grpcserver.NewTicketServer(ticketSvc, authSvc))
reflection.Register(grpcSrv)   // увімкнення Reflection API

log.Printf("gRPC listening on %s", grpcAddr)
grpcSrv.Serve(lis)
```

**JWT-автентифікація** у методах `TicketService` передається через gRPC metadata за ключем `authorization`. Сервер витягує токен і валідує його перед виконанням методу:

```go
func (s *TicketServer) tokenFromMeta(ctx context.Context) (uint, string, error) {
    md, _ := metadata.FromIncomingContext(ctx)
    vals := md.Get("authorization")
    return s.auth.ValidateToken(vals[0])
}
```

---

### 5. Реєстрація Reflection API

Для можливості тестування без локального `.proto`-файлу підключено `grpc/reflection`:

```go
import "google.golang.org/grpc/reflection"
// ...
reflection.Register(grpcSrv)
```

Тепер `grpcurl` може автоматично отримувати схему сервісів:

```
$ grpcurl -plaintext localhost:50051 list
ticketpb.AuthService
ticketpb.TicketService
grpc.reflection.v1alpha.ServerReflection
```

---

### 6. Тестування повного потоку

Для тестування використано утиліту `grpcurl`:

```bash
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
```

**Крок 1 — Реєстрація нового користувача:**

```bash
$ grpcurl -plaintext \
  -d '{"username":"alice","password":"secret","email":"alice@example.com"}' \
  localhost:50051 ticketpb.AuthService/Register

{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "username": "alice"
}
```

**Крок 2 — Логін та отримання токена:**

```bash
$ grpcurl -plaintext \
  -d '{"username":"alice","password":"secret"}' \
  localhost:50051 ticketpb.AuthService/Login

{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "username": "alice"
}
```

**Крок 3 — Перегляд списку подій (без автентифікації):**

```bash
$ grpcurl -plaintext -d '{}' \
  localhost:50051 ticketpb.TicketService/ListEvents

{
  "events": [
    {
      "id": "1",
      "title": "Rock Legends Concert",
      "venue": "Madison Square Garden",
      "date": "2025-07-20 19:00",
      "availableSeats": 200,
      "totalSeats": 200,
      "price": 79.99
    },
    ...
  ]
}
```

**Крок 4 — Бронювання квитків (з JWT-токеном у metadata):**

```bash
$ grpcurl -plaintext \
  -H "authorization: $TOKEN" \
  -d '{"event_id":1,"seats":2}' \
  localhost:50051 ticketpb.TicketService/BookTicket

{
  "id": "1",
  "eventId": "1",
  "eventTitle": "Rock Legends Concert",
  "seats": 2,
  "status": "confirmed",
  "totalPrice": 159.98
}
```

**Крок 5 — Перегляд власних бронювань:**

```bash
$ grpcurl -plaintext \
  -H "authorization: $TOKEN" \
  -d '{}' \
  localhost:50051 ticketpb.TicketService/ListMyBookings

{
  "bookings": [
    {
      "id": "1",
      "eventId": "1",
      "eventTitle": "Rock Legends Concert",
      "seats": 2,
      "status": "confirmed",
      "totalPrice": 159.98
    }
  ]
}
```

**Крок 6 — Скасування бронювання:**

```bash
$ grpcurl -plaintext \
  -H "authorization: $TOKEN" \
  -d '{"id":1}' \
  localhost:50051 ticketpb.TicketService/CancelBooking

{
  "ok": true
}
```

---

## Лістинг програми

### proto/tickets.proto
```protobuf
syntax = "proto3";
package ticketpb;
option go_package = "github.com/necutya/decentrilized_apps/lab1/gen/ticketpb";

message RegisterRequest { string username = 1; string password = 2; string email = 3; }
message LoginRequest    { string username = 1; string password = 2; }
message AuthResponse    { string token = 1; string username = 2; }

service AuthService {
  rpc Register (RegisterRequest) returns (AuthResponse);
  rpc Login    (LoginRequest)    returns (AuthResponse);
}

message Event {
  int64  id              = 1; string title           = 2;
  string venue           = 3; string date            = 4;
  int32  available_seats = 5; int32  total_seats     = 6;
  double price           = 7;
}
message ListEventsRequest  {}
message ListEventsResponse { repeated Event events = 1; }
message GetEventRequest    { int64 id = 1; }
message BookTicketRequest  { int64 event_id = 1; int32 seats = 2; }
message Booking {
  int64  id = 1; int64  event_id = 2; string event_title = 3;
  int32  seats = 4; string status = 5; double total_price = 6;
}
message ListBookingsRequest  {}
message ListBookingsResponse { repeated Booking bookings = 1; }
message CancelBookingRequest  { int64 id = 1; }
message CancelBookingResponse { bool  ok = 1; }

service TicketService {
  rpc ListEvents     (ListEventsRequest)    returns (ListEventsResponse);
  rpc GetEvent       (GetEventRequest)      returns (Event);
  rpc BookTicket     (BookTicketRequest)    returns (Booking);
  rpc ListMyBookings (ListBookingsRequest)  returns (ListBookingsResponse);
  rpc CancelBooking  (CancelBookingRequest) returns (CancelBookingResponse);
}
```

### cmd/main.go
```go
package main

import (
    "log"
    "net"
    "os"
    "time"

    pb "github.com/necutya/decentrilized_apps/lab1/gen/ticketpb"
    "github.com/necutya/decentrilized_apps/lab1/internal/grpcserver"
    "github.com/necutya/decentrilized_apps/lab1/internal/model"
    "github.com/necutya/decentrilized_apps/lab1/internal/repo"
    "github.com/necutya/decentrilized_apps/lab1/internal/service"

    "google.golang.org/grpc"
    "google.golang.org/grpc/reflection"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

func main() {
    db, _ := gorm.Open(sqlite.Open(envOr("DB_PATH", "tickets.db")), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Warn),
    })
    db.AutoMigrate(&model.User{}, &model.Event{}, &model.Booking{})
    seed(db)

    userRepo    := repo.NewUserRepo(db)
    eventRepo   := repo.NewEventRepo(db)
    bookingRepo := repo.NewBookingRepo(db)
    authSvc     := service.NewAuthService(userRepo)
    ticketSvc   := service.NewTicketService(eventRepo, bookingRepo)

    lis, _ := net.Listen("tcp", envOr("GRPC_ADDR", ":50051"))
    grpcSrv := grpc.NewServer()
    pb.RegisterAuthServiceServer(grpcSrv, grpcserver.NewAuthServer(authSvc))
    pb.RegisterTicketServiceServer(grpcSrv, grpcserver.NewTicketServer(ticketSvc, authSvc))
    reflection.Register(grpcSrv)

    log.Printf("gRPC listening on :50051")
    grpcSrv.Serve(lis)
}
```

### internal/grpcserver/auth_server.go
```go
package grpcserver

import (
    "context"
    pb "github.com/necutya/decentrilized_apps/lab1/gen/ticketpb"
    "github.com/necutya/decentrilized_apps/lab1/internal/service"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

type AuthServer struct {
    pb.UnimplementedAuthServiceServer
    svc *service.AuthService
}

func NewAuthServer(svc *service.AuthService) *AuthServer { return &AuthServer{svc: svc} }

func (s *AuthServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.AuthResponse, error) {
    token, err := s.svc.Register(req.Username, req.Password, req.Email)
    if err != nil {
        return nil, status.Error(codes.InvalidArgument, err.Error())
    }
    return &pb.AuthResponse{Token: token, Username: req.Username}, nil
}

func (s *AuthServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.AuthResponse, error) {
    token, err := s.svc.Login(req.Username, req.Password)
    if err != nil {
        return nil, status.Error(codes.Unauthenticated, err.Error())
    }
    return &pb.AuthResponse{Token: token, Username: req.Username}, nil
}
```

### internal/grpcserver/ticket_server.go
```go
package grpcserver

import (
    "context"
    pb "github.com/necutya/decentrilized_apps/lab1/gen/ticketpb"
    "github.com/necutya/decentrilized_apps/lab1/internal/service"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/metadata"
    "google.golang.org/grpc/status"
)

type TicketServer struct {
    pb.UnimplementedTicketServiceServer
    svc  *service.TicketService
    auth *service.AuthService
}

func NewTicketServer(svc *service.TicketService, auth *service.AuthService) *TicketServer {
    return &TicketServer{svc: svc, auth: auth}
}

func (s *TicketServer) ListEvents(ctx context.Context, _ *pb.ListEventsRequest) (*pb.ListEventsResponse, error) {
    events, err := s.svc.ListEvents()
    if err != nil { return nil, status.Error(codes.Internal, err.Error()) }
    return &pb.ListEventsResponse{Events: toProtoEvents(events)}, nil
}

func (s *TicketServer) BookTicket(ctx context.Context, req *pb.BookTicketRequest) (*pb.Booking, error) {
    userID, _, err := s.tokenFromMeta(ctx)
    if err != nil { return nil, status.Error(codes.Unauthenticated, err.Error()) }
    b, err := s.svc.BookTicket(userID, req.EventId, req.Seats)
    if err != nil { return nil, status.Error(codes.FailedPrecondition, err.Error()) }
    return toProtoBooking(*b), nil
}

func (s *TicketServer) CancelBooking(ctx context.Context, req *pb.CancelBookingRequest) (*pb.CancelBookingResponse, error) {
    userID, _, err := s.tokenFromMeta(ctx)
    if err != nil { return nil, status.Error(codes.Unauthenticated, err.Error()) }
    if err := s.svc.CancelBooking(userID, req.Id); err != nil {
        return nil, status.Error(codes.FailedPrecondition, err.Error())
    }
    return &pb.CancelBookingResponse{Ok: true}, nil
}

func (s *TicketServer) tokenFromMeta(ctx context.Context) (uint, string, error) {
    md, _ := metadata.FromIncomingContext(ctx)
    vals := md.Get("authorization")
    if len(vals) == 0 { return 0, "", status.Error(codes.Unauthenticated, "missing authorization") }
    return s.auth.ValidateToken(vals[0])
}
```

---

## Висновки

У ході виконання лабораторної роботи було розроблено клієнт-серверний застосунок з використанням технології gRPC. Описано контракт взаємодії між клієнтом і сервером у форматі Protocol Buffers, що включає два сервіси: `AuthService` (реєстрація та автентифікація) і `TicketService` (перегляд подій, бронювання та скасування квитків).

З `.proto`-специфікації за допомогою `protoc` автоматично згенеровано Go-код — серверні інтерфейси та клієнтські заглушки. Це унеможливлює розсинхронізацію між клієнтом і сервером та суттєво прискорює розробку.

Реалізовано JWT-автентифікацію через механізм gRPC metadata, що є аналогом HTTP-заголовків. Підключення Reflection API дозволило проводити тестування без наявності `.proto`-файлу на стороні клієнта.

Порівняно зі звичайним REST/HTTP підхід gRPC забезпечує: строго типізований контракт API, ефективнішу серіалізацію (бінарний protobuf замість JSON), а також можливість стрімінгу даних у обох напрямках.
