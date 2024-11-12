#!/bin/bash
sudo systemctl daemon-reload
# Restart Gunicorn backend service
sudo systemctl restart backend
# Restart Nginx
sudo systemctl restart nginx
# Completion message
echo "Backend and Nginx services have been restarted."
