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

---

## Downloading the Binary

You can download the Quacker binary directly from the [GitHub Releases page](https://github.com/your-repo-name/releases):

1. Go to the [Releases](https://github.com/your-repo-name/releases) section of the repository.
2. Download the appropriate binary for your operating system and architecture:
   - `quacker-linux-amd64` for Linux 64-bit systems
   - `quacker-linux-arm64` for Linux ARM systems
   - `quacker-darwin-amd64` for macOS 64-bit systems
   - `quacker-darwin-arm64` for macOS ARM systems
   - `quacker-windows-amd64.exe` for Windows 64-bit systems
3. Place the binary in a directory included in your system's `PATH` or execute it directly.

Make sure the binary has execution permissions:
```bash
chmod +x ./quacker-linux-amd64
```

---

## Running Quacker

### Setup Configuration
Before running any other commands, set up the Quacker application:
```bash
./quacker --setup
```
This command will prompt you to:
- Enter your Mailgun API key.
- Enter the hostname for your server (e.g., `example.com`).

Note: you might have to run via sudo in order to create certificates.


### Generate Invitation Codes
To create an invitation code:
```bash
./quacker --generate
```
This generates a code and stores it in Redis with a 48-hour expiration.

### Start the Quacker Server
To start the Quacker application:
```bash
./quacker --run
```
The server will start on port 443 using HTTPS.

Note: you might have to do this to allow 443 binding or run via sudo

```bash
 sudo setcap 'cap_net_bind_service=+ep' ./quacker
 ```


### Process Scheduled Jobs
To run the Quacker background job (e.g., for sending emails):
```bash
./quacker --job
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

You are now ready to manage Quacker on your server!