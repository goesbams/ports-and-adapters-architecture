## The System

The e-wallet system consists of multiple components:

1. **API Service**: Handles HTTP requests for wallet operations
2. **Consumer Service**: Processes events from Kafka for background tasks
3. **Migration Tool**: Manages database schema migrations
4. **Database**: PostgreSQL for storing wallet and transaction data
5. **Cache**: Redis for improved performance
6. **Message Broker**: Kafka for event-driven architecture

These components work together to provide a scalable, event-driven system with clean architecture.# Mini E-Wallet Application

This is a Mini E-Wallet application built using Golang with Hexagonal (Ports and Adapters) Architecture. The application provides basic wallet functionality including deposits, withdrawals, and transfers.

## Architecture

This project follows the Hexagonal Architecture (Ports and Adapters pattern) to ensure:

- Clean separation of concerns
- Domain logic isolated from external systems
- Testability and maintainability
- Flexibility to change infrastructure components

Key components:

- **Domain Layer:** Core business models and rules (`internal/domain/`)
- **Application Layer:** Use cases orchestrating business logic (`internal/usecase/`)
- **Ports Layer:** Interfaces defining how to interact with external systems (`internal/ports/`)
- **Adapters Layer:** Implementations of ports for specific technologies (`internal/adapters/`)
- **API Layer:** HTTP handlers using Echo framework (`api/rest/handlers/`)
- **Event Processing:** Asynchronous event consumers for background tasks

## Features

- Create and manage wallets
- Deposit funds to wallet
- Withdraw funds from wallet
- Transfer funds between wallets
- Process payments via payment gateways
- View transaction history
- Asynchronous event processing for background tasks
- Event-driven architecture for scalability

## Tech Stack

- **Go 1.24:** Programming language
- **Echo:** HTTP web framework
- **PostgreSQL:** Primary database for wallet and transaction data
- **Redis:** Caching layer
- **Kafka:** Event streaming for async processing
- **Elasticsearch:** (Optional) For logging and analytics
- **Docker & Docker Compose:** Containerization

## Prerequisites

- Docker and Docker Compose
- Go 1.24 or higher (for local development)
- Make (optional, for using Makefile commands)

## Getting Started

### Running with Docker

1. Clone the repository:
   ```
   git clone https://github.com/yourusername/mini-ewallet.git
   cd mini-ewallet
   ```

2. Start the application using Docker Compose:
   ```
   make docker-start
   ```
   or
   ```
   docker-compose up
   ```

3. Run database migrations:
   ```
   make docker-migrate
   ```

The API will be available at http://localhost:8080

### Local Development

1. Install dependencies:
   ```
   make install-deps
   ```

2. Set up local PostgreSQL, Redis, and Kafka instances or modify `config/config.local.yaml` to point to your existing instances.

3. Run database migrations:
   ```
   make migrate-up
   ```

4. Start the application:
   ```
   make run
   ```

## API Endpoints

### Wallet Endpoints

- `POST /api/v1/wallets` - Create a new wallet
- `GET /api/v1/wallets/:id` - Get wallet details
- `GET /api/v1/users/:user_id/wallets` - Get all wallets for a user
- `POST /api/v1/wallets/:id/deposit` - Deposit funds to wallet
- `POST /api/v1/wallets/:id/withdraw` - Withdraw funds from wallet
- `POST /api/v1/wallets/:id/transfer` - Transfer funds to another wallet
- `GET /api/v1/wallets/:id/transactions` - Get transaction history for a wallet

### Payment Endpoints

- `POST /api/v1/payments/process` - Process a payment
- `POST /api/v1/payments/verify` - Verify a payment status

## Configuration

Configuration files are located in the `config` directory:

- `config.yaml` - Base configuration
- `config.local.yaml` - Local development overrides
- `config.dev.yaml` - Development environment configuration
- `config.staging.yaml` - Staging environment configuration
- `config.prod.yaml` - Production environment configuration

The application uses the Viper library to manage configuration, allowing for:
- Configuration from files
- Environment variable overrides
- Nested configuration values

## Helper Commands

Use the Makefile for common tasks:

```
make help               # Show available commands
make build              # Build the application
make run                # Run locally
make test               # Run tests
make docker-build       # Build Docker images
make docker-start       # Start services with Docker Compose
make docker-down        # Stop all services
make migrate-up         # Apply database migrations
make migrate-down       # Rollback database migrations
make migrate-create     # Create a new migration file
```

## Implementation Notes

1. **Security Considerations**
   - Production deployments should use environment variables or secrets management for sensitive configuration
   - Authentication and authorization not implemented in this demo
   - Input validation is implemented using validator package

2. **Testing**
   - Unit tests are available in the `tests` directory
   - In-memory repository implementations are used for testing

3. **Extending the Application**
   - To add new features, define domain models and use cases first
   - Create appropriate ports (interfaces) for external dependencies
   - Implement adapters for specific technologies
   - Create REST handlers that use the use cases

## License

This project is licensed under the MIT License - see the LICENSE file for details.