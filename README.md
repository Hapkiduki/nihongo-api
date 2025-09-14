# Nihongo API

A production-ready REST API for Japanese language learning built with Go, Fiber, and Clean Architecture. This backend serves as the data source for a Flutter mobile application focused on learning Hiragana, Katakana, and Kanji through drawing exercises.

## 🚀 Features

- **Syllable Management**: Complete Hiragana and Katakana character database with SVG stroke data
- **Kanji Learning**: JLPT-level organized Kanji with meanings, readings, and drawing paths
- **Course System**: Structured learning courses with lessons and exercises
- **Progress Tracking**: User progress monitoring across all learning entities
- **JWT Authentication**: Secure user authentication and authorization
- **Premium Content**: RevenueCat integration for subscription-based premium courses
- **Clean Architecture**: Modular, testable, and maintainable codebase
- **Docker Support**: Containerized deployment with multi-stage builds
- **Structured Logging**: Zerolog for consistent, structured logging
- **Input Validation**: Comprehensive validation using go-playground/validator
- **Error Handling**: Consistent error wrapping and handling across all layers
- **Configuration Management**: Viper-based configuration with environment variable support

## 🏗️ Architecture

This project follows Clean Architecture principles with SOLID design patterns, Repository Pattern, and Modular Monolith approach. The structure ensures separation of concerns, testability, and maintainability.

```
├── cmd/server/          # Application entry point
├── internal/
│   ├── domain/          # Business entities with validation (User, Kanji, Course, etc.)
│   ├── application/     # Use cases and business logic
│   │   └── service/     # Application services with dependency injection
│   ├── ports/           # Interfaces (Repository contracts - Dependency Inversion)
│   └── adapters/        # External concerns implementations
│       ├── http/        # HTTP handlers and routing with JWT middleware
│       └── storage/     # Database implementations (MongoDB)
├── pkg/                 # Shared packages
│   ├── config/          # Configuration management with Viper
│   └── database/        # Database connection utilities
└── config.yml           # Default configuration file
```

### Design Patterns Implemented

- **Clean Architecture**: Clear separation between domain, application, and infrastructure layers
- **Repository Pattern**: Abstract data access through interfaces
- **Dependency Injection**: Services receive dependencies through interfaces
- **SOLID Principles**:
  - **S**ingle Responsibility: Each component has one reason to change
  - **O**pen/Closed: Open for extension, closed for modification
  - **L**iskov Substitution: Implementations can be substituted for interfaces
  - **I**nterface Segregation: Specific interfaces for different clients
  - **D**ependency Inversion: Depend on abstractions, not concretions
- **Modular Monolith**: Organized modules within a single deployable unit

## 📋 Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose
- MongoDB (handled by Docker)
- Redis (handled by Docker)

## 🛠️ Installation

1. **Clone the repository:**

   ```bash
   git clone https://github.com/hapkiduki/nihongo-api.git
   cd nihongo-api
   ```

2. **Install dependencies:**

   ```bash
   go mod tidy
   ```

3. **Set up environment variables:**

   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. **Start the development environment:**

   ```bash
   docker-compose up -d
   ```

5. **Run the application:**
   ```bash
   go run cmd/server/main.go
   ```

The API will be available at `http://localhost:3000`

## 📖 API Documentation

### Authentication Endpoints

#### Register User

```http
POST /api/auth/register
Content-Type: application/json

{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "securepassword"
}
```

#### Login

```http
POST /api/auth/login
Content-Type: application/json

{
  "email": "john@example.com",
  "password": "securepassword"
}
```

### Syllable Endpoints

#### Get All Syllables

```http
GET /api/syllables
Authorization: Bearer <jwt_token>
```

#### Get Syllable by ID

```http
GET /api/syllables/{id}
Authorization: Bearer <jwt_token>
```

### Protected Endpoints

#### Get User Profile

```http
GET /api/protected/profile
Authorization: Bearer <jwt_token>
```

#### Get Courses

```http
GET /api/protected/courses
Authorization: Bearer <jwt_token>
```

## ⚙️ Configuration

The application uses Viper for configuration management. Key configuration options:

| Variable             | Description               | Default                         |
| -------------------- | ------------------------- | ------------------------------- |
| `SERVER_PORT`        | HTTP server port          | `3000`                          |
| `MONGO_URI`          | MongoDB connection string | `mongodb://mongo:27017/nihongo` |
| `REDIS_ADDR`         | Redis server address      | `redis:6379`                    |
| `JWT_SECRET`         | JWT signing secret        | Required                        |
| `REVENUECAT_API_KEY` | RevenueCat API key        | Required                        |

## Nueva Configuración con Viper

La aplicación ahora usa un sistema de configuración híbrido con Viper para mayor seguridad y flexibilidad.

### Flujo de Configuración

Viper prioriza las fuentes en este orden:

1. Variables de entorno (e.g., `APP_SERVER_PORT=3000`)
2. Archivo .env (gitignore, para secrets locales)
3. config.yml (defaults no sensibles en repo)

### Setup

1. Copia `.env.example` a `.env.dev` para dev o `.env.prod` para prod, y edita secrets (e.g., `APP_AUTH_JWT_SECRET=your-secret`).
2. Para dev: `docker-compose -f docker-compose.dev.yml up` (carga .env.dev).
3. Para prod: `docker-compose -f docker-compose.prod.yml up` (carga .env.prod).
4. Validación: La struct Config valida required fields y formatos (e.g., URL para base_url).

### Variables Clave

