"""Tests for base compiler classes."""
from compiler.compilers.base import CompileResult


def test_compile_result_defaults():
    result = CompileResult(
        output_path="/tmp/model.onnx",
        runtime="mock",
        file_size=1024,
    )
    assert result.metadata == {}
    assert result.file_size == 1024
    assert result.runtime == "mock"


def test_compile_result_with_metadata():
    result = CompileResult(
        output_path="/tmp/model.trt",
        runtime="tensorrt",
        file_size=2048,
        metadata={"fp16": True, "version": "8.6"},
    )
    assert result.metadata["fp16"] is True
    assert result.output_path == "/tmp/model.trt"


# ---------------------------------------------------------------------------
# Edge-case tests
# ---------------------------------------------------------------------------


def test_compile_result_zero_file_size():
    """Zero file size is valid (empty compiled model)."""
    result = CompileResult(output_path="/tmp/empty.bin", runtime="mock", file_size=0)
    assert result.file_size == 0
    assert result.metadata == {}


def test_compile_result_negative_file_size():
    """Negative file size — no validation at dataclass level."""
    result = CompileResult(output_path="/tmp/model.bin", runtime="mock", file_size=-1)
    assert result.file_size == -1


def test_compile_result_empty_strings():
    """Empty strings for output_path and runtime."""
    result = CompileResult(output_path="", runtime="", file_size=0)
    assert result.output_path == ""
    assert result.runtime == ""


def test_compile_result_metadata_not_shared():
    """Two CompileResults should have independent metadata dicts."""
    r1 = CompileResult(output_path="a", runtime="mock", file_size=1)
    r2 = CompileResult(output_path="b", runtime="mock", file_size=2)
    r1.metadata["key"] = "value"
    assert "key" not in r2.metadata


def test_compile_result_large_metadata():
    """Metadata with many keys should work fine."""
    meta = {f"key_{i}": f"value_{i}" for i in range(1000)}
    result = CompileResult(output_path="/tmp/m.bin", runtime="mock", file_size=100, metadata=meta)
    assert len(result.metadata) == 1000
    assert result.metadata["key_500"] == "value_500"


def test_compile_result_nested_metadata():
    """Metadata can contain nested dicts and lists."""
    meta = {
        "versions": {"major": 1, "minor": 2},
        "tags": ["fp16", "optimized"],
    }
    result = CompileResult(output_path="/tmp/m.bin", runtime="trt", file_size=100, metadata=meta)
    assert result.metadata["versions"]["major"] == 1
    assert "fp16" in result.metadata["tags"]


def test_base_compiler_is_abstract():
    """BaseCompiler cannot be instantiated directly."""
    from compiler.compilers.base import BaseCompiler
    import pytest

    with pytest.raises(TypeError):
        BaseCompiler()
