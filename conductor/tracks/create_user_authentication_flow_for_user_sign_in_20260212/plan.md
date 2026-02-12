# Implementation Plan: Create User Authentication Flow for User Sign In

This plan details the steps to implement the user authentication flow for signing into the Dealna application.

## Phase 1: User Registration and Email Verification

- [ ] Task: Design and implement database schema for user registration.
    - [ ] Write Tests: For user schema validation and storage.
    - [ ] Implement Feature: Update existing database migrations or create new ones for user table.
- [ ] Task: Implement backend API endpoint for user registration.
    - [ ] Write Tests: For successful user registration, invalid input, and duplicate email handling.
    - [ ] Implement Feature: Create API endpoint to register new users with email, password, and optional details.
- [ ] Task: Implement email verification service.
    - [ ] Write Tests: For sending unique activation links and verifying them.
    - [ ] Implement Feature: Integrate email service to send activation links and verify user accounts upon click.
- [ ] Task: Implement password hashing and secure storage.
    - [ ] Write Tests: For correct password hashing and comparison during login.
    - [ ] Implement Feature: Apply bcrypt hashing with salting for all new user passwords.
- [ ] Task: Conductor - User Manual Verification 'User Registration and Email Verification' (Protocol in workflow.md)

## Phase 2: User Login and Session Management

- [ ] Task: Implement backend API endpoint for user login.
    - [ ] Write Tests: For successful login, invalid credentials, and account status checks.
    - [ ] Implement Feature: Create API endpoint for user login, returning JWT upon success.
- [ ] Task: Implement JWT-based session creation and validation.
    - [ ] Write Tests: For JWT generation, parsing, and expiration.
    - [ ] Implement Feature: Generate and validate JWTs for authenticated sessions.
- [ ] Task: Implement single active session enforcement.
    - [ ] Write Tests: For invalidating previous sessions on new login.
    - [ ] Implement Feature: Ensure only one active session token is valid per user at any time.
- [ ] Task: Implement session expiration and invalidation on logout.
    - [ ] Write Tests: For automatic session expiry and immediate invalidation on logout.
    - [ ] Implement Feature: Configure JWT expiry and provide an API endpoint for logout.
- [ ] Task: Conductor - User Manual Verification 'User Login and Session Management' (Protocol in workflow.md)

## Phase 3: Password Reset and Account Status Handling

- [ ] Task: Implement backend API endpoint for password reset request.
    - [ ] Write Tests: For sending password reset activation codes to registered email/phone.
    - [ ] Implement Feature: Create API endpoint to initiate password reset flow.
- [ ] Task: Implement backend API endpoint for password reset confirmation.
    - [ ] Write Tests: For verifying activation code and updating password securely.
    - [ ] Implement Feature: Create API endpoint to allow users to set a new password after verifying the code.
- [ ] Task: Implement handling of failed login attempts.
    - [ ] Write Tests: For temporary lockout after multiple failed attempts.
    - [ ] Implement Feature: Introduce rate limiting for login attempts.
- [ ] Task: Implement account status verification during login.
    - [ ] Write Tests: For suspended, restricted, or unapproved provider accounts.
    - [ ] Implement Feature: Enhance login logic to check full account status.
- [ ] Task: Conductor - User Manual Verification 'Password Reset and Account Status Handling' (Protocol in workflow.md)
