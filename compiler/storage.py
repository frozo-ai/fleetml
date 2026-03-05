import os
from typing import Optional, Tuple
from urllib.parse import urlparse

import boto3
from botocore.config import Config as BotoConfig


class S3Client:
    """S3-compatible storage client for downloading source models and uploading compiled variants."""

    def __init__(
        self,
        endpoint: Optional[str] = None,
        access_key: Optional[str] = None,
        secret_key: Optional[str] = None,
        bucket: Optional[str] = None,
    ):
        self.endpoint = endpoint or os.getenv("S3_ENDPOINT", "http://localhost:9000")
        self.access_key = access_key or os.getenv("S3_ACCESS_KEY", "minioadmin")
        self.secret_key = secret_key or os.getenv("S3_SECRET_KEY", "minioadmin")
        self.bucket = bucket or os.getenv("S3_BUCKET", "fleetml-models")

        # Ensure endpoint has scheme
        if not self.endpoint.startswith("http"):
            self.endpoint = f"http://{self.endpoint}"

        self._client = boto3.client(
            "s3",
            endpoint_url=self.endpoint,
            aws_access_key_id=self.access_key,
            aws_secret_access_key=self.secret_key,
            config=BotoConfig(signature_version="s3v4"),
            region_name="us-east-1",
        )

    def _parse_s3_url(self, s3_url: str) -> Tuple[str, str]:
        """Parse s3://bucket/key into (bucket, key)."""
        parsed = urlparse(s3_url)
        bucket = parsed.netloc
        key = parsed.path.lstrip("/")
        return bucket, key

    def download(self, s3_url: str, local_path: str) -> None:
        """Download a file from S3 to a local path."""
        bucket, key = self._parse_s3_url(s3_url)
        self._client.download_file(bucket, key, local_path)

    def upload(self, local_path: str, s3_key: str) -> str:
        """Upload a local file to S3. Returns the s3:// URL."""
        self._client.upload_file(local_path, self.bucket, s3_key)
        return f"s3://{self.bucket}/{s3_key}"
