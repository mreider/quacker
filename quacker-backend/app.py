from flask import Flask, request, jsonify
from datetime import datetime, timedelta
import os
import hashlib
import boto3
import json
from token_utils import validate_token, generate_presigned_urls

app = Flask(__name__)

# Load environment variables
AWS_URL = os.getenv("AWS_URL")
AWS_KEY = os.getenv("AWS_KEY")
AWS_SECRET = os.getenv("AWS_SECRET")
TOKEN_SECRET = os.getenv("TOKEN_SECRET")

# Initialize boto3 client for S3
s3_client = boto3.client(
    's3',
    endpoint_url=AWS_URL,
    aws_access_key_id=AWS_KEY,
    aws_secret_access_key=AWS_SECRET
)

# Flood control dictionary
request_count = {}

@app.route('/validate_token', methods=['POST'])
def validate_token_route():
    ip = request.remote_addr
    token = request.json.get("token")

    # Implement simple flood control
    if ip in request_count:
        if request_count[ip]['count'] > 5 and (datetime.now() - request_count[ip]['time']).seconds < 60:
            return jsonify({"error": "Too many attempts, please try again later"}), 429
        elif (datetime.now() - request_count[ip]['time']).seconds > 60:
            request_count[ip] = {'count': 1, 'time': datetime.now()}
        else:
            request_count[ip]['count'] += 1
    else:
        request_count[ip] = {'count': 1, 'time': datetime.now()}

    # Validate token
    is_valid, bucket_name = validate_token(token, TOKEN_SECRET)
    if not is_valid:
        return jsonify({"error": "Invalid token"}), 400

    # Generate pre-signed URLs
    urls = generate_presigned_urls(bucket_name, s3_client)
    return jsonify({"urls": urls}), 200

if __name__ == '__main__':
    app.run()
