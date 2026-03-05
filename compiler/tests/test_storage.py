"""Tests for S3Client URL parsing and operations."""
from unittest.mock import MagicMock, patch

import pytest

from compiler.storage import S3Client


# ---------------------------------------------------------------------------
# Helper: create an S3Client without calling __init__ (avoids boto3)
# ---------------------------------------------------------------------------

def _bare_client():
    """Create an S3Client without calling __init__ so boto3 is not needed."""
    client = S3Client.__new__(S3Client)
    return client


# ---------------------------------------------------------------------------
# Existing URL-parsing tests
# ---------------------------------------------------------------------------


def test_parse_s3_url():
    client = _bare_client()
    bucket, key = client._parse_s3_url("s3://fleetml-models/test/v1/model.onnx")
    assert bucket == "fleetml-models"
    assert key == "test/v1/model.onnx"


def test_parse_s3_url_nested():
    client = _bare_client()
    bucket, key = client._parse_s3_url("s3://my-bucket/a/b/c/d/model.trt")
    assert bucket == "my-bucket"
    assert key == "a/b/c/d/model.trt"


def test_endpoint_scheme_added():
    # Test that endpoint without scheme gets http:// prepended
    client = _bare_client()
    client.endpoint = "minio:9000"
    if not client.endpoint.startswith("http"):
        client.endpoint = f"http://{client.endpoint}"
    assert client.endpoint == "http://minio:9000"


def test_endpoint_scheme_preserved():
    client = _bare_client()
    client.endpoint = "https://s3.amazonaws.com"
    if not client.endpoint.startswith("http"):
        client.endpoint = f"http://{client.endpoint}"
    assert client.endpoint == "https://s3.amazonaws.com"


# ---------------------------------------------------------------------------
# New tests with mocked boto3
# ---------------------------------------------------------------------------


@patch("compiler.storage.boto3")
def test_download_success(mock_boto3):
    """Mock boto3 client.download_file, verify called with correct args."""
    mock_s3_client = MagicMock()
    mock_boto3.client.return_value = mock_s3_client

    client = S3Client(
        endpoint="http://localhost:9000",
        access_key="testkey",
        secret_key="testsecret",
        bucket="test-bucket",
    )

    client.download("s3://my-bucket/models/v1/model.onnx", "/tmp/model.onnx")

    mock_s3_client.download_file.assert_called_once_with(
        "my-bucket", "models/v1/model.onnx", "/tmp/model.onnx"
    )


@patch("compiler.storage.boto3")
def test_download_file_not_found(mock_boto3):
    """Mock boto3 raises ClientError for missing file, verify exception."""
    from botocore.exceptions import ClientError

    mock_s3_client = MagicMock()
    mock_boto3.client.return_value = mock_s3_client
    mock_s3_client.download_file.side_effect = ClientError(
        {"Error": {"Code": "404", "Message": "Not Found"}}, "GetObject"
    )

    client = S3Client(
        endpoint="http://localhost:9000",
        access_key="testkey",
        secret_key="testsecret",
        bucket="test-bucket",
    )

    with pytest.raises(ClientError):
        client.download("s3://my-bucket/nonexistent.onnx", "/tmp/out.onnx")


@patch("compiler.storage.boto3")
def test_upload_success(mock_boto3):
    """Mock boto3 client.upload_file, verify returns correct s3:// URL."""
    mock_s3_client = MagicMock()
    mock_boto3.client.return_value = mock_s3_client

    client = S3Client(
        endpoint="http://localhost:9000",
        access_key="testkey",
        secret_key="testsecret",
        bucket="fleetml-models",
    )

    url = client.upload("/tmp/compiled.trt", "model123/compiled/tensorrt/model.trt")

    mock_s3_client.upload_file.assert_called_once_with(
        "/tmp/compiled.trt", "fleetml-models", "model123/compiled/tensorrt/model.trt"
    )
    assert url == "s3://fleetml-models/model123/compiled/tensorrt/model.trt"


@patch("compiler.storage.boto3")
def test_upload_failure(mock_boto3):
    """Mock upload raises exception, verify propagated."""
    mock_s3_client = MagicMock()
    mock_boto3.client.return_value = mock_s3_client
    mock_s3_client.upload_file.side_effect = Exception("Connection refused")

    client = S3Client(
        endpoint="http://localhost:9000",
        access_key="testkey",
        secret_key="testsecret",
        bucket="test-bucket",
    )

    with pytest.raises(Exception, match="Connection refused"):
        client.upload("/tmp/model.trt", "key/model.trt")


def test_parse_s3_url_no_key():
    """URL like 's3://bucket' with no key produces empty key string."""
    client = _bare_client()
    bucket, key = client._parse_s3_url("s3://bucket")
    assert bucket == "bucket"
    assert key == ""


def test_parse_s3_url_empty():
    """Empty string should parse but yield empty bucket and key."""
    client = _bare_client()
    bucket, key = client._parse_s3_url("")
    # urlparse("") gives empty netloc and path
    assert bucket == ""
    assert key == ""


