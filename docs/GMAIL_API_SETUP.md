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

### 4. Configure Environment Variables

Add the following to your `.env` file or set as environment variables:

```env
EMAIL_CREDENTIALS_PATH=credentials.json
EMAIL_TOKEN_PATH=token.json
EMAIL_FROM_EMAIL=your-email@gmail.com
EMAIL_FROM_NAME=Lam Phuong
```

**Note:** 
- `credentials.json` should be the path to your downloaded OAuth credentials file
- `token.json` will be created automatically after the first OAuth authorization
- `EMAIL_FROM_EMAIL` should be the Gmail address you want to send emails from

### 5. First-Time Authorization

When you first run the application:

1. The application will print an authorization URL to the console
2. Open the URL in your browser
3. Sign in with the Google account you want to use for sending emails
4. Grant the necessary permissions
5. Copy the authorization code from the browser
6. Paste it into the console when prompted
7. The application will save the token to `token.json` for future use

**Important:** After the first authorization, `token.json` will be created and reused automatically. You won't need to authorize again unless you revoke access or delete the token file.

## Configuration

The email service uses the following configuration options:

- `EMAIL_CREDENTIALS_PATH`: Path to the OAuth credentials JSON file (default: `credentials.json`)
- `EMAIL_TOKEN_PATH`: Path where the OAuth token will be stored (default: `token.json`)
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

### "Unable to read client secret file"
- Make sure `credentials.json` exists in the specified path
- Check that the path in `EMAIL_CREDENTIALS_PATH` is correct

### "Unable to retrieve token from web"
- Make sure you copied the entire authorization code
- The code expires quickly, so paste it immediately after copying

### "Failed to send email via Gmail API"
- Verify that the Gmail API is enabled in your Google Cloud project
- Check that the OAuth consent screen is properly configured
- Ensure the token hasn't expired (refresh tokens are used automatically)

### Token Expired
- If the token expires, delete `token.json` and re-run the authorization flow
- The refresh token should automatically renew the access token

## References

- [Gmail API Go Quickstart](https://developers.google.com/workspace/gmail/api/quickstart/go)
- [Gmail API Documentation](https://developers.google.com/gmail/api)
- [OAuth 2.0 for Google APIs](https://developers.google.com/identity/protocols/oauth2)

