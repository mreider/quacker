### Backend Requirements:
1. Python backend using Flask, Gunicorn, and Nginx.
2. REST API with JSON communication.
3. Environment variables needed:
   - `AWS_URL`: The endpoint for AWS-compatible object store.
   - `AWS_KEY` and `AWS_SECRET`: Credentials for S3 access.
   - `TOKEN_SECRET`: Secret key for generating expiring tokens (15 characters).
4. Interactive installation script (Bash + Python) for Ubuntu setup using systemd, including:
   - Prompt for domain name.
   - Email address for Let's Encrypt SSL certificate.
   - AWS URL, key, and secret.
   - TOKEN_SECRET value.
5. Flood control to prevent User 1 from brute-forcing tokens.
6. Expiring token format: Includes bucket name, expiration time, and a hash for validation.

### Backend REST API:
1. `/validate_token` (POST):
   - Input: Token from User 1.
   - Validates the token and bucket name.
   - Returns the list of pre-signed S3 URLs for downloading.

### Interactive Installation Script:
- Asks for necessary input (domain, email, AWS keys).
- Automatically installs Flask, Gunicorn, Nginx, configures Let's Encrypt SSL.
- Sets up service using systemd.

### Front End Requirements (Electron Apps):
1. Electron Frontend for User 1 (Downloader):
   - Text input for Token.
   - Send token to backend via POST request.
   - If valid, display list of presigned URLs in a textarea, with each URL on a new line.
   - Error messages for invalid tokens or bucket names ("Tell the admin...").
2. Electron Frontend for User 2 (Token Generator):
   - Input for `TOKEN_SECRET`, `bucket_name`, and `valid_minutes`.
   - Validates inputs to ensure non-empty and numeric values.
   - Upon submission, generate a token and display countdown timer for expiration.

### GitHub Action:
- Workflow file to build Electron apps for both User 1 and User 2.
- Triggered based on user input with version number prompt.
- Jobs to install Node.js, Electron, and package the apps.
- Archive and save Electron builds as artifacts for distribution.

### README File:
1. Overview of the project.
2. Instructions for running the backend installation script on Ubuntu.
3. Detailed API documentation for `/validate_token`.
4. How to run Electron apps for User 1 and User 2.
5. Troubleshooting tips and requirements.

### Notes:
- Ensure environment variables are set properly on the backend.
- Electron apps handle token validation locally before contacting the backend.
- Flood control is crucial to prevent brute force attacks by User 1.
