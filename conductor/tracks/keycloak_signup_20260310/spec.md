# Track Specification: Implement Keycloak Signup Flow

## Goal
Migrate the student registration (signup) flow to use Keycloak as the Identity Provider, completely removing local password hashing from the `User` creation process.

## Scope
1. **Ports:** Extend `IIdentityProvider` with a `RegisterUser` method that accepts email and password and returns a `keycloakSub`. Update `IUserRepository` to accept `keycloak_sub` instead of `password_hash`.
2. **Infrastructure (Adapters):** Implement `RegisterUser` in the Keycloak adapter by calling the Keycloak Admin API. Update `user.sql` and run `sqlc generate` to ensure `keycloak_sub` is correctly handled by the Postgres adapter.
3. **Core (Services):** Refactor `StudentRegistrationService.CompleteStudentRegistration` to verify the email token, call the new Keycloak registration method, and then persist the resulting `keycloak_sub` along with domain-specific user details in the local database.

## Assumptions & Dependencies
- Pre-registration email flow remains entirely local in Dealna to ensure domain verification (`student.birzeit.edu`) before interacting with Keycloak.
- Keycloak will be configured with a Client Secret or Admin User to allow the backend to create users via the Admin API.
- Hexagonal architecture and Clean Architecture principles are strictly followed (no infrastructure code in `core/services` and no business logic in `adapters/primary`).