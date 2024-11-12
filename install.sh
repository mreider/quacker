#!/bin/bash

# Prompt for necessary inputs
echo "Enter the domain name for your backend (e.g., example.com):"
read DOMAIN_NAME
echo "Enter the email address for Let's Encrypt SSL certificate registration:"
read EMAIL_ADDRESS
echo "Enter the AWS URL for S3-compatible object storage:"
read AWS_URL
echo "Enter the AWS Access Key:"
read AWS_KEY
echo "Enter the AWS Secret Key:"
read AWS_SECRET
echo "Enter a secret key for generating tokens (15 characters):"
read TOKEN_SECRET

# Update and install dependencies
sudo apt update && sudo apt upgrade -y
sudo apt install -y python3 python3-pip nginx software-properties-common certbot python3-certbot-nginx

# Install Python dependencies
pip3 install flask gunicorn boto3 python-dotenv

# Ensure user and group exist
sudo useradd -m -s /bin/bash ubuntu || true
sudo groupadd www-data || true

# Create environment file
cat <<EOT >> /home/ubuntu/backend/.env
AWS_URL=$AWS_URL
AWS_KEY=$AWS_KEY
AWS_SECRET=$AWS_SECRET
TOKEN_SECRET=$TOKEN_SECRET
EOT

# Create systemd service file for Gunicorn
sudo tee /etc/systemd/system/backend.service > /dev/null <<EOT
[Unit]
Description=Gunicorn instance to serve backend
After=network.target

[Service]
User=ubuntu
Group=www-data
WorkingDirectory=/home/ubuntu/backend
ExecStart=/usr/bin/gunicorn --workers 3 --bind 0.0.0.0:8000 app:app

[Install]
WantedBy=multi-user.target
EOT

# Start and enable the backend service
sudo systemctl daemon-reload
sudo systemctl start backend
sudo systemctl enable backend

# Configure Nginx
echo "Configuring Nginx..."
sudo tee /etc/nginx/sites-available/$DOMAIN_NAME > /dev/null <<EOT
server {
    listen 80;
    server_name $DOMAIN_NAME;

    location / {
        proxy_pass http://127.0.0.1:8000;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
}
EOT

# Enable Nginx configuration
sudo ln -s /etc/nginx/sites-available/$DOMAIN_NAME /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl restart nginx

# Set up Let's Encrypt SSL certificate
sudo certbot --nginx -d $DOMAIN_NAME --email $EMAIL_ADDRESS --agree-tos --redirect

# Enable auto-renewal of SSL certificate
sudo systemctl enable certbot.timer

# Completion message
echo "Backend setup complete. Your service is now running and configured with Nginx and SSL."
