import os
import json
import boto3
import time
from flask import Flask, request, jsonify
from dotenv import load_dotenv
from itsdangerous import TimedJSONWebSignatureSerializer as Serializer, BadSignature, SignatureExpired
from collections import defaultdict

# Load environment variables
load_dotenv()

AWS_URL = os.getenv("AWS_URL")
AWS_KEY = os.getenv("AWS_KEY")
AWS_SECRET = os.getenv("AWS_SECRET")
TOKEN_SECRET = os.getenv("TOKEN_SECRET")

# Initialize Flask app
app = Flask(__name__)
app.config['SECRET_KEY'] = TOKEN_SECRET

# Initialize S3 client
s3_client = boto3.client('s3', endpoint_url=AWS_URL, aws_access_key_id=AWS_KEY, aws_secret_access_key=AWS_SECRET)

# Flood control tracking
request_counts = defaultdict(lambda: {'count': 0, 'last_request_time': 0})
REQUEST_LIMIT = 5  # Max number of requests allowed
TIME_WINDOW = 60  # Time window in seconds

# Validate Token Endpoint
@app.route('/validate_token', methods=['POST'])
def validate_token():
    client_ip = request.remote_addr
    current_time = time.time()

    # Flood control logic
    if client_ip in request_counts:
        elapsed_time = current_time - request_counts[client_ip]['last_request_time']
        if elapsed_time > TIME_WINDOW:
            request_counts[client_ip] = {'count': 1, 'last_request_time': current_time}
        else:
            if request_counts[client_ip]['count'] >= REQUEST_LIMIT:
                return jsonify({"error": "Too many requests. Please try again later."}), 429
            request_counts[client_ip]['count'] += 1
            request_counts[client_ip]['last_request_time'] = current_time
    else:
        request_counts[client_ip] = {'count': 1, 'last_request_time': current_time}

    data = request.get_json()
    token = data.get('token')

    if not token:
        return jsonify({"error": "Token is required"}), 400

    s = Serializer(app.config['SECRET_KEY'])
    try:
        # Validate token
        token_data = s.loads(token)
    except SignatureExpired:
        return jsonify({"error": "Token has expired"}), 400
    except BadSignature:
        return jsonify({"error": "Invalid token"}), 400

    bucket_name = token_data.get('bucket_name')
    expiration = token_data.get('exp')

    # List objects in the bucket
    try:
        response = s3_client.list_objects_v2(Bucket=bucket_name)
        if 'Contents' not in response:
            return jsonify({"error": "No files found in the bucket"}), 404

        # Generate presigned URLs
        presigned_urls = []
        for obj in response['Contents']:
            url = s3_client.generate_presigned_url('get_object',
                                                  Params={'Bucket': bucket_name, 'Key': obj['Key']},
                                                  ExpiresIn=expiration)
            presigned_urls.append(url)

        return jsonify({"urls": presigned_urls}), 200
    except s3_client.exceptions.NoSuchBucket:
        return jsonify({"error": "Bucket does not exist"}), 400
    except Exception as e:
        return jsonify({"error": str(e)}), 500

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000)
