# Entregador

Delivery management for residential buildings. The front desk registers packages through an HTTP API, and residents are notified on WhatsApp when a package arrives and when it is picked up.

- Residents and deliveries are saved **synchronously** (MongoDB), so the client gets the real result right away.
- Notifications are sent **asynchronously**: the API publishes a command to Kafka and a worker sends it through Twilio. A slow or unavailable WhatsApp never slows down or breaks the API.

## Requirements

* Go 1.25 or higher: [Install Guide](https://golang.org/doc/install)
* Docker and Docker Compose: [Install Guide](https://docs.docker.com/compose/install/) (runs Kafka and MongoDB locally)
* Environment variables: copy `.env.example` to `.env`

## Applications

There are two binaries, deployed and scaled independently. They share the same core and differ only in the adapter that drives it:

| Binary       | Entry point | What it does |
|--------------|-------------|--------------|
| `cmd/api`    | HTTP        | Residents CRUD, delivery registration/pickup and queries. Schedules notifications on Kafka. |
| `cmd/worker` | Kafka       | Consumes `delivery-internal.commands` and sends the WhatsApp notifications, with retry and DLQ. |

```
            ┌──────────────────────── core ────────────────────────┐
 HTTP ──▶ adapters/in/http ──▶ ports/in ──▶ usecases ──▶ ports/out ──▶ adapters/out/mongodb
                               ▲                         │          ──▶ adapters/out/kafka    (schedules)
 Kafka ─▶ adapters/in/kafka ───┘                         │          ──▶ adapters/out/notifier (Twilio / log)
            └────────────────────────────────────────────┘
```

## Architecture

The project follows **Hexagonal Architecture (Ports & Adapters)**:

- The **domain** and the **use cases** are the core and know nothing about HTTP, Kafka, MongoDB or Twilio.
- The core declares what it offers (**driving ports**, `ports/in`) and what it needs (**driven ports**, `ports/out`).
- The **adapters** implement those ports on the edges.
- `bootstrap` is the composition root that wires everything together.

Dependencies always point inward: `adapters → application → domain`.

```
entregador/
├── cmd/
│   ├── api/                    # HTTP API binary
│   └── worker/                 # Kafka worker binary (notifications)
├── config/                     # Env vars and the worker consumer config (subscriber/deployments/*.json)
├── internal/
│   ├── domain/                 # Entities, business rules and domain errors (Resident, Delivery, NotificationType)
│   ├── application/
│   │   ├── ports/
│   │   │   ├── in/             # Driving ports: the use case interfaces the adapters call
│   │   │   └── out/            # Driven ports: repositories, Notifier, NotificationScheduler
│   │   └── usecases/           # Use case implementations (RegisterDelivery, NotifyDelivery, CreateResident, ...)
│   ├── adapters/
│   │   ├── in/
│   │   │   ├── http/           # REST handlers and router
│   │   │   └── kafka/          # Consumer (retry + DLQ) and message → use case routing
│   │   ├── out/
│   │   │   ├── mongodb/        # Repositories
│   │   │   ├── kafka/          # NotificationScheduler: publishes NotifyDelivery
│   │   │   └── notifier/       # Twilio WhatsApp and log notifiers
│   │   └── messages/           # Kafka message contracts shared by the in/out Kafka adapters
│   └── bootstrap/              # Composition root: builds the API and the worker
├── pkg/                        # Generic, domain-agnostic libraries (logger, pubsub, watermill, events, duration)
└── docs/                       # API and messaging guides
```

### Use cases

| Use case                                          | Driven by | Description |
|---------------------------------------------------|-----------|-------------|
| `CreateResident` / `UpdateResident` / `DeleteResident` | HTTP | Resident CRUD, keeping one primary resident and the "Other" resident per apartment. |
| `ListResidentsByApartment` / `ListResidentsByPhone` | HTTP | Resident queries. |
| `RegisterDelivery`                                | HTTP      | Saves the delivery and schedules the arrival notification. |
| `DeleteDelivery`                                  | HTTP      | Marks the delivery as picked up and schedules the pickup notification. Idempotent. |
| `ListDeliveriesByApartment`                       | HTTP      | Delivery queries. |
| `NotifyDelivery`                                  | Kafka     | Resolves the recipients and sends the WhatsApp message. Idempotent. |

### Adding a feature

1. Business rule or entity → `internal/domain`.
2. Declare the use case interface in `ports/in` and implement it in `usecases`. If it needs something new from outside, declare an interface in `ports/out`.
3. Implement the new driven port in `adapters/out/...` and expose the use case in `adapters/in/...`.
4. Wire it in `internal/bootstrap`.

## Running the application

```bash
make up        # Kafka, MongoDB, Kafka UI, api and worker (foreground)
make app-up    # same, in background
make down
```

To run a binary outside Docker (with `.env` pointing to `localhost`):

```bash
make api
make worker
```

Use `NOTIFIER_PROVIDER=log` in `.env` to print notifications to the worker log instead of calling Twilio.

## API Endpoints

Details, payloads and status codes in [docs/api.md](docs/api.md). How notifications flow through Kafka in [docs/eventos.md](docs/eventos.md).

| Method   | Path                                         | Description |
|----------|----------------------------------------------|-------------|
| `POST`   | `/v1/residents`                              | Create a resident |
| `GET`    | `/v1/residents?apartment=<apt>` or `?phone=` | List residents |
| `PATCH`  | `/v1/residents/{resident_id}`                | Partially update a resident |
| `DELETE` | `/v1/residents/{resident_id}`                | Delete a resident |
| `POST`   | `/v1/deliveries`                             | Register a delivery (notifies the arrival) |
| `GET`    | `/v1/deliveries?apartment=<apt>`             | List deliveries of an apartment |
| `DELETE` | `/v1/deliveries/{delivery_id}`               | Pick up a delivery (notifies the pickup) |

## Tests

Unit tests use [testify](https://github.com/stretchr/testify) and mocks generated by [mockery](https://vektra.github.io/mockery/) from the ports.

```bash
make test      # all unit tests, with the race detector
make coverage  # total coverage (generated mocks excluded) and coverage.html
make linter    # golangci-lint
```

- Use cases are tested against mocks of the driven ports (`ports/out/mocks`).
- Inbound adapters are tested against mocks of the driving ports (`ports/in/mocks`).

### Mocks

Each port has a mock in a `mocks/` folder next to it, configured in `.mockery.yaml`. After adding or changing a port, register it in `.mockery.yaml` (if new) and regenerate:

```bash
make mocks
```

In tests, create the mock with its constructor, which fails the test on unmet expectations:

```go
residents := outmocks.NewResidentRepository(t)
residents.EXPECT().FindByResidentID(mock.Anything, "r1").Return(resident, nil).Once()
```

## CI

The GitHub Actions workflow in `.github/workflows/go.yml` runs lint, unit tests and build on pushes and pull requests to `main` and `develop`.
