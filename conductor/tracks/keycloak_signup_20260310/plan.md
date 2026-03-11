# Implementation Plan: Implement Keycloak Signup Flow

## Phase 1: Update IIdentityProvider and IUserRepository
- [x] Task: Update `database/query/user.sql` to accept `keycloak_sub` instead of `password_hash` during student creation
    - [x] Write Tests (if applicable)
    - [x] Implement Feature
- [x] Task: Regenerate sqlc models via `make sqlc`
    - [x] Write Tests (if applicable)
    - [x] Implement Feature
- [x] Task: Update `ports.IUserRepository.CreateStudent` signature
    - [x] Write Tests (if applicable)
    - [x] Implement Feature
- [x] Task: Update `ports.IIdentityProvider` with `RegisterUser` method
    - [x] Write Tests (if applicable)
    - [x] Implement Feature
- [x] Task: Conductor - User Manual Verification 'Phase 1: Update IIdentityProvider and IUserRepository' (Protocol in workflow.md)

## Phase 2: Implement Keycloak Adapter
- [x] Task: Implement `RegisterUser` in `KeycloakIdentityProvider` using Keycloak Admin API
    - [x] Write Tests (if applicable)
    - [x] Implement Feature
- [x] Task: Update backend configuration to supply Keycloak Admin credentials (client secret or user/pass)
    - [x] Write Tests (if applicable)
    - [x] Implement Feature
- [x] Task: Conductor - User Manual Verification 'Phase 2: Implement Keycloak Adapter' (Protocol in workflow.md)

## Phase 3: Update Core Service
- [x] Task: Refactor `CompleteStudentRegistration` in `StudentRegistrationService` to use Keycloak adapter
    - [x] Write Tests
    - [x] Implement Feature
- [x] Task: Ensure all HTTP handlers gracefully forward the new request structures if required
    - [x] Write Tests
    - [x] Implement Feature
- [x] Task: Conductor - User Manual Verification 'Phase 3: Update Core Service' (Protocol in workflow.md)