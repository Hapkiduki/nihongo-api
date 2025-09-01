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

## 🏗️ Architecture

This project follows Clean Architecture principles with the following structure:

```
├── cmd/server/          # Application entry point
├── internal/
│   ├── domain/          # Business entities (User, Kanji, Course, etc.)
│   ├── application/     # Use cases and business logic
│   │   └── service/     # Application services
│   ├── ports/           # Interfaces (Repository contracts)
│   └── adapters/        # External concerns implementations
│       ├── http/        # HTTP handlers and routing
│       └── storage/     # Database implementations
├── pkg/                 # Shared packages (database, logger)
└── config/              # Configuration files
```

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

### Adding New Features

1. Define domain entities in `internal/domain/`
2. Create repository interfaces in `internal/ports/`
3. Implement business logic in `internal/application/service/`
4. Create HTTP handlers in `internal/adapters/http/`
5. Update routing in `internal/adapters/http/router/`

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
