# Gmail API Setup Guide

This guide explains how to set up the Gmail API for sending emails, following the [Google Gmail API Go quickstart](https://developers.google.com/workspace/gmail/api/quickstart/go).

## Prerequisites

- Latest version of Go
- A Google Cloud project
- A Google account with Gmail enabled

## Setup Steps

### 1. Enable the Gmail API

1. Go to the [Google Cloud Console](https://console.cloud.google.com/)
2. Select your project (or create a new one)
3. Navigate to **APIs & Services** > **Library**
4. Search for "Gmail API"
5. Click on **Gmail API** and click **Enable**

### 2. Configure the OAuth Consent Screen

1. In the Google Cloud Console, go to **APIs & Services** > **OAuth consent screen**
2. If you see "Google Auth platform not configured yet", click **Get Started**
3. Under **App Information**:
   - Enter an **App name** (e.g., "Lam Phuong API")
   - Choose a **User support email** address
   - Click **Next**
4. Under **Audience**, select **Internal** (for testing) or **External** (for production)
5. Click **Next**
6. Under **Contact Information**, enter an **Email address**
7. Click **Next**
8. Review and accept the Google API Services User Data Policy
9. Click **Continue** and then **Create**

### 3. Create OAuth 2.0 Credentials

1. In the Google Cloud Console, go to **APIs & Services** > **Credentials**
2. Click **Create Credentials** > **OAuth client ID**
3. Select **Application type** > **Desktop app**
4. Enter a **Name** for the credential (e.g., "Lam Phuong API Client")
5. **Important:** In the **Authorized redirect URIs** section, add:
   - `http://localhost:8082/oauth2callback`
   - `http://localhost:8083/oauth2callback`
   - `http://localhost:8084/oauth2callback`
   - `http://localhost:8085/oauth2callback`
   
   (The application will automatically find an available port from these options)
6. Click **Create**
7. Download the JSON file and save it as `credentials.json` in your project root directory

**Note:** The application uses a local HTTP server to receive the OAuth callback. When you authorize, Google will redirect to one of the configured localhost ports, and the application will automatically capture the authorization code. The server will automatically shut down after receiving the code.

### 4. Get Refresh Token

To get a refresh token, you need to complete the OAuth flow once. You can use a tool or script to do this:

1. Use the OAuth 2.0 Playground: https://developers.google.com/oauthplayground/
   - Click the gear icon (⚙️) in the top right
   - Check "Use your own OAuth credentials"
   - Enter your Client ID and Client Secret
   - In the left panel, find "Gmail API v1" and select `https://www.googleapis.com/auth/gmail.send`
   - Click "Authorize APIs"
   - Sign in and grant permissions
   - Click "Exchange authorization code for tokens"
   - Copy the "Refresh token" value

2. Or use a simple script to generate the refresh token (see below)

### 5. Configure Environment Variables

Add the following to your `.env` file or set as environment variables:

```env
EMAIL_CLIENT_ID=your-client-id.apps.googleusercontent.com
EMAIL_CLIENT_SECRET=your-client-secret
EMAIL_REFRESH_TOKEN=your-refresh-token
EMAIL_FROM_EMAIL=your-email@gmail.com
EMAIL_FROM_NAME=Lam Phuong
```

**Note:** 
- `EMAIL_CLIENT_ID` and `EMAIL_CLIENT_SECRET` come from your OAuth 2.0 credentials in Google Cloud Console
- `EMAIL_REFRESH_TOKEN` is obtained from the OAuth flow (see step 4)
- `EMAIL_FROM_EMAIL` should be the Gmail address you want to send emails from
- The refresh token doesn't expire (unless revoked), so you only need to get it once

## Configuration

The email service uses the following configuration options:

- `EMAIL_CLIENT_ID`: Gmail OAuth Client ID (from Google Cloud Console)
- `EMAIL_CLIENT_SECRET`: Gmail OAuth Client Secret (from Google Cloud Console)
- `EMAIL_REFRESH_TOKEN`: Gmail OAuth Refresh Token (obtained from OAuth flow)
- `EMAIL_FROM_EMAIL`: The Gmail address to send emails from
- `EMAIL_FROM_NAME`: Display name for the sender

## Security Notes

1. **Never commit credentials.json or token.json to version control**
   - Add them to `.gitignore`:
     ```
     credentials.json
     token.json
     ```

2. **File Permissions**
   - The `token.json` file is created with restricted permissions (0600)
   - Keep `credentials.json` secure and limit access

3. **Scopes**
   - The application uses `gmail.GmailSendScope` which allows sending emails only
   - This is the minimum required scope for sending emails

## Testing

Use the test endpoint to verify email functionality:

```bash
curl -X POST http://localhost:8080/api/email/test \
  -H "Content-Type: application/json" \
  -d '{"email": "recipient@example.com"}'
```

## Troubleshooting

### "Unable to retrieve Gmail client"
- Verify that `EMAIL_CLIENT_ID`, `EMAIL_CLIENT_SECRET`, and `EMAIL_REFRESH_TOKEN` are set correctly
- Check that the refresh token is valid and hasn't been revoked
- Ensure the Gmail API is enabled in your Google Cloud project

### "Failed to send email via Gmail API"
- Verify that the Gmail API is enabled in your Google Cloud project
- Check that the OAuth consent screen is properly configured
- Ensure the refresh token is valid and hasn't expired
- Verify that `EMAIL_FROM_EMAIL` matches the Gmail account associated with the refresh token

### Refresh Token Invalid
- If the refresh token is invalid or expired, you need to generate a new one
- Use the OAuth 2.0 Playground or a script to generate a new refresh token
- Update `EMAIL_REFRESH_TOKEN` in your environment variables

## References

- [Gmail API Go Quickstart](https://developers.google.com/workspace/gmail/api/quickstart/go)
- [Gmail API Documentation](https://developers.google.com/gmail/api)
- [OAuth 2.0 for Google APIs](https://developers.google.com/identity/protocols/oauth2)

