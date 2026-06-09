# Design Spec: Phone+Email Registration and Forgot/Reset Password Flow

This document outlines the architecture, database schema, API contracts, and UI flow for supporting user registration with both phone and email, and allowing users to retrieve/reset their passwords via email validation codes.

---

## 1. Goal Description

Currently, users register using only their phone number. If they forget their password, there is no self-service way to retrieve or reset it. 
To improve security and user convenience:
- Update the registration flow to require **both phone number and email**.
- Explicitly display a note that the email acts as the primary recovery credential.
- Build a public **Forgot Password / Reset Password** mechanism using database-backed verification codes.
- Add "Forgot Password" entry points and forms to both PC and H5 client apps.

---

## 2. Proposed Changes

### 2.1 Database Schema
We will create a new migration SQL file to define the verification codes table:
- **File**: `backend/migrations/20260609_05_password_reset_codes.sql`

```sql
CREATE TABLE IF NOT EXISTS `password_reset_codes` (
    `id` VARCHAR(64) NOT NULL,
    `email` VARCHAR(128) NOT NULL,
    `code` VARCHAR(10) NOT NULL,
    `status` VARCHAR(20) NOT NULL DEFAULT 'UNUSED', -- UNUSED, USED
    `expired_at` TIMESTAMP NOT NULL,
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    INDEX `idx_email_code` (`email`, `code`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

---

### 2.2 Backend API Contracts
We will expose two public endpoints in `backend/router/auth.go`:

#### A. Forgot Password (Request Verification Code)
- **Endpoint**: `POST /api/v1/auth/forgot-password`
- **Request Body**:
  ```json
  {
    "email": "user@example.com"
  }
  ```
- **Response Body**:
  ```json
  {
    "code": 0,
    "message": "success",
    "data": {
      "email": "user@example.com",
      "code": "123456" // Returned directly for local development and testing convenience
    }
  }
  ```
- **Logic**:
  1. Validate that `email` exists in the `users` table. If not, return `404 Not Found` (or `400 Bad Request` with message "email not found").
  2. Generate a random 6-digit numeric string as the verification code.
  3. Insert a record into `password_reset_codes` with status `UNUSED` and expiration set to `10 minutes` from now.
  4. Return the verification code in the response data payload.

#### B. Reset Password
- **Endpoint**: `POST /api/v1/auth/reset-password`
- **Request Body**:
  ```json
  {
    "email": "user@example.com",
    "code": "123456",
    "new_password": "NewSecurePassword123"
  }
  ```
- **Response Body**:
  ```json
  {
    "code": 0,
    "message": "密码重置成功",
    "data": {}
  }
  ```
- **Logic**:
  1. Validate that the new password meets security requirements (e.g. `min=8`).
  2. Query `password_reset_codes` for a record matching the `email`, `code`, and `status = 'UNUSED'` where `expired_at > NOW()`. If not found or expired, return `400 Bad Request` with "invalid or expired verification code".
  3. Update the `password_reset_codes` record's status to `USED`.
  4. Perform bcrypt hash on the `new_password`.
  5. Update the user's `password_hash` in the `users` table matching the `email`.
  6. Return success response.

---

### 2.3 Frontend Components

#### A. Registration Pages
Update frontend forms to require **both phone and email**:
- **PC View**: [Register.vue](file:///Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/user/Register.vue)
- **H5 View**: [Register.vue](file:///Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/user/Register.vue)
- **UI Details**:
  - Add an "邮箱" (Email) input field.
  - Right below the Email input, display: `邮箱为帐号找回，密码重置唯一凭证` in muted yellow/secondary font.
  - Require both phone and email in payload validation before permitting submission.

#### B. Forgot/Reset Password Flow
- **PC View**: Add new page `ForgotPassword.vue` at `newclient/src/apps/pc/views/user/ForgotPassword.vue`.
  - Add path `/forgot-password` in `newclient/src/apps/pc/router/index.js`.
- **H5 View**: Add new page `ForgotPassword.vue` at `newclient/src/apps/h5/views/user/ForgotPassword.vue`.
  - Add path `/profile/forgot-password` in `newclient/src/apps/h5/router/index.js`.
- **Login Pages**: Add a "忘记密码？" (Forgot Password?) text button link pointing to the respective forgot password paths.
- **UI Flow**:
  - Phase 1: Input email -> Click "获取验证码" (Get Code). Call `/auth/forgot-password`. Since the code is returned in the response payload for testing, auto-fill it into the code input.
  - Phase 2: Enter code, new password, and confirm new password -> Click "重置密码". Call `/auth/reset-password`.
  - Phase 3: Display success message and redirect back to `/login`.

---

## 3. Verification Plan

### 3.1 Automated & Compilation Verification
- Run `go build ./...` in the backend directory to check for compiler errors.
- Run `npm run build` in the newclient directory to verify correct compilation of all PC and H5 frontend files.

### 3.2 Manual API & UI Verification
1. Open the Registration page, fill in a new phone and email (ensuring the note is visible), click Register, and verify that the user is registered with both fields stored in the database.
2. Log out, click "忘记密码？" on the login page.
3. Input the registered email, click "获取验证码", and verify that the code returns in the response and is auto-filled.
4. Input the new password, submit, and verify that the password is reset successfully.
5. Try logging in with the new password to confirm.
