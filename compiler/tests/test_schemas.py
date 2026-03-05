"""Tests for compiler schemas."""
import pytest
from pydantic import ValidationError

from compiler.schemas import CompileRequest, CompileResponse, RuntimeInfo


# ---------------------------------------------------------------------------
# Existing tests
# ---------------------------------------------------------------------------


def test_compile_request_defaults():
    req = CompileRequest(
        model_url="s3://bucket/model.onnx",
        model_id="test-id",
        target_runtime="mock",
    )
    assert req.options == {}
    assert req.target_runtime == "mock"
    assert req.model_id == "test-id"


def test_compile_request_with_options():
    req = CompileRequest(
        model_url="s3://bucket/model.onnx",
        model_id="test-id",
        target_runtime="tensorrt",
        options={"fp16": True, "workspace_mb": 1024},
    )
    assert req.options["fp16"] is True
    assert req.options["workspace_mb"] == 1024


def test_compile_response():
    resp = CompileResponse(
        runtime="mock",
        artifact_url="s3://bucket/compiled/model.onnx",
        checksum="sha256:abc123",
        file_size=1024,
        compile_time_seconds=0.5,
    )
    assert resp.metadata == {}
    assert resp.file_size == 1024


def test_compile_response_with_metadata():
    resp = CompileResponse(
        runtime="tensorrt",
        artifact_url="s3://bucket/compiled/model.trt",
        checksum="sha256:def456",
        file_size=2048,
        compile_time_seconds=30.5,
        metadata={"trt_version": "8.6", "fp16": True},
    )
    assert resp.metadata["trt_version"] == "8.6"


def test_runtime_info():
    info = RuntimeInfo(
        name="mock",
        status="available",
        description="Mock compiler for testing",
    )
    assert info.name == "mock"
    assert info.status == "available"


# ---------------------------------------------------------------------------
# New tests
# ---------------------------------------------------------------------------


def test_compile_request_missing_required():
    """model_url missing should raise ValidationError."""
    with pytest.raises(ValidationError):
        CompileRequest(
            model_id="test-id",
            target_runtime="mock",
        )

    # model_id missing
    with pytest.raises(ValidationError):
        CompileRequest(
            model_url="s3://bucket/model.onnx",
            target_runtime="mock",
        )

    # target_runtime missing
    with pytest.raises(ValidationError):
        CompileRequest(
            model_url="s3://bucket/model.onnx",
            model_id="test-id",
        )


def test_compile_request_empty_strings():
    """model_url="" -- Pydantic accepts empty strings (no min_length constraint)."""
    req = CompileRequest(
        model_url="",
        model_id="",
        target_runtime="",
    )
    assert req.model_url == ""
    assert req.model_id == ""
    assert req.target_runtime == ""


def test_compile_response_negative_file_size():
    """Negative int accepted -- no constraint defined in the schema."""
    resp = CompileResponse(
        runtime="mock",
        artifact_url="s3://bucket/model.onnx",
        checksum="sha256:abc",
        file_size=-1,
        compile_time_seconds=1.0,
    )
    assert resp.file_size == -1


def test_compile_response_zero_compile_time():
    """Zero compile_time_seconds is valid."""
    resp = CompileResponse(
        runtime="mock",
        artifact_url="s3://bucket/model.onnx",
        checksum="sha256:abc",
        file_size=0,
        compile_time_seconds=0.0,
    )
    assert resp.compile_time_seconds == 0.0


def test_runtime_info_all_statuses():
    """Both 'available' and 'unavailable' statuses are accepted."""
    avail = RuntimeInfo(
        name="mock",
        status="available",
        description="Mock compiler",
    )
    assert avail.status == "available"

    unavail = RuntimeInfo(
        name="tensorrt",
        status="unavailable",
        description="NVIDIA TensorRT",
    )
    assert unavail.status == "unavailable"


# ---------------------------------------------------------------------------
# Edge-case tests
# ---------------------------------------------------------------------------


def test_compile_request_none_options():
    """Explicitly passing None for options should raise ValidationError (dict expected)."""
    with pytest.raises(ValidationError):
        CompileRequest(
            model_url="s3://bucket/model.onnx",
            model_id="test-id",
            target_runtime="mock",
            options=None,
        )


def test_compile_request_extra_fields_ignored():
    """Extra fields not in the model should be ignored by Pydantic."""
    req = CompileRequest(
        model_url="s3://bucket/model.onnx",
        model_id="test-id",
        target_runtime="mock",
        unknown_field="should be ignored",
    )
    assert not hasattr(req, "unknown_field")


def test_compile_request_unicode_values():
    """Unicode strings should be accepted."""
    req = CompileRequest(
        model_url="s3://bucket/模型.onnx",
        model_id="テスト",
        target_runtime="mock",
    )
    assert req.model_url == "s3://bucket/模型.onnx"
    assert req.model_id == "テスト"


def test_compile_response_very_large_file_size():
    """Large file size (multi-GB) should be accepted."""
    resp = CompileResponse(
        runtime="mock",
        artifact_url="s3://bucket/model.onnx",
        checksum="sha256:abc",
        file_size=10 * 1024 * 1024 * 1024,  # 10GB
        compile_time_seconds=3600.0,
    )
    assert resp.file_size == 10 * 1024 * 1024 * 1024


def test_compile_response_float_file_size_rejected():
    """Pydantic v2 rejects floats with fractional parts for int fields."""
    with pytest.raises(ValidationError):
        CompileResponse(
            runtime="mock",
            artifact_url="s3://bucket/model.onnx",
            checksum="sha256:abc",
            file_size=1024.7,
            compile_time_seconds=1.0,
        )


def test_compile_request_json_roundtrip():
    """Serialize to JSON and back, verify equality."""
    req = CompileRequest(
        model_url="s3://bucket/model.onnx",
        model_id="test-id",
        target_runtime="tensorrt",
        options={"fp16": True, "workspace_mb": 2048},
    )
    json_str = req.model_dump_json()
    restored = CompileRequest.model_validate_json(json_str)
    assert restored.model_url == req.model_url
    assert restored.model_id == req.model_id
    assert restored.target_runtime == req.target_runtime
    assert restored.options == req.options


def test_compile_response_json_roundtrip():
    """Serialize CompileResponse to JSON and back."""
    resp = CompileResponse(
        runtime="openvino",
        artifact_url="s3://bucket/compiled/model.xml",
        checksum="sha256:deadbeef",
        file_size=4096,
        compile_time_seconds=12.34,
        metadata={"openvino_version": "2023.2"},
    )
    json_str = resp.model_dump_json()
    restored = CompileResponse.model_validate_json(json_str)
    assert restored.runtime == resp.runtime
    assert restored.checksum == resp.checksum
    assert restored.metadata == resp.metadata


def test_runtime_info_empty_description():
    """Empty description should be accepted."""
    info = RuntimeInfo(name="test", status="available", description="")
    assert info.description == ""


def test_runtime_info_arbitrary_status():
    """No enum constraint — any status string is accepted."""
    info = RuntimeInfo(name="test", status="degraded", description="Some compiler")
    assert info.status == "degraded"
