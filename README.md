# Quacker Admin Guide

This guide explains how to set up and manage the Quacker application on Ubuntu.

---

## Prerequisites

Ensure you have the following installed on your system:

### Certbot
Certbot is used to manage SSL certificates.
1. Install Certbot:
   ```bash
   sudo apt update
   sudo apt install certbot
   ```

2. Verify installation:
   ```bash
   certbot --version
   ```

### Redis
Redis is used for application data storage.

1. Install Redis:
   ```bash
   sudo apt update
   sudo apt install redis
   ```

2. Enable and start Redis as a systemd service:
   ```bash
   sudo systemctl enable redis-server
   sudo systemctl start redis-server
   ```

3. Verify Redis is running:
   ```bash
   redis-cli ping
   ```
   It should return `PONG`.

4. Install nginx
   ```bash
   sudo apt install nginx
   sudo apt install certbot python3-certbot-nginx -y
   ```

5. Configure nginx
   ```bash
   sudo nano /etc/nginx/sites-available/example.com
   ```

6. Add this to the file
   ```bash
  server {
    listen 80;
    server_name quacker.eu;

    # Redirect HTTP to HTTPS
    return 301 https://$host$request_uri;
   }
   
   ```

7. Enable the domain
   ```bash
   sudo ln -s /etc/nginx/sites-available/example.com /etc/nginx/sites-enabled/
   ```

8. Get a certificate
   ```bash
   sudo certbot --nginx -d example.com
   ```

9. Add the quacker process to nginx for HTTPS


10. Add the certificate and https info to the NGINX file
   ```bash

   server {
      listen 443 ssl;
      server_name example.com;

      location / {
         proxy_pass http://127.0.0.1:8085;
         proxy_set_header Host $host;
         proxy_set_header X-Real-IP $remote_addr;
         proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
         proxy_set_header X-Forwarded-Proto $scheme;
      }

      # Optional: Improve performance and logging
      access_log /var/log/nginx/quacker.access.log;
      error_log /var/log/nginx/quacker.error.log;

      # Additional settings for large uploads
      client_max_body_size 50M;
   }
   ```

11. Verify the config and restart nginx
   ```bash
   sudo nginx -t
   sudo systemctl reload nginx
   ```
---

## Installing Quacker

To add the `Quacker` binary to your system's `PATH` so you can run it from anywhere, follow these steps:

---

### Instructions:

1. **Download the Binary:**
   - Go to the [Releases](https://github.com/mreider/releases) section.
   - Download the appropriate binary for your operating system and architecture:
     - For Linux 64-bit systems: `quacker-linux-amd64`
     - For Linux ARM systems: `quacker-linux-arm64`

2. **Make the Binary Executable:**
   After downloading the binary, navigate to its directory and make it executable:
   ```bash
   chmod +x ./quacker-linux-amd64
   ```

3. **Move the Binary to a Directory in Your `PATH`:**
   The easiest way to make the binary globally accessible is to move it to a directory that's already in your system's `PATH`. A common choice is `/usr/local/bin`:
   ```bash
   sudo mv ./quacker-linux-amd64 /usr/local/bin/quacker
   ```
---

## Running Quacker

### Setup Configuration
Before running any other commands, set up the Quacker application:
```bash
quacker --setup
```
This command will prompt you to:
- Enter your Mailgun API key.
- Enter the hostname for your server (e.g., `example.com`).

Note: you might have to run via sudo in order to create certificates.


### Generate Invitation Codes
To create an invitation code:
```bash
quacker --generate
```
This generates a code and stores it in Redis with a 48-hour expiration.

### Start the Quacker Server
To start the Quacker application:
```bash
quacker --run
```
The server will start on port 443 using HTTPS.

Note: you might have to do this to allow 443 binding or run via sudo

```bash
 sudo setcap 'cap_net_bind_service=+ep' quacker
 ```


### Process Scheduled Jobs
To run the Quacker background job (e.g., for sending emails):
```bash
quacker --job
```
This can be scheduled using `cron`.

---

## Setting Up Quacker as a Systemd Service

### Create a Systemd Service File
1. Create a service file for Quacker:
   ```bash
   sudo nano /etc/systemd/system/quacker.service
   ```

2. Add the following content:
   ```ini
   [Unit]
   Description=Quacker Service
   After=network.target redis.service

   [Service]
   ExecStart=/path/to/quacker --run
   Restart=always
   User=www-data
   Group=www-data
   Environment=PATH=/usr/bin:/usr/local/bin
   WorkingDirectory=/path/to/

   [Install]
   WantedBy=multi-user.target
   ```
   Replace `/path/to/quacker` with the path to your Quacker binary.

### Enable and Start the Service
1. Enable the service to start on boot:
   ```bash
   sudo systemctl enable quacker
   ```

2. Start the service:
   ```bash
   sudo systemctl start quacker
   ```

3. Verify the service is running:
   ```bash
   sudo systemctl status quacker
   ```

---

## Scheduling the Background Job with Cron

To schedule the `--job` command every 5 minutes:

1. Open the crontab editor:
   ```bash
   crontab -e
   ```

2. Add the following line:
   ```
   */5 * * * * /path/to/quacker --job
   ```

Replace `/path/to/quacker` with the path to your Quacker binary.

---

You are now ready to run Quacker on your server!