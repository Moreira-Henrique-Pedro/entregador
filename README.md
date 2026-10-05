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
front ──HTTP──▶ DeliveriesController ──▶ RegisterDelivery ──▶ MongoDB
                                                  └──▶ NotificationScheduler ──publish──▶ Pub/Sub topic
                                                                                               │
               NotificationsController ◀── POST /internal/pubsub/notifications ◀──── push ─────┘  (retry + dead-letter)
                 └──▶ NotifyDelivery ──▶ Twilio WhatsApp (pkg/httpclient)
```

Retry and dead-letter are configured on the push subscription, not in code. Details in [docs/mensageria.md](docs/mensageria.md).

## Architecture

The project is split in three layers, the same layout as our other Go services:

- **domain**: the entities and errors (`entities`) and the interfaces of what the application offers and needs (`interfaces/usecases`, `interfaces/repositories`, `interfaces/services`). It knows nothing about HTTP, Pub/Sub, MongoDB or Twilio.
- **application**: the entry points (controllers and their DTOs), the use cases and the commands the API sends to itself.
- **infrastructure**: the implementations of the domain interfaces (MongoDB, Pub/Sub, Twilio), the HTTP server and middlewares, and `providers`, which wires everything together.

Dependencies always point inward: `infrastructure → application → domain`.

```
entregador/
├── cmd/
│   └── api/                        # The only binary: HTTP API + Pub/Sub push endpoint
├── config/                         # Environment variables
├── internal/
│   ├── domain/
│   │   ├── entities/               # Entities, business rules and errors (Resident, Delivery, Notification)
│   │   └── interfaces/
│   │       ├── usecases/           # Use case interfaces the controllers call
│   │       ├── repositories/       # ResidentRepository, DeliveryRepository
│   │       └── services/           # Notifier, NotificationScheduler
│   ├── application/
│   │   ├── controllers/            # One controller per domain: its routes + handler functions (Gin)
│   │   │   └── dtos/
│   │   │       ├── request/        # Bodies received (JSON tags + FromDTO to the entity)
│   │   │       └── response/       # Bodies returned (built from the entities)
│   │   ├── commands/               # NotifyDeliveryCommand: Pub/Sub message contract
│   │   └── usecases/               # Use case implementations (RegisterDelivery, NotifyDelivery, CreateResident, ...)
│   └── infrastructure/
│       ├── providers/              # Composition root: builds the API, wiring controllers → use cases → infrastructure
│       ├── server/                 # Gin engine: mounts every controller + global middlewares
│       ├── middlewares/            # Request logger, Pub/Sub push OIDC auth
│       ├── repositories/mongodb/   # Repositories
│       └── services/
│           ├── notifier/           # Twilio WhatsApp and log notifiers
│           └── pubsub/             # NotificationScheduler: publishes NotifyDeliveryCommand to Pub/Sub
├── pkg/
│   ├── httpclient/                 # HTTP client for external APIs (base URL, timeout, auth, logging)
│   └── logger/
└── docs/                           # API, messaging and Cloud Run deploy guides
```

### Controllers

Each controller owns the routes of one domain and their handler functions, in the same file. It binds the body or query into a `dtos/request` DTO, validated by its `binding` tags (`required`, `oneof`, `required_without`, ...), converts it with `FromDTO()` into the entity the use case receives, and answers with a `dtos/response` DTO. Domain errors become HTTP statuses in one place (`controllers/errors.go`). Business rules and validations live in the entities and are reused by the use cases. `server.New` (Gin) only mounts every controller and adds the request logger and panic recovery; route-specific middlewares (like the Pub/Sub push auth) are applied by the controller itself:

| Controller                | Routes |
|---------------------------|--------|
| `HealthController`        | `GET /health` |
| `ResidentsController`     | `GET/POST /v1/residents`, `PATCH/DELETE /v1/residents/{resident_id}` |
| `DeliveriesController`    | `GET/POST /v1/deliveries`, `DELETE /v1/deliveries/{delivery_id}` |
| `NotificationsController` | `POST /internal/pubsub/notifications` (Pub/Sub push, OIDC token checked) |

```go
func (c *ResidentsController) RegisterRoutes(router gin.IRouter) {
	router.GET("/v1/residents", c.list)
	router.POST("/v1/residents", c.create)
	// ...
}

