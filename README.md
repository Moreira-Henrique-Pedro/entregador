# Entregador API

CRUD service with PostgreSQL.

# Requirements

* Golang 1.23 or higher: [Install Guide](https://golang.org/doc/install)
* Docker and Docker Compose: [Install Guide](https://docs.docker.com/compose/install/)
* PostgreSQL: [Install Guide](https://www.postgresql.org/)
* MongoDB: [Install Guide](https://www.mongodb.com/pt-br/docs/manual/installation/)
* Set environment variables (see .env file)

## Architecture

The project follows the principles of Hexagonal Architecture, which promotes separation of concerns and independence of frameworks. 
The main components of the architecture are organized as shown below:
```
entregador/
├── .github/                           # GitHub actions for CI/CD
├── cmd/                               # Entry point of the application
│    └── entregador/
│       └── main.go                 
├── internal/                          # Internal packages
│   ├── adapters/
│   │   ├── entrypoints/               # Entry points for HTTP API
│   │   └── ... 
│   ├── domain/                        # Domain layer
│   │ ├── entities/                    # Business entities
│   │ ├── ports/                       # Interfaces for external services, such as APIs, databases, and other services.
│   │ └── usecases/                    # Application logic
├── scripts                            # Scripts for development or test environment configurations.
├── go.mod                             # Go module definition 
├── go.sum                             # Go module dependencies
├── Makefile                           # Makefile  
└── README.md                          # Project documentation
```

# Running the Application

1. To run only the containers:
```bash
make up 
```

2. To run the entire application:
```bash
make app-up
```

# API Endpoints

Here are the available API endpoints:

POST /v1/entregador/ - Create a new delivery

## Tests

This project uses the `stretchr/testify` testing library and the `Mockery` mock generation tool to ensure code quality and reliability.

### Unit Tests

To run the unit tests from all folders starting from the root, you can use the command:

```bash
make test
```

# Deployment

There is a GitHub Actions workflow that deploys the application when you push to the develop or main branches. 

# About Mockery
Mockery is a tool that automatically generates mocks from Go interfaces. This allows you to create consistent and up-to-date mocks without manually implementing each method. In case you need to create new mocks, you can use the following command:

```bash
mockery --name=DeliveryRepositoryPort --dir=internal/domain/ports --output=internal/domain/ports/mocks --outpkg=mocks --filename=postgres.port.mock.go
```