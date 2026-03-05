"""Tests for the compiler FastAPI application."""
import os
import tempfile
from unittest.mock import MagicMock, patch

import pytest
from fastapi.testclient import TestClient

from compiler.main import app, _runtime_extension
from compiler.compilers.base import CompileResult


@pytest.fixture
def client():
    return TestClient(app)


# ---------------------------------------------------------------------------
# Existing tests
# ---------------------------------------------------------------------------


def test_health(client):
    resp = client.get("/health")
    assert resp.status_code == 200
    data = resp.json()
    assert data["status"] == "healthy"
    assert "available_runtimes" in data
    assert isinstance(data["available_runtimes"], list)


def test_supported_runtimes(client):
    resp = client.get("/supported-runtimes")
    assert resp.status_code == 200
    data = resp.json()
    assert "runtimes" in data
    runtimes = data["runtimes"]
    assert len(runtimes) >= 1
    # Mock should be available
    names = [r["name"] for r in runtimes]
    assert "mock" in names
    mock_rt = next(r for r in runtimes if r["name"] == "mock")
    assert mock_rt["status"] == "available"


def test_compile_missing_fields(client):
    resp = client.post("/compile", json={})
    assert resp.status_code == 422  # Pydantic validation error


def test_compile_unsupported_runtime(client):
    resp = client.post("/compile", json={
        "model_url": "s3://bucket/model.onnx",
        "model_id": "test-id",
        "target_runtime": "nonexistent_runtime",
    })
    assert resp.status_code == 400
    assert "Unsupported runtime" in resp.json()["detail"]


def test_compile_bad_s3_url(client):
    resp = client.post("/compile", json={
        "model_url": "s3://nonexistent-bucket/nonexistent-key",
        "model_id": "test-id",
        "target_runtime": "mock",
    })
    # Should fail because S3 download will fail
    assert resp.status_code in (400, 500)


# ---------------------------------------------------------------------------
# New tests with mocked S3 and compiler
# ---------------------------------------------------------------------------


def test_compile_success_mock_runtime(client):
    """Mock S3 download/upload and compiler, verify 200 + CompileResponse structure."""
    fake_result = CompileResult(
        output_path="/tmp/fake_output.onnx",
        runtime="mock",
        file_size=512,
        metadata={"compiler": "mock", "source_format": "onnx"},
    )

    mock_compiler = MagicMock()
    mock_compiler.validate.return_value = True
    mock_compiler.compile.return_value = fake_result

    with patch("compiler.main.registry") as mock_registry, \
         patch("compiler.main.s3") as mock_s3, \
         patch("compiler.main._sha256", return_value="abc123def456"):
        mock_registry.get.return_value = mock_compiler
        # download writes a fake file so compile can work
        def fake_download(url, local_path):
            with open(local_path, "wb") as f:
                f.write(b"fake onnx model data")
        mock_s3.download.side_effect = fake_download
        mock_s3.upload.return_value = "s3://fleetml-models/test-model/compiled/mock/model.onnx"

        # Make compile side-effect create the output file so _sha256 can read it
        def fake_compile(input_path, output_path, options):
            with open(output_path, "wb") as f:
                f.write(b"compiled model data")
            return fake_result

        mock_compiler.compile.side_effect = fake_compile

        resp = client.post("/compile", json={
            "model_url": "s3://bucket/model.onnx",
            "model_id": "test-model",
            "target_runtime": "mock",
            "options": {},
        })

    assert resp.status_code == 200
    data = resp.json()
    assert data["runtime"] == "mock"
    assert "artifact_url" in data
    assert data["artifact_url"].startswith("s3://")
    assert "checksum" in data
    assert data["checksum"].startswith("sha256:")
    assert data["file_size"] == 512
    assert "compile_time_seconds" in data
    assert isinstance(data["compile_time_seconds"], float)
    assert "metadata" in data


def test_compile_onnx_validation_failure(client):
    """Mock compiler.validate() returning False, verify 400."""
    mock_compiler = MagicMock()
    mock_compiler.validate.return_value = False

    with patch("compiler.main.registry") as mock_registry, \
         patch("compiler.main.s3") as mock_s3:
        mock_registry.get.return_value = mock_compiler
        mock_s3.download.side_effect = lambda url, path: open(path, "wb").close()

        resp = client.post("/compile", json={
            "model_url": "s3://bucket/model.onnx",
            "model_id": "test-id",
            "target_runtime": "mock",
        })

    assert resp.status_code == 400
    assert "Invalid ONNX model" in resp.json()["detail"]


def test_compile_compilation_error(client):
    """Mock compiler.compile() raising exception, verify 500."""
    mock_compiler = MagicMock()
    mock_compiler.validate.return_value = True
    mock_compiler.compile.side_effect = RuntimeError("GPU out of memory")

    with patch("compiler.main.registry") as mock_registry, \
         patch("compiler.main.s3") as mock_s3:
        mock_registry.get.return_value = mock_compiler
        mock_s3.download.side_effect = lambda url, path: open(path, "wb").close()

        resp = client.post("/compile", json={
            "model_url": "s3://bucket/model.onnx",
            "model_id": "test-id",
            "target_runtime": "mock",
        })

    assert resp.status_code == 500
    assert "Compilation failed" in resp.json()["detail"]
    assert "GPU out of memory" in resp.json()["detail"]


