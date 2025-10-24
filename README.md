# Todo List App

A full-stack todo list application built with Go backend and React frontend.

## Tech Stack

### Backend
- Go with SQLite database
- OpenAPI spec generated from Go code using swag
- Unit testing with Go's built-in testing package
- Linting and formatting with golangci-lint and gofmt

### Frontend
- React + Vite + TypeScript
- TypeScript API client generated with openapi-generator-cli
- Linting with Oxlint
- Formatting with Prettier
- Unit testing with Vitest and @testing-library/react

## Project Structure

```
.
├── cmd/
│   └── server/          # Main server entry point
├── internal/
│   ├── database/        # Database layer and repository
│   ├── handlers/        # HTTP handlers
│   └── models/          # Data models
├── docs/                # Generated OpenAPI documentation
├── frontend/            # React frontend application
│   └── src/
│       ├── generated/   # Generated API client
│       └── test/        # Test setup
├── Makefile            # Build and development commands
└── README.md
```

## Getting Started

### Prerequisites

- Go 1.25 or higher
- Node.js 18 or higher
- Make (optional, but recommended)

### Backend Setup

1. Install Go dependencies:
   ```bash
   go mod download
   ```

2. Install development tools:
   ```bash
   make install
   # Or manually:
   go install github.com/swaggo/swag/cmd/swag@latest
   go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
   ```

3. Generate OpenAPI documentation:
   ```bash
   make docs
   # Or manually:
   swag init -g cmd/server/main.go -o ./docs
   ```

4. Run the backend server:
   ```bash
   make run
   # Or manually:
   go run ./cmd/server/main.go
   ```

   The server will start on http://localhost:8080

### Frontend Setup

1. Navigate to the frontend directory:
   ```bash
   cd frontend
   ```

2. Install dependencies:
   ```bash
   npm install
   ```

3. Generate the TypeScript API client:
   ```bash
   npm run generate:api
   ```

4. Start the development server:
   ```bash
   npm run dev
   ```

   The frontend will start on http://localhost:5173

## Development Commands

### Backend

```bash
make build          # Build the server binary
make run            # Run the server
make test           # Run tests
make test-coverage  # Run tests with coverage report
make lint           # Run linter
make fmt            # Format code
make docs           # Generate OpenAPI documentation
make clean          # Clean build artifacts
make dev            # Generate docs and run server
```

### Frontend

```bash
npm run dev              # Start development server
npm run build            # Build for production
npm test                 # Run tests
npm run test:ui          # Run tests with UI
npm run test:coverage    # Run tests with coverage
npm run lint             # Run Oxlint
npm run format           # Format code with Prettier
npm run format:check     # Check code formatting
npm run generate:api     # Generate TypeScript API client
```

## API Endpoints

- `GET /api/todos` - Get all todos
- `GET /api/todos/{id}` - Get a single todo
- `POST /api/todos` - Create a new todo
- `PATCH /api/todos/{id}` - Update a todo
- `DELETE /api/todos/{id}` - Delete a todo
- `GET /health` - Health check endpoint

## Testing

### Backend Tests

The backend includes unit tests for the HTTP handlers. Tests use an in-memory SQLite database.

Run tests:
```bash
make test
```

View coverage:
```bash
make test-coverage
```

### Frontend Tests

The frontend uses Vitest and React Testing Library for unit tests.

Run tests:
```bash
cd frontend
npm test
```

Run tests with UI:
```bash
npm run test:ui
```

## Code Quality

### Backend

- **Linting**: golangci-lint with multiple linters enabled
- **Formatting**: gofmt with simplification enabled
- **Configuration**: `.golangci.yml`

Run linter:
```bash
make lint
```

Format code:
```bash
make fmt
```

### Frontend

- **Linting**: Oxlint for fast, accurate linting
- **Formatting**: Prettier with consistent configuration
- **Configuration**: `.prettierrc`

Run linter:
```bash
cd frontend
npm run lint
```

Format code:
```bash
npm run format
```

## Environment Variables

### Backend

- `DB_PATH` - Path to SQLite database file (default: `./todos.db`)
- `PORT` - Server port (default: `8080`)

### Frontend

The frontend connects to the backend at `http://localhost:8080` by default. This can be configured in the generated API client.

## Production Build

### Backend

```bash
make build
./bin/server
```

### Frontend

```bash
cd frontend
npm run build
```

The built files will be in `frontend/dist/`

## License

MIT
