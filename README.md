# Entregador

Delivery management for residential buildings. The front desk registers packages through an HTTP API, and residents are notified on WhatsApp when a package arrives and when it is picked up.

- Residents and deliveries are saved **synchronously** (MongoDB), so the client gets the real result right away.
- Notifications are sent **asynchronously**: the API publishes a command to a queue and the WhatsApp message is sent outside the request, through Twilio. A slow or unavailable WhatsApp never slows down or breaks the API.

## Requirements

* Go 1.25 or higher: [Install Guide](https://golang.org/doc/install)
* Docker and Docker Compose: [Install Guide](https://docs.docker.com/compose/install/) (runs MongoDB and the Pub/Sub emulator locally)
* Environment variables: copy `.env.example` to `.env`

## Deployment topology

Production runs on **Google Cloud Run** as **a single service** + **Pub/Sub** + **MongoDB Atlas** (step by step in [docs/deploy-cloud-run.md](docs/deploy-cloud-run.md)). Pub/Sub *pushes* each notification back to the API over HTTP, so there is no long-running consumer and the service can scale to zero.

```
front ──HTTP──▶ cmd/api ──▶ RegisterDelivery ──▶ MongoDB
                  ▲    └──▶ NotificationScheduler ──publish──▶ Pub/Sub topic
                  │                                                │
                  └──── POST /internal/pubsub/notifications ◀──push┘  (retry + dead-letter)
                         └──▶ NotifyDelivery ──▶ Twilio WhatsApp
```

The messaging adapter is chosen by `MESSAGING_PROVIDER` (details in [docs/mensageria.md](docs/mensageria.md)):

| `MESSAGING_PROVIDER` | Binaries | Notes |
|----------------------|----------|-------|
| `pubsub` (default)   | `cmd/api` only | Cloud Run. Retry and dead-letter are configured on the push subscription. |
| `kafka`              | `cmd/api` + `cmd/worker` | The worker consumes Kafka; retry and DLQ are done in code. |

## Architecture

The project follows **Hexagonal Architecture (Ports & Adapters)**:

- The **domain** and the **use cases** are the core and know nothing about HTTP, Pub/Sub, Kafka, MongoDB or Twilio.
- The core declares what it offers (**driving ports**, `ports/in`) and what it needs (**driven ports**, `ports/out`).
- The **adapters** implement those ports on the edges.
- `providers` is the composition root that wires everything together.

Dependencies always point inward: `adapters → application → domain`.

```
entregador/
├── cmd/
│   ├── api/                    # HTTP API binary (+ Pub/Sub push endpoint)
│   └── worker/                 # Kafka worker binary (only with MESSAGING_PROVIDER=kafka)
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
│   │   │   ├── pubsub/         # Pub/Sub push handler (OIDC token check) → NotifyDelivery
│   │   │   └── kafka/          # Kafka consumer (retry + DLQ) → NotifyDelivery
│   │   ├── out/
│   │   │   ├── mongodb/        # Repositories
│   │   │   ├── pubsub/         # NotificationScheduler: publishes NotifyDelivery to Pub/Sub
│   │   │   ├── kafka/          # NotificationScheduler: publishes NotifyDelivery to Kafka
│   │   │   └── notifier/       # Twilio WhatsApp and log notifiers
│   │   └── messages/           # NotifyDelivery message contract, shared by the messaging adapters
│   └── providers/              # Composition root: builds the API and the worker, picks the messaging adapter
├── pkg/                        # Generic, domain-agnostic libraries (logger, pubsub, watermill, events, duration)
└── docs/                       # API, messaging and Cloud Run deploy guides
```

### Use cases

| Use case                                          | Driven by | Description |
|---------------------------------------------------|-----------|-------------|
| `CreateResident` / `UpdateResident` / `DeleteResident` | HTTP | Resident CRUD, keeping one primary resident and the "Other" resident per apartment. |
| `ListResidentsByApartment` / `ListResidentsByPhone` | HTTP | Resident queries. |
| `RegisterDelivery`                                | HTTP      | Saves the delivery and schedules the arrival notification. |
| `DeleteDelivery`                                  | HTTP      | Marks the delivery as picked up and schedules the pickup notification. Idempotent. |
| `ListDeliveriesByApartment`                       | HTTP      | Delivery queries. |
| `NotifyDelivery`                                  | Pub/Sub push or Kafka | Resolves the recipients and sends the WhatsApp message. Idempotent. |

### Adding a feature

1. Business rule or entity → `internal/domain`.
2. Declare the use case interface in `ports/in` and implement it in `usecases`. If it needs something new from outside, declare an interface in `ports/out`.
3. Implement the new driven port in `adapters/out/...` and expose the use case in `adapters/in/...`.
4. Wire it in `internal/providers`.

## Running the application

```bash
make up        # MongoDB + Pub/Sub emulator + api (foreground), same topology as Cloud Run
make app-up    # same, in background
make up-kafka  # Kafka mode: Kafka + Kafka UI + api + worker (set MESSAGING_PROVIDER=kafka in .env)
make down
```

To run the API outside Docker (for debugging), with the local values from `.env.example`:

```bash
make infra     # MongoDB + Pub/Sub emulator, pushing to the api on the host
make api
```

Use `NOTIFIER_PROVIDER=log` in `.env` to print notifications to the log instead of calling Twilio.

## Deploy

See [docs/deploy-cloud-run.md](docs/deploy-cloud-run.md): service accounts, secrets, Pub/Sub topic and push subscription with dead-letter, and `gcloud run deploy`.

## API Endpoints

Details, payloads and status codes in [docs/api.md](docs/api.md). How notifications flow through Pub/Sub (or Kafka) in [docs/mensageria.md](docs/mensageria.md).

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
