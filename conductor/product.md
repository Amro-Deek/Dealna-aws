# Initial Concept
Dealna is a secure university marketplace platform where students can buy/sell services and items within their university ecosystem.

# Architecture & Engineering Rules
- **Role:** Senior Backend Architect and Technical Partner
- **Architecture:** Hexagonal Architecture (Ports & Adapters) and Clean Architecture principles.
- **Backend Stack:** Go (Golang), PostgreSQL, SQLC, Keycloak (migrating from JWT), REST APIs.
- **Layers:**
  - `core/services`: Business logic
  - `core/ports`: Interfaces
  - `adapters/secondary`: Infrastructure implementations
  - `adapters/primary`: HTTP controllers
  - `core/domain`: Domain models
- **Rules:** No infrastructure dependency inside the core layer. SQL access goes through sqlc generated queries. Proper error handling using ErrorFrame. No business logic in handlers.

# Current Goal
Implement a complete authentication track using Keycloak, specifically the **Signup Flow** where:
1. User registers in Keycloak.
2. Keycloak creates identity.
3. Backend stores additional domain data.
4. Backend links `keycloak_sub` to local User.

Remaining APIs to build/adapt: signup, refresh token, logout, email verification, activation flow, password reset.
