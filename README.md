# entregador

CRUD service with PostgreSQL.

# Requirements

* Golang: [Install Guide](https://golang.org/doc/install)
* PostgreSQL: [Install Guide](https://www.postgresql.org/)
* Set environment variables (see .env file)

# Running the Application Locally

To run the application locally, execute the following command:

```bash
go run main.go
```

# API Endpoints

Here are the available API endpoints:

GET /api/entregador/:id - Get box by ID
POST /api/entregador/ - Create a new box
PUT /api/entregador/:id - Update a box by ID
DELETE /api/entregador/:id - Delete a box by ID

# Deployment

There is a GitHub Actions workflow that deploys the application when you push to the develop or main branches. 

### About Mockery
Mockery is a tool that automatically generates mocks from Go interfaces. This allows you to create consistent and up-to-date mocks without manually implementing each method. In case you need to create new mocks, you can use the following command:

```bash
mockery --name=TwilioPort --dir=internal/domain/ports --output=internal/domain/ports/mocks --outpkg=mocks --filename=twilio.port.mock.go
```