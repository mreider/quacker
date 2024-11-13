#!/bin/bash

echo "Starting Quacker backend setup..."

# Ask for user input
read -p "Enter domain name: " DOMAIN
read -p "Enter email for SSL certificate: " EMAIL
read -p "Enter AWS URL: " AWS_URL
read -p "Enter AWS Access Key: " AWS_KEY
read -p "Enter AWS Secret Key: " AWS_SECRET
read -p "Enter Token Secret (15 characters): " TOKEN_SECRET

# Create quacker user
sudo useradd -m quacker -s /bin/bash

# Set up Python virtual environment
sudo -u quacker bash << EOF
cd /home/quacker
python3 -m venv venv
source venv/bin/activate
pip install -r /home/quacker/quacker-backend/requirements.txt
EOF

# Configure environment variables
sudo tee /home/quacker/quacker-backend/.env <<EOF
AWS_URL=$AWS_URL
AWS_KEY=$AWS_KEY
AWS_SECRET=$AWS_SECRET
TOKEN_SECRET=$TOKEN_SECRET
EOF

# Install Nginx and obtain SSL certificate
sudo apt update
sudo apt install -y nginx
sudo apt install -y python3-certbot-nginx
sudo certbot --nginx -d "$DOMAIN" --email "$EMAIL" --agree-tos --non-interactive

# Configure systemd service
sudo tee /etc/systemd/system/quacker.service <<EOF
[Unit]
Description=Quacker Backend Service
After=network.target

[Service]
User=quacker
WorkingDirectory=/home/quacker/quacker-backend
EnvironmentFile=/home/quacker/quacker-backend/.env
ExecStart=/home/quacker/quacker-backend/venv/bin/gunicorn -w 4 -b localhost:8000 app:app
Restart=always

[Install]
WantedBy=multi-user.target
EOF

# Reload and start the service
sudo systemctl daemon-reload
sudo systemctl start quacker
sudo systemctl enable quacker

echo "Quacker backend setup complete."
