# Create User Authentication Flow for User Sign In Specification

## 1. Overview
This specification outlines the requirements for implementing a secure and robust user authentication flow for signing into the Dealna application. The primary goal is to allow verified university students and approved external service providers to securely log in to the platform.

## 2. Goals
- Enable secure user authentication via email and password.
- Verify account status and role during login.
- Implement session management using JWT.
- Ensure single active session per user.
- Provide a mechanism for password reset.

## 3. Functional Requirements
### FR-3.1 Login Input
- The system shall allow users to sign in using their registered email and password.

### FR-3.2 Account Status Verification During Login
- The system shall verify that the account is activated.
- The system shall verify that the provider is approved (if applicable).
- The system shall verify that the account is not suspended or restricted.

### FR-3.3 Secure Login
- The system shall authenticate users using email and password.

### FR-3.4 Password Hashing
- Passwords shall be stored in hashed and encrypted form.

### FR-3.5 Failed Login Attempts
- The system shall temporarily restrict login attempts after multiple consecutive failed authentication attempts.

### FR-3.6 Token-Based Sessions (JWT)
- The system shall use secure tokens (JWT) for session management.

### FR-3.7 Single Active Session Enforcement
- The system shall allow only one active session per user. Logging in from a new device shall invalidate previous sessions.

### FR-3.8 Session Expiration
- Inactive sessions shall automatically expire after a predefined period of inactivity.

### FR-3.9 Session Invalidation on Logout
- The system shall invalidate the active session token immediately upon user logout to prevent further use.

### FR-3.10 Forgot Password - Activation Code Reset
- The system shall allow users to reset their password by sending an activation code to their registered email or phone number. The user must enter the code before creating a new password.

## 4. Non-Functional Requirements
### NFR-4.1 Password Policy Constraints
- Minimum length: 8 characters.
- Mixed uppercase and lowercase letters.
- At least one number and one symbol.

### NFR-4.2 Session Security Constraints
- Session inactivity timeout shall be set to 25 minutes, after which the user is automatically logged out.
- Only one active session is allowed per user at any given time.
- A new session token is generated for each successful login.
- Session tokens become invalid immediately after user logout.

## 5. Out of Scope
- Social login (e.g., Google, Facebook).
- Multi-factor authentication (MFA).
- Biometric authentication.