def test_parse_s3_url_wrong_scheme():
    """Non-s3 scheme like http:// still parses via urlparse (no scheme enforcement)."""
    client = _bare_client()
    # urlparse handles any scheme; _parse_s3_url does not enforce s3://
    bucket, key = client._parse_s3_url("http://bucket/key/model.onnx")
    # With http:// scheme, urlparse treats 'bucket' as netloc
    assert bucket == "bucket"
    assert key == "key/model.onnx"


@patch("compiler.storage.boto3")
def test_init_endpoint_variations_no_scheme(mock_boto3):
    """Endpoint without http:// prefix gets http:// prepended."""
    mock_boto3.client.return_value = MagicMock()

    client = S3Client(
        endpoint="minio:9000",
        access_key="key",
        secret_key="secret",
        bucket="b",
    )
    assert client.endpoint == "http://minio:9000"


@patch("compiler.storage.boto3")
def test_init_endpoint_variations_with_http(mock_boto3):
    """Endpoint already with http:// is preserved."""
    mock_boto3.client.return_value = MagicMock()

    client = S3Client(
        endpoint="http://minio:9000",
        access_key="key",
        secret_key="secret",
        bucket="b",
    )
    assert client.endpoint == "http://minio:9000"


@patch("compiler.storage.boto3")
def test_init_endpoint_variations_with_https(mock_boto3):
    """Endpoint with https:// is preserved."""
    mock_boto3.client.return_value = MagicMock()

    client = S3Client(
        endpoint="https://s3.amazonaws.com",
        access_key="key",
        secret_key="secret",
        bucket="b",
    )
    assert client.endpoint == "https://s3.amazonaws.com"


# ---------------------------------------------------------------------------
# Edge-case tests
# ---------------------------------------------------------------------------


def test_parse_s3_url_trailing_slash():
    """URL with trailing slash — key should include trailing slash."""
    client = _bare_client()
    bucket, key = client._parse_s3_url("s3://bucket/path/")
    assert bucket == "bucket"
    assert key == "path/"


def test_parse_s3_url_special_characters():
    """URL with spaces and special chars in key."""
    client = _bare_client()
    bucket, key = client._parse_s3_url("s3://bucket/path/my model (v2).onnx")
    assert bucket == "bucket"
    assert "my model (v2).onnx" in key


def test_parse_s3_url_unicode():
    """URL with unicode characters in key."""
    client = _bare_client()
    bucket, key = client._parse_s3_url("s3://bucket/models/模型.onnx")
    assert bucket == "bucket"
    assert "模型.onnx" in key


def test_parse_s3_url_double_slash():
    """URL with double slashes in path."""
    client = _bare_client()
    bucket, key = client._parse_s3_url("s3://bucket//double//slash//model.onnx")
    assert bucket == "bucket"
    # urlparse preserves double slashes in path
    assert "model.onnx" in key


@patch("compiler.storage.boto3")
def test_upload_empty_key(mock_boto3):
    """Upload with empty string key should still call upload_file."""
    mock_s3_client = MagicMock()
    mock_boto3.client.return_value = mock_s3_client

    client = S3Client(
        endpoint="http://localhost:9000",
        access_key="key",
        secret_key="secret",
        bucket="test-bucket",
    )

    url = client.upload("/tmp/model.onnx", "")
    mock_s3_client.upload_file.assert_called_once_with("/tmp/model.onnx", "test-bucket", "")
    assert url == "s3://test-bucket/"


@patch("compiler.storage.boto3")
def test_download_called_with_parsed_url(mock_boto3):
    """Verify download parses the s3:// URL and passes bucket + key correctly."""
    mock_s3_client = MagicMock()
    mock_boto3.client.return_value = mock_s3_client

    client = S3Client(
        endpoint="http://localhost:9000",
        access_key="key",
        secret_key="secret",
        bucket="default-bucket",
    )

    client.download("s3://other-bucket/deep/nested/path/model.onnx", "/tmp/out.onnx")
    mock_s3_client.download_file.assert_called_once_with(
        "other-bucket", "deep/nested/path/model.onnx", "/tmp/out.onnx"
    )


@patch("compiler.storage.boto3")
def test_init_stores_bucket(mock_boto3):
    """Verify the bucket attribute is stored correctly."""
    mock_boto3.client.return_value = MagicMock()

    client = S3Client(
        endpoint="http://localhost:9000",
        access_key="key",
        secret_key="secret",
        bucket="my-special-bucket",
    )
    assert client.bucket == "my-special-bucket"


@patch("compiler.storage.boto3")
def test_init_empty_credentials(mock_boto3):
    """Empty credentials should still create a client (no validation)."""
    mock_boto3.client.return_value = MagicMock()

    client = S3Client(
        endpoint="http://localhost:9000",
        access_key="",
        secret_key="",
        bucket="bucket",
    )
    assert client.bucket == "bucket"
