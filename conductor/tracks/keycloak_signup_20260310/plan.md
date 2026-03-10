# Implementation Plan: Implement Keycloak Signup Flow

## Phase 1: Update IIdentityProvider and IUserRepository
- [ ] Task: Update `database/query/user.sql` to accept `keycloak_sub` instead of `password_hash` during student creation
    - [ ] Write Tests (if applicable)
    - [ ] Implement Feature
- [ ] Task: Regenerate sqlc models via `make sqlc`
    - [ ] Write Tests (if applicable)
    - [ ] Implement Feature
- [ ] Task: Update `ports.IUserRepository.CreateStudent` signature
    - [ ] Write Tests (if applicable)
    - [ ] Implement Feature
- [ ] Task: Update `ports.IIdentityProvider` with `RegisterUser` method
    - [ ] Write Tests (if applicable)
    - [ ] Implement Feature
- [ ] Task: Conductor - User Manual Verification 'Phase 1: Update IIdentityProvider and IUserRepository' (Protocol in workflow.md)

## Phase 2: Implement Keycloak Adapter
- [ ] Task: Implement `RegisterUser` in `KeycloakIdentityProvider` using Keycloak Admin API
    - [ ] Write Tests (if applicable)
    - [ ] Implement Feature
- [ ] Task: Update backend configuration to supply Keycloak Admin credentials (client secret or user/pass)
    - [ ] Write Tests (if applicable)
    - [ ] Implement Feature
- [ ] Task: Conductor - User Manual Verification 'Phase 2: Implement Keycloak Adapter' (Protocol in workflow.md)

## Phase 3: Update Core Service
- [ ] Task: Refactor `CompleteStudentRegistration` in `StudentRegistrationService` to use Keycloak adapter
    - [ ] Write Tests
    - [ ] Implement Feature
- [ ] Task: Ensure all HTTP handlers gracefully forward the new request structures if required
    - [ ] Write Tests
    - [ ] Implement Feature
- [ ] Task: Conductor - User Manual Verification 'Phase 3: Update Core Service' (Protocol in workflow.md)