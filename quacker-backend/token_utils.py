import hashlib
import hmac
import time
from datetime import datetime, timedelta

def validate_token(token, secret):
    # Logic to split and validate token structure
    try:
        token_parts = token.split('.')
        bucket_name = token_parts[0]
        expiration_time = int(token_parts[1])
        provided_hash = token_parts[2]
        
        # Validate expiration
        if datetime.now().timestamp() > expiration_time:
            return False, None

        # Verify the hash
        data = f"{bucket_name}.{expiration_time}"
        expected_hash = hmac.new(secret.encode(), data.encode(), hashlib.sha256).hexdigest()
        if hmac.compare_digest(provided_hash, expected_hash):
            return True, bucket_name
        else:
            return False, None
    except Exception as e:
        return False, None

def generate_presigned_urls(bucket_name, s3_client):
    try:
        # Example of generating URLs - replace with actual S3 bucket structure
        response = s3_client.list_objects_v2(Bucket=bucket_name)
        urls = []
        for obj in response.get('Contents', []):
            url = s3_client.generate_presigned_url(
                'get_object',
                Params={'Bucket': bucket_name, 'Key': obj['Key']},
                ExpiresIn=3600  # 1 hour
            )
            urls.append(url)
        return urls
    except Exception as e:
        return []