def test_compile_s3_download_failure(client):
    """Mock S3Client.download() raising, verify 400."""
    mock_compiler = MagicMock()

    with patch("compiler.main.registry") as mock_registry, \
         patch("compiler.main.s3") as mock_s3:
        mock_registry.get.return_value = mock_compiler
        mock_s3.download.side_effect = Exception("NoSuchBucket")

        resp = client.post("/compile", json={
            "model_url": "s3://missing-bucket/model.onnx",
            "model_id": "test-id",
            "target_runtime": "mock",
        })

    assert resp.status_code == 400
    assert "Failed to download model" in resp.json()["detail"]


def test_compile_s3_upload_failure(client):
    """Mock S3Client.upload() raising, verify 500."""
    fake_result = CompileResult(
        output_path="/tmp/out.onnx",
        runtime="mock",
        file_size=100,
        metadata={},
    )
    mock_compiler = MagicMock()
    mock_compiler.validate.return_value = True

    with patch("compiler.main.registry") as mock_registry, \
         patch("compiler.main.s3") as mock_s3, \
         patch("compiler.main._sha256", return_value="deadbeef"):
        mock_registry.get.return_value = mock_compiler

        def fake_download(url, local_path):
            with open(local_path, "wb") as f:
                f.write(b"fake model")
        mock_s3.download.side_effect = fake_download

        def fake_compile(input_path, output_path, options):
            with open(output_path, "wb") as f:
                f.write(b"compiled")
            return fake_result
        mock_compiler.compile.side_effect = fake_compile

        mock_s3.upload.side_effect = Exception("S3 connection timeout")

        resp = client.post("/compile", json={
            "model_url": "s3://bucket/model.onnx",
            "model_id": "test-id",
            "target_runtime": "mock",
        })

    assert resp.status_code == 500
    assert "Failed to upload compiled artifact" in resp.json()["detail"]


def test_health_endpoint(client):
    """GET /health returns status and available_runtimes."""
    resp = client.get("/health")
    assert resp.status_code == 200
    data = resp.json()
    assert data["status"] == "healthy"
    assert data["version"] == "0.2.0"
    assert "available_runtimes" in data
    assert isinstance(data["available_runtimes"], list)
    # mock should always be in available_runtimes
    assert "mock" in data["available_runtimes"]


def test_supported_runtimes_endpoint(client):
    """GET /supported-runtimes returns structured runtime info."""
    resp = client.get("/supported-runtimes")
    assert resp.status_code == 200
    data = resp.json()
    assert "runtimes" in data
    runtimes = data["runtimes"]

    # Verify structure of each runtime entry
    for rt in runtimes:
        assert "name" in rt
        assert "status" in rt
        assert "description" in rt
        assert rt["status"] in ("available", "unavailable")

    # Verify all four known runtimes are listed
    names = [r["name"] for r in runtimes]
    assert "mock" in names
    assert "tensorrt" in names
    assert "openvino" in names
    assert "tflite" in names
    assert len(runtimes) == 4


def test_runtime_extension_mapping():
    """Test _runtime_extension() for all known runtimes and unknown fallback."""
    assert _runtime_extension("tensorrt") == "trt"
    assert _runtime_extension("openvino") == "xml"
    assert _runtime_extension("tflite") == "tflite"
    assert _runtime_extension("mock") == "onnx"
    assert _runtime_extension("unknown") == "bin"
    assert _runtime_extension("") == "bin"
    assert _runtime_extension("some_future_runtime") == "bin"


# ---------------------------------------------------------------------------
# Edge-case tests
# ---------------------------------------------------------------------------


def test_compile_empty_body(client):
    """POST /compile with empty string body should return 422."""
    resp = client.post("/compile", content="", headers={"Content-Type": "application/json"})
    assert resp.status_code == 422


def test_compile_null_body(client):
    """POST /compile with JSON null body should return 422."""
    resp = client.post("/compile", json=None)
    assert resp.status_code == 422


def test_compile_array_body(client):
    """POST /compile with JSON array instead of object should return 422."""
    resp = client.post("/compile", json=[{"model_url": "s3://b/m.onnx"}])
    assert resp.status_code == 422


def test_compile_extra_fields_accepted(client):
    """Extra fields in request body should be ignored (not cause errors)."""
    resp = client.post("/compile", json={
        "model_url": "s3://bucket/model.onnx",
        "model_id": "test-id",
        "target_runtime": "nonexistent_runtime",
        "extra_field": "should be ignored",
    })
    # Should fail on runtime, not on extra field
    assert resp.status_code == 400
    assert "Unsupported runtime" in resp.json()["detail"]


def test_health_response_format(client):
    """Verify /health response contains all expected keys."""
    resp = client.get("/health")
    data = resp.json()
    assert "status" in data
    assert "version" in data
    assert "available_runtimes" in data
    assert isinstance(data["available_runtimes"], list)


def test_health_method_not_allowed(client):
    """POST to /health should return 405."""
    resp = client.post("/health", json={})
    assert resp.status_code == 405


def test_supported_runtimes_method_not_allowed(client):
    """POST to /supported-runtimes should return 405."""
    resp = client.post("/supported-runtimes", json={})
    assert resp.status_code == 405


def test_compile_get_not_allowed(client):
    """GET /compile should return 405."""
    resp = client.get("/compile")
    assert resp.status_code == 405


def test_nonexistent_endpoint(client):
    """GET /nonexistent should return 404."""
    resp = client.get("/nonexistent")
    assert resp.status_code == 404


def test_compile_partial_fields(client):
    """POST /compile with only model_url should return 422 (missing required fields)."""
    resp = client.post("/compile", json={"model_url": "s3://bucket/model.onnx"})
    assert resp.status_code == 422
