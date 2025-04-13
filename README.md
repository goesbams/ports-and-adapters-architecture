# Port And Adapter Architecture
## Overview
Port and adapter architecture or well-known as **Hexagonal Architecture** is firstly introduced by **Alistair Cockburn** in 2005. The main idea is  decoupling an application's core business logic from external systems (e.g., databases, UIs, APIs) by organizing code into ports (interfaces) and adapters (implementations). In Go, this is achieved using interfaces, dependency injection, and a clean separation of concerns.

![Hexagonal Architecture Concept](documentations/hexagonal_architecture_concept.png)

## Core Principles in Go
1. **Decouple Logic from Infrastructure**
    - The core domain never directly depends on frameworks, databases, or UI.
    - External dependencies (e.g., PostgreSQL, HTTP servers) implement interfaces defined by the core.

2. **Ports (Interfaces)**
    - **Inbound Ports:** Define how external actors (e.g., users, APIs) interact with the core.
    - **Outbound Ports:** Define how the core interacts with external services (e.g., databases).

3. **Adapters (Implementations)**
    - **Primary Adapters:** Convert external input (e.g., HTTP requests) into core method calls.
    - **Secondary Adapters:** Implement outbound ports for specific technologies (e.g., PostgreSQL, Redis).

## Benefit in Go
- **Testability:** Mock adapters (e.g., in-memory databases) simplify unit testing.
- **Flexibility:** Swap databases, UIs, or APIs without changing core logic.
- **Clear Separation:** Business logic remains untouched by infrastructure changes.


## Project Structure
Here is implementation of port and adapter (hexagonal architecture) on golang.

```plaintext
├── cmd/                                    #⚡Primary Adapters (Driving Side)
│   ├── api/                                # Primary Adapter: HTTP entry (User Interface)
│   ├── grpc/                               # Primary Adapter: gRPC entry (User Interface)
│   ├── consumer/                           # Primary Adapter: Kafka message consumer
│   ├── publisher/                          # Primary Adapter: Kafka message producer
│   ├── migrate/                            # Infrastructure Tool: DB migrations
│   └── job/                                # Primary Adapter: Cron job scheduler
│
├── config/                                 #🔌Infrastructure
│   ├── config.yaml                         # Main Configuration File
│   ├── config.local.yaml                   # Local Overrides
│   ├── config.dev.yaml                     # Dev Environment Config
│   ├── config.prod.yaml                    # Production Environment Config
│
├── internal/
│   ├── domain/                             # 💎 Core Domain (Business Models, Rules, & Validation)
│   │   ├── wallet.go                       # Wallet Entity & Methods
│   │   ├── transaction.go                  # Transaction Entity & Methods
│   │   ├── user.go                         # User Entity & Methods
│   │
│   ├── usecase/                            # ⚙️ Application Core (Orchestrates domain+ports & Business workflows)
│   │   ├── wallet_service.go               # Wallet Logic
│   │   ├── payment_service.go              # Third-Party Payment Logic
│   │
│   ├── ports/                              # 🔗Ports (Abstract Interfaces for Adapters)
│   │   ├── wallet_repository.go            # DB Interface for Wallet
│   │   ├── payment_gateway.go              # Interface for Midtrans, Doku, Stripe
│   │   ├── event_publisher.go              # Kafka Producer Interface
│   │   ├── event_consumer.go               # Kafka Consumer Interface
│   │
│   ├── adapters/                           #🔌 Secondary Adapters (Driven Side)
│   │   ├── postgres_wallet_repo.go         # PostgreSQL Adapter Implementation
│   │   ├── redis_cache.go                  # Redis Adapter Implementation
│   │   ├── mongo_transaction_repo.go       # MongoDB Adapter Implementation
│   │   ├── kafka_publisher.go              # Kafka Producer Implementation
│   │   ├── kafka_consumer.go               # Kafka Consumer Impelementation
│   │   ├── midtrans_gateway.go             # Midtrans Payment Adapter Implementation
│   │   ├── doku_gateway.go                 # Doku Payment Adapter Implementation
│   │   ├── stripe_gateway.go               # Stripe Payment Adapter Implementation
│
├── migrations/                             #🗄️Infrastructure (DB Schema Management)
│   ├── 001_create_wallets.up.sql
│   ├── 001_create_wallets.down.sql
│   ├── 002_create_transactions.up.sql
│   ├── 002_create_transactions.down.sql  
│
├── api/                                    #🖥️ User Interface Layer
│   ├── rest/
│   │   ├── handlers/
│   │   │   ├── wallet_handler.go           # Wallet HTTP Handlers
│   │   │   ├── transaction_handler.go      # Transaction HTTP Handlers
│   │   ├── routes.go                       # REST API Router
│   ├── grpc/
│   │   ├── wallet_service.proto            # gRPC Definition
│   │   ├── transaction_service.proto       # gRPC Definition
│
├── tests/
│   ├── wallet_test.go                      # Unit Test for Wallet Logic
│   ├── transaction_test.go                 # Unit Test for Transactions
│
├── docker-compose.yml                      # Services for PostgreSQL, Redis, Kafka, etc.
├── Makefile                                # Useful Commands (build, test, run)
└── go.mod                                  # Go Module Definition
```

