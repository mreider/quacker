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
sudo apt install -y python3 python3-pip python3-venv nginx software-properties-common certbot python3-certbot-nginx git

# Create a new user named quacker and set up the backend directory
sudo useradd -m -s /bin/bash quacker || true
sudo -u quacker mkdir -p /home/quacker/backend

# Clone the GitHub repository into the backend directory as quacker user
sudo -u quacker git clone https://github.com/mreider/quacker.git /home/quacker/backend

# Set up a virtual environment in the backend directory and install dependencies
sudo -u quacker python3 -m venv /home/quacker/backend/venv
sudo -u quacker /home/quacker/backend/venv/bin/pip install -r /home/quacker/backend/requirements.txt

# Create environment file
sudo tee /home/quacker/backend/.env > /dev/null <<EOT
AWS_URL=$AWS_URL
AWS_KEY=$AWS_KEY
AWS_SECRET=$AWS_SECRET
TOKEN_SECRET=$TOKEN_SECRET
EOT

# Create systemd service file for Gunicorn with venv
sudo tee /etc/systemd/system/backend.service > /dev/null <<EOT
[Unit]
Description=Gunicorn instance to serve backend
After=network.target

[Service]
User=quacker
Group=www-data
WorkingDirectory=/home/quacker/backend
EnvironmentFile=/home/quacker/backend/.env
ExecStart=/home/quacker/backend/venv/bin/gunicorn --workers 3 --bind 0.0.0.0:8000 app:app
StandardOutput=append:/var/log/gunicorn/gunicorn.out.log
StandardError=append:/var/log/gunicorn/gunicorn.err.log

[Install]
WantedBy=multi-user.target
EOT

# Create log directory for Gunicorn
sudo mkdir -p /var/log/gunicorn
sudo chown quacker:www-data /var/log/gunicorn

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
