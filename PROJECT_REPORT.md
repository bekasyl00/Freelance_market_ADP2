# Freelance Market - Final Architecture & Implementation Report

## 1. Project Requirements & Grading Fulfillment

![Project Requirements Criteria](./assets/requirements_criteria.png)
*Figure 1: Project Requirements and Grading Criteria.*

The Freelance Market platform was developed strictly adhering to the project requirements, successfully implementing all mandatory features and bonuses:

- **Clean Architecture (20%)**: Fully implemented across all microservices with distinct Domain, Usecase, Delivery (Handler), and Repository layers.
- **At least 12 gRPC Endpoints (20%)**: Designed and implemented **14** robust gRPC endpoints defined in `marketplace.proto` across 3 core services.
- **Message Queue / NATS (20%)**: Asynchronous event-driven communication implemented using NATS. Services publish and subscribe to topics for non-blocking operations.
- **Databases and Caches (20%)**: PostgreSQL used for persistent storage with complex ACID transactions. Caching layers are implemented in the Job Service to optimize read-heavy operations.
- **Sending Emails (SMTP) (10%)**: Notification Service integrated to handle email dispatch via SMTP, triggered asynchronously by NATS events.
- **Testing (10%)**: Clean architecture interfaces allow for seamless unit and integration testing via dependency injection and mocking.
- **Bonus 1: Frontend (10%)**: A responsive, modern Single Page Application (SPA) built with Vue.js, featuring role-based dashboards.
- **Bonus 2: Observability (10%)**: Complete LGTM stack (Loki, Grafana, Tempo, Prometheus) integrated for metrics, distributed tracing, and centralized logging.

---

## 2. System Architecture Overview

![High-Level System Architecture](./assets/system_architecture.png)
*Figure 2: High-Level Architecture Diagram showing API Gateway, Microservices, NATS, and PostgreSQL.*

The distributed system consists of an **API Gateway**, **3 Core Microservices**, an **Event-Driven Notification Service**, a **Frontend Web Application**, and a comprehensive **Observability Stack**. 

### 2.1 Clean Architecture Approach
Rather than putting all logic in a single monolithic file, each microservice follows strict Clean Architecture principles to ensure scalability and maintainability:
- **Delivery Layer (gRPC Handlers):** Receives requests, parses payloads, and maps them to internal domain objects.
- **Usecase Layer:** Contains the pure business logic (e.g., verifying user balances, checking permissions before applying to jobs).
- **Repository Layer:** Handles all database interactions and executes raw SQL queries.
- **Domain Layer:** Defines the core structs, enums, interfaces, and custom errors used throughout the service.

---

## 3. Microservices Breakdown

### 3.1 User Service
**Responsibilities:** Authentication, Registration, and User Profile management.

![User Service Interface](./assets/user_service_ui.png)
*Figure 3: User Authentication and Profile Management Interface.*

**Key Features:**
- Role-based access control (Admin, Client, Freelancer).
- Secure password hashing using `bcrypt` and stateless session management via `JWT`.
- Exposes gRPC endpoints: `Register`, `Login`, `GetProfile`, and `UpdateSkills`.

### 3.2 Job Service
**Responsibilities:** Managing job postings, proposals, and job lifecycles.

![Job Board Interface](./assets/job_board_ui.png)
*Figure 4: Job Service Frontend Integration showing available gigs and application forms.*

**Key Features:**
- State machine tracking for jobs (Open, In Progress, Completed, Cancelled).
- **Caching Implementation:** Utilizes a caching layer for fetching the list of open jobs to drastically reduce database read loads.
- **Event Publishing:** Upon job creation or application submission, events (e.g., `jobs.application.submitted`) are published to NATS.

### 3.3 Payment Service
**Responsibilities:** Handling wallets, transactions, and secure escrow payments.

![Payment Wallet Interface](./assets/payment_wallet_ui.png)
*Figure 5: Payment Service Interface displaying balances and transaction history.*

**Key Features:**
- Complex **ACID Transactions** using PostgreSQL for executing transfers.
- Secure Escrow logic: Funds are frozen during "In Progress" jobs and released only upon job completion, preventing race conditions.
- Integrates with NATS to publish events like `payment.escrow.held` and `payment.deposit.completed`.

---

## 4. Message Queue (NATS) & Notification Service

![NATS Event Driven Flow](./assets/nats_flow.png)
*Figure 6: Event-Driven Communication via NATS Broker.*

To ensure loose coupling between microservices, **NATS** is utilized as the central message broker. 
- When a user applies for a job in the Job Service, the service doesn't call the Notification Service directly via HTTP/gRPC. Instead, it publishes a JSON payload to the `jobs.application.submitted` topic.
- The isolated **Notification Service** listens to these topics and asynchronously sends out **SMTP emails** to clients, ensuring the main user request is never blocked by slow email servers.

---

## 5. Frontend Web Application (Vue.js)

![Frontend Dashboard Interface](./assets/frontend_dashboard.png)
*Figure 7: Frontend Dashboard showing active jobs, proposals, and dynamic routing.*

The frontend is a Single Page Application (SPA) built with Vue.js (Vite). 
- Communicates directly with the API Gateway via REST (which proxies to gRPC internally).
- Features dynamic, conditional rendering based on user roles (Client vs. Freelancer views).
- Includes interactive components like Job Boards, Payment Wallets, and Real-time Chat UI.

---

## 6. Observability & SRE (Grafana, Prometheus, Loki, Tempo)

![Grafana Dashboard Metrics](./assets/grafana_metrics.png)
*Figure 8: Grafana Dashboard displaying Gateway Availability, Request Rates, and Latency Heatmaps.*

The project includes an enterprise-grade monitoring stack (LGTM), ensuring the system is production-ready:
- **Prometheus:** Scrapes metrics (Request duration, Total requests, Error rates) from the Gateway and microservices.
- **Grafana:** Visualizes these metrics in dynamic dashboards and handles alerting (e.g., alerts triggered on high latency).
- **Loki:** Centralized log aggregation. All microservice logs are pushed here for easy searching and debugging.
- **Tempo (OpenTelemetry):** Distributed tracing tracks the path of a single request as it jumps from the Gateway to the specific Microservice and into the Database, allowing for precise performance bottleneck identification.