| Variable                  | Descripción                    | Ejemplo Default                   |
| ------------------------- | ------------------------------ | --------------------------------- |
| `APP_SERVER_PORT`         | Puerto del servidor HTTP       | 3000                              |
| `APP_DATABASE_MONGO_URI`  | URI de MongoDB                 | mongodb://localhost:27017/nihongo |
| `APP_DATABASE_REDIS_ADDR` | Dirección de Redis             | localhost:6379                    |
| `APP_AUTH_JWT_SECRET`     | Secret para JWT (min 32 chars) | your-secret-here                  |
| `APP_REVENUECAT_API_KEY`  | API key de RevenueCat          | your-key-here                     |
| `APP_REVENUECAT_BASE_URL` | Base URL de RevenueCat         | https://api.revenuecat.com/v1     |

Nunca commitees .env o secrets. Para Docker, env vars se inyectan via env_file.

## Despliegue a Fly.io

Fly.io es una plataforma para desplegar apps Go/Docker fácilmente.

### Pasos

1. Instala flyctl: `brew install flyctl` (macOS) o descarga desde fly.io/docs/hq/install.
2. `fly auth login`.
3. `fly launch`: Genera fly.toml basado en Dockerfile.
4. Edita fly.toml: Agrega [env] para defaults no secrets (e.g., APP_SERVER_PORT = "3000").
5. Set secrets: `fly secrets set APP_AUTH_JWT_SECRET=your-secret APP_REVENUECAT_API_KEY=your-key`.
6. Para DB: `fly postgres create` para Mongo-like (o usa external), set URI como secret.
7. `fly deploy`: Build y deploy desde Dockerfile.
8. Accede: `fly open`, monitorea `fly logs`.
9. Escala: `fly scale count 2`.

Viper leerá env vars de Fly. Compatible con el nuevo sistema de config.

## 🧪 Testing

```bash
go test ./...
```

## 📦 Deployment

### Using Docker

```bash
# Build and run with Docker Compose
docker-compose up --build
```

### Production Deployment

1. Build the Docker image:

   ```bash
   docker build -t nihongo-api .
   ```

2. Run the container:
   ```bash
   docker run -p 3000:3000 nihongo-api
   ```

## 🔧 Development

### Project Structure Details

- **Domain Layer**: Contains business entities and rules
- **Application Layer**: Contains use cases and application services
- **Ports Layer**: Defines interfaces for external dependencies
- **Adapters Layer**: Implements external concerns (HTTP, database)

### Go Best Practices Implemented

Following the latest Go team's recommendations (as of September 2025):

- **Error Handling**: Consistent error wrapping with `fmt.Errorf` and error chains
- **Context Usage**: Proper context propagation throughout the application
- **Structured Logging**: Zerolog for performance and structured output
- **Configuration**: Viper for flexible configuration management
- **Validation**: Input validation using struct tags and validator library
- **Dependency Injection**: Clean dependency management through interfaces
- **Graceful Shutdown**: Proper cleanup of resources (database, Redis connections)
- **Modular Design**: Clear separation of concerns with internal package organization

### Adding New Features

1. Define domain entities in `internal/domain/` with validation tags
2. Create repository interfaces in `internal/ports/`
3. Implement business logic in `internal/application/service/`
4. Create HTTP handlers in `internal/adapters/http/`
5. Update routing in `internal/adapters/http/router/`
6. Add structured logging where appropriate
7. Ensure proper error handling and context usage

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html) by Robert C. Martin
- [Fiber](https://gofiber.io/) web framework
- [MongoDB](https://www.mongodb.com/) for data persistence
- [RevenueCat](https://www.revenuecat.com/) for subscription management

---

Made with ❤️ for Japanese language learners worldwide

## RevenueCat Webhook: secret rotation & best practices

This project integrates RevenueCat webhooks for subscription events. Follow these guidelines to keep webhook handling secure and reliable:

- Secret rotation

  - Store webhook secrets outside the repository (use environment variables or a secrets manager).
  - The application supports multiple comma-separated webhook secrets (`APP_REVENUECAT_WEBHOOK_SECRET` or `revenuecat.webhook_secrets` in `config.yml`) to allow rolling rotations: add the new secret alongside the old one, deploy, then remove the old secret after verification.

- Verification

  - Webhook signature verification uses HMAC-SHA256 on the raw request body and constant-time comparison.
  - The code verifies against all configured secrets to support rotation.

- Error handling and retries

  - Return 401 for missing/invalid signatures.
  - Return 5xx for transient server errors (database/network) so RevenueCat can retry.
  - Treat duplicate-event DB errors as idempotent and respond 200 to prevent duplicate processing.

- Operational best practices

  - Limit request body size and validate Content-Type to avoid resource exhaustion.
  - Redact PII (emails, names) from logs. Log event IDs and non-sensitive metadata for correlation.
  - Add metrics for signature failures and processing errors to detect attacks or misconfiguration.
  - Consider enqueueing heavy processing into a background worker for low-latency responses.
  - Use HTTPS and a WAF when possible; consider IP allowlisting if RevenueCat publishes IP ranges.

- Local testing
  - Use the test utilities and Unit tests included in `internal/adapters/http/webhook`.
  - For local development, set `APP_REVENUECAT_WEBHOOK_SECRET` in `.env.dev` and use a tool (curl/postman) to send signed requests.

Para detalles más extensos, pasos con ngrok y configuración por ambientes (dev/sandbox vs prod), consulta la guía completa en `docs/revenuecat-webhook.md`.

Following these patterns keeps the webhook integration secure, testable, and operationally robust.