func (c *ResidentsController) list(ctx *gin.Context) { /* ... */ }
```

### HTTP client

Outbound calls to external APIs go through `pkg/httpclient` (today: Twilio). It sets base URL, timeout, basic auth and default headers once, limits the response size and logs method, path, status and duration. It never logs bodies or query strings, which may carry phones or tokens. A non-2xx status is not an error: each API reports errors in its own format, so the caller checks `Response.IsSuccess()`.

```go
client := httpclient.New(
	httpclient.WithBaseURL("https://api.twilio.com"),
	httpclient.WithBasicAuth(accountSID, authToken),
	httpclient.WithTimeout(10*time.Second),
)
response, err := client.PostForm(ctx, "/2010-04-01/Accounts/"+accountSID+"/Messages.json", form)
```

### Use cases

| Use case                                          | Driven by | Description |
|---------------------------------------------------|-----------|-------------|
| `CreateResident` / `UpdateResident` / `DeleteResident` | HTTP | Resident CRUD, keeping one primary resident and the "Other" resident per apartment. |
| `ListResidentsByApartment` / `ListResidentsByPhone` | HTTP | Resident queries. |
| `RegisterDelivery`                                | HTTP      | Saves the delivery and schedules the arrival notification. |
| `DeleteDelivery`                                  | HTTP      | Marks the delivery as picked up and schedules the pickup notification. Idempotent. |
| `ListDeliveriesByApartment`                       | HTTP      | Delivery queries. |
| `NotifyDelivery`                                  | Pub/Sub push | Resolves the recipients and sends the WhatsApp message. Idempotent. |

### Adding a feature

1. Business rule, entity or error → `domain/entities`.
2. Declare the use case interface in `domain/interfaces/usecases` and implement it in `application/usecases`. If it needs something new from outside, declare an interface in `domain/interfaces/repositories` or `domain/interfaces/services`.
3. Implement that interface in `infrastructure/...` and expose the use case in a controller (`application/controllers`, DTOs in `controllers/dtos`): a new one for a new domain.
4. Wire it in `infrastructure/providers`.

## Running the application

```bash
make up        # MongoDB + Pub/Sub emulator + api (foreground), same topology as Cloud Run
make app-up    # same, in background
make down
```

To run the API outside Docker (for debugging), with the local values from `.env.example`:

```bash
make infra     # MongoDB + Pub/Sub emulator (+ UI), pushing to the api on the host
make api
```

Locally the variables come from `.env.test` (`ENV_FILE`, default in the Makefile and docker-compose), which prints notifications to the log instead of calling Twilio. To send real WhatsApp messages, uncomment its Twilio block. To use another file: `make up ENV_FILE=.env`.

The Pub/Sub emulator UI runs at http://localhost:7200 (set its host to `http://localhost:8085`). See [docs/mensageria.md](docs/mensageria.md#ver-as-mensagens-na-ui).

## Deploy

See [docs/deploy-cloud-run.md](docs/deploy-cloud-run.md): service accounts, secrets, Pub/Sub topic and push subscription with dead-letter, and `gcloud run deploy`.

## API Endpoints

Details, payloads and status codes in [docs/api.md](docs/api.md). How notifications flow through Pub/Sub in [docs/mensageria.md](docs/mensageria.md).

| Method   | Path                                         | Description |
|----------|----------------------------------------------|-------------|
| `POST`   | `/v1/residents`                              | Create a resident |
| `GET`    | `/v1/residents?apartment=<apt>` or `?phone=` | List residents |
| `PATCH`  | `/v1/residents/{resident_id}`                | Partially update a resident |
| `DELETE` | `/v1/residents/{resident_id}`                | Delete a resident |
| `POST`   | `/v1/deliveries`                             | Register a delivery (notifies the arrival) |
| `GET`    | `/v1/deliveries?apartment=<apt>`             | List deliveries of an apartment |
| `DELETE` | `/v1/deliveries/{delivery_id}`               | Pick up a delivery (notifies the pickup) |
| `POST`   | `/v1/users`                                  | Create an admin or doorman user (admin only) |

Every `/v1` route requires a Firebase Auth ID token (`Authorization: Bearer <token>`). Roles: `admin` (síndico) manages residents and users; `doorman` (porteiro) handles deliveries and reads residents. Locally `.env.test` sets `AUTH_ENABLED=false`, so the API accepts every request while the front is being built (the API refuses to start with it in production). With auth on, Firebase Auth runs in an emulator (UI at http://localhost:4000); create the first admin with `make create-admin EMAIL=... NAME=... PASSWORD=...`. See [docs/api.md](docs/api.md#autenticação).

## Tests

Unit tests use [testify](https://github.com/stretchr/testify) and mocks generated by [mockery](https://vektra.github.io/mockery/) from the domain interfaces.

```bash
make test      # all unit tests, with the race detector
make coverage  # total coverage (generated mocks excluded) and coverage.html
make linter    # golangci-lint
```

- Use cases are tested against mocks of the repositories and services (`domain/interfaces/{repositories,services}/mocks`).
- Controllers are tested through their routes (`serveTo`), against mocks of the use cases (`domain/interfaces/usecases/mocks`).

### Mocks

Each interface has a mock in a `mocks/` folder next to it, configured in `.mockery.yaml`. After adding or changing an interface, register it in `.mockery.yaml` (if new) and regenerate:

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
