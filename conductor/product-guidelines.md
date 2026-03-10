# Dealna Product & Engineering Guidelines

## Architectural Principles
1. **Hexagonal Architecture (Ports & Adapters):** Strict separation of concerns. The core domain and business logic must remain completely agnostic of external systems, frameworks, or databases.
2. **Clean Architecture:** Dependencies must point inward toward the core domain.
3. **Domain-Driven Design:** Align the code structure with the business domain.

## Code Organization & Layering
- `core/domain`: Contains pure Go structs representing business entities. No external dependencies.
- `core/ports`: Defines interfaces for all external interactions (repositories, email services, identity providers).
- `core/services`: Orchestrates business rules and use cases. No infrastructure dependencies.
- `adapters/primary`: HTTP controllers (Chi router) and DTOs. Responsible only for parsing requests, orchestrating service calls, and formatting responses. No business logic.
- `adapters/secondary`: Infrastructure implementations (e.g., PostgreSQL repositories, Keycloak integration, SMTP).

## Development Standards
- **Database Access:** SQL access must go exclusively through `sqlc` generated queries. Repositories should only interact with the database.
- **Error Handling:** Use `ErrorFrame` or standardized middleware error responses to ensure consistent API errors.
- **Security First:** Protect endpoints with appropriate middleware. Sensitive operations must be logged appropriately.