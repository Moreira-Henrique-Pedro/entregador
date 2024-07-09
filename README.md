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
