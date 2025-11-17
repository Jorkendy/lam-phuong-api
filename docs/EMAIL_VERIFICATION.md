# Email Verification Implementation

This document describes the email verification system implemented for user registration.

## Overview

When a user registers through the public registration endpoint (`POST /api/auth/register`), they receive an email with a verification link. The user must click the link to verify their email address before they can log in.

## Features

1. **User Status**: Users have a `status` field that can be:
   - `pending` - Email not yet verified (default for public registration)
   - `active` - Email verified and account active

2. **Verification Token**: A secure random token is generated and stored with the user account

3. **Email Sending**: Verification email is sent automatically upon registration

4. **Login Protection**: Users with `pending` status cannot log in until they verify their email

5. **Admin Bypass**: Users created by admins are automatically set to `active` status

## User Model Changes

### New Fields

- `Status` (string): User account status (`pending` or `active`)
- `EmailVerificationToken` (string): Token used for email verification (never serialized in JSON)

### Airtable Fields

- `Status` - User status field
- `Email Verification Token` - Verification token field

## API Endpoints

### 1. Register User (Public)

**Endpoint**: `POST /api/auth/register`

**Request**:
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response** (201 Created):
```json
{
  "success": true,
  "data": {
    "id": "123",
    "email": "user@example.com",
    "role": "User",
    "status": "pending"
  },
  "message": "User registered successfully. Please check your email to verify your account."
}
```

**Behavior**:
- Creates user with `status: "pending"`
- Generates verification token
- Sends verification email
- Does NOT return JWT token (user must verify email first)

### 2. Verify Email

**Endpoint**: `GET /api/auth/verify-email?token=<verification_token>`

**Response** (200 OK):
```json
{
  "success": true,
  "data": {
    "id": "123",
    "email": "user@example.com",
    "role": "User",
    "status": "active"
  },
  "message": "Email verified successfully. You can now log in."
}
```

**Behavior**:
- Finds user by verification token
- Updates status to `active`
- Clears verification token
- User can now log in

### 3. Login (Updated)

**Endpoint**: `POST /api/auth/login`

**Behavior**:
- Checks if user status is `active`
- Returns 403 Forbidden if status is `pending`:
```json
{
  "success": false,
  "error": {
    "code": "FORBIDDEN",
    "message": "Please verify your email address before logging in. Check your email for the verification link."
  }
}
```

## Configuration

Add the following environment variables for email configuration:

```bash
# Email Configuration
EMAIL_SMTP_HOST=smtp.gmail.com          # SMTP server hostname
EMAIL_SMTP_PORT=587                    # SMTP server port
EMAIL_SMTP_USERNAME=your-email@gmail.com
EMAIL_SMTP_PASSWORD=your-app-password
EMAIL_FROM_EMAIL=noreply@lamphuong.com  # From email address
EMAIL_FROM_NAME=Lam Phuong             # From name
EMAIL_BASE_URL=http://localhost:8080   # Base URL for verification links
```

### Development Mode

If email configuration is not provided, the system will:
- Log email content to console instead of sending
- Still create users successfully
- Allow manual verification via API

## Email Template

The verification email contains:
- Subject: "Verify Your Email Address"
- Body: Includes verification link with token
- Link format: `{BASE_URL}/api/auth/verify-email?token={TOKEN}`

## Admin User Creation

When admins create users via `POST /api/users`:
- Users are automatically set to `status: "active"`
- No verification email is sent
- Users can log in immediately

## Security Considerations

1. **Token Generation**: Uses `crypto/rand` for secure random token generation (64-character hex string)

2. **Token Storage**: Verification tokens are never serialized in JSON responses

3. **Token Expiration**: Currently tokens don't expire, but this can be added by storing token creation timestamp

4. **Email Security**: Uses SMTP authentication for secure email sending

## Future Enhancements

1. **Token Expiration**: Add expiration time (e.g., 24 hours) for verification tokens
2. **Resend Verification**: Add endpoint to resend verification email
3. **Email Templates**: Support HTML email templates
4. **Rate Limiting**: Add rate limiting for verification attempts
5. **Token Cleanup**: Periodically clean up expired tokens

## Testing

### Without Email Configuration

1. Register a user - email will be logged to console
2. Copy the verification token from logs
3. Call verification endpoint: `GET /api/auth/verify-email?token=<token>`
4. User can now log in

### With Email Configuration

1. Configure SMTP settings in environment variables
2. Register a user - email will be sent
3. Click verification link in email
4. User can now log in

## Example Flow

1. **User Registration**:
   ```
   POST /api/auth/register
   → User created with status="pending"
   → Verification email sent
   → Response: "Please check your email to verify your account"
   ```

2. **Email Verification**:
   ```
   User clicks link: /api/auth/verify-email?token=abc123...
   → Status updated to "active"
   → Token cleared
   → Response: "Email verified successfully"
   ```

3. **Login**:
   ```
   POST /api/auth/login
   → Status check: must be "active"
   → JWT token returned
   ```