## Hexagonal Architecture Mapping 
| Component              | Hexagonal Term       | Project Location                     | Key Characteristics                              |
|------------------------|----------------------|--------------------------------------|--------------------------------------------------|
| **Business Rules**     | Core Domain          | `internal/domain/`                   | Zero external dependencies                       |
| **Application Workflow** | Application Core   | `internal/usecase/`                  | Orchestrates domain logic & port interactions    |
| **Input Contracts**    | Driving Ports        | `internal/ports/event_consumer.go`   | Define how the system receives external input    |
| **Output Contracts**   | Driven Ports         | `internal/ports/*_repository.go`     | Define how the system interacts with the world   |
| **User Facing**        | Primary Adapters     | `cmd/api/`, `api/rest/handlers/`     | HTTP/gRPC/Kafka input handlers                   |
| **Infrastructure**     | Secondary Adapters   | `internal/adapters/`                 | DBs, payment gateways, message brokers           |
| **Configuration**      | Infrastructure       | `config/`                            | Environment-specific implementations            |

### Notes:
- **Driving Ports**: Act as input interfaces (e.g., `EventConsumer` for Kafka messages).  
- **Driven Ports**: Act as output interfaces (e.g., `WalletRepository` for database operations).  
- **Primary Adapters**: Entry points for user/system input (HTTP handlers, message consumers).  
- **Secondary Adapters**: Implement Driven Ports to connect to external services (PostgreSQL, Stripe).  
- **Core Domain**: Contains *pure business logic* with no dependencies on frameworks or tools.  

## Driving & Driven Side Explanation
1. **Driving Side (Left)**
  - **Primary Adapters:** `cmd/` (HTTP/gRPC servers), `api/rest/handlers/`
  - **Ports:** `ports/event_consumer.go` (input contracts)
  - **Flow:** 
    ```
      HTTP Request → Primary Adapter (handler) → Application Core → Domain Driven Side (Right)
    ```
2. **Driven Side (Right)**
  - **Secondary Adapters:** `adapters/postgres_wallet_repo.go`
  - **Ports:** `ports/wallet_repository.go` (output contracts)
  - **Flow:**
  ```
    Domain → Application Core → Driven Port → Secondary Adapter → PostgreSQL
  ```

## Component Relationships
![Hexagonal Architecture Flow](documentations/hexagonal_architecture_flow.png)

## Mini E-Wallet Project with Ports and Adapters (Hexagonal) Architecture Implementation

### 1. ERD Design for Defining Data Models
```sql
Table users as U {
  id int [pk, increment, not null]
  fullname varchar [not null]
  email varchar [unique, not null]
  phone varchar [unique, not null]
  status varchar
  created_at datetime [not null, default: `now()`]
  updated_at datetime [not null, default: `now()`]
}

Table wallets as W {
  id int [pk, increment, not null]
  user_id int [Ref: - U.id, not null]
  balance int [not null, default: 0]
  currency_code varchar [not null]
  description varchar [null, note: 'optional notes']
  status varchar
  created_at datetime [not null, default: `now()`]
  updated_at datetime [not null, default: `now()`]
}

Table Transactions as T {
  id int [pk, increment, not null]
  wallet_id int [Ref: > W.id, not null]
  type varchar [not null, note: 'Transaction Type (deposit, withdrawal, transfer)']
  amount int [not null]
  status varchar [not null, note: 'pending, completed, failed']
  created_at datetime [not null, default: `now()`]
  updated_at datetime [not null, default: `now()`]
}

Table Payments as P {
  id int [pk, increment, not null]
  transaction_id int [Ref: - T.id, not null]
  gateway int [not null, note: 'Payment gateway (Midtrans, Doku, Stripe)']
  external_reference varchar [note: 'External reference ID']
  details json [null, note: 'extra info from gateway']
  status varchar
  created_at datetime [not null, default: `now()`]
  updated_at datetime [not null, default: `now()`]
}
```
### 2. Mini E-Wallet Key Usecases
Based on industry standards and common functionalities, the minimalized primary use cases include:

  - **Wallet Management:** Enabling users to view and manage their wallet balances and transaction histories.
  ![Get Fund Balance](documentations/get_fund_balance.png)
  - **Adding Funds to Wallet:** Allowing users to deposit money into their wallets.  
    ![Add Fund to Wallet](documentations/add_fund_to_wallet.png)
  - **Transferring Funds:** Facilitating peer-to-peer transfers between users within the system.​
    ![Transfer Fund to Users](documentations/transfer_fund_to_users.png)
  - **Withdrawing Funds:** Permitting users to withdraw money from their wallets to external accounts.
  ![Fund Withdrawal from Wallet](documentations/fund_withdrawal_from_wallet.png)

## References
- [Netflix TechBlog: Ready for Changes with Hexagonal Architecture](https://netflixtechblog.com/ready-for-changes-with-hexagonal-architecture-b315ec967749)
- [Hexagonal Architecture: There are Always Two Sides to Every Story](https://medium.com/ssense-tech/hexagonal-architecture-there-are-always-two-sides-to-every-story-bc0780ed7d9c)
- [Why Clean Architecture Struggles in Golang and What Works Better](https://dev.to/lucasdeataides/why-clean-architecture-struggles-in-golang-and-what-works-better-m4g)
