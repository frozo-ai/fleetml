"""Tests for the MockCompiler backend."""
import os
import tempfile

import pytest

from compiler.compilers.mock import MockCompiler
from compiler.compilers.base import CompileResult


# ---------------------------------------------------------------------------
# Existing tests
# ---------------------------------------------------------------------------


def test_runtime_name():
    compiler = MockCompiler()
    assert compiler.runtime_name() == "mock"


def test_supported_formats():
    compiler = MockCompiler()
    assert "onnx" in compiler.supported_formats()


def test_validate_existing_file():
    compiler = MockCompiler()
    with tempfile.NamedTemporaryFile(suffix=".onnx", delete=False) as f:
        f.write(b"fake onnx content")
        f.flush()
        assert compiler.validate(f.name) is True
    os.unlink(f.name)


def test_validate_missing_file():
    compiler = MockCompiler()
    assert compiler.validate("/nonexistent/model.onnx") is False


def test_compile_copies_file():
    compiler = MockCompiler()
    with tempfile.NamedTemporaryFile(suffix=".onnx", delete=False) as src:
        src.write(b"test model content 12345")
        src.flush()
        src_path = src.name

    with tempfile.NamedTemporaryFile(suffix=".onnx", delete=False) as dst:
        dst_path = dst.name

    try:
        result = compiler.compile(src_path, dst_path, {})
        assert result.runtime == "mock"
        assert result.file_size == len(b"test model content 12345")
        assert result.output_path == dst_path
        assert result.metadata["compiler"] == "mock"

        # Verify content was copied
        with open(dst_path, "rb") as f:
            assert f.read() == b"test model content 12345"
    finally:
        os.unlink(src_path)
        os.unlink(dst_path)


def test_compile_with_options():
    compiler = MockCompiler()
    with tempfile.NamedTemporaryFile(suffix=".onnx", delete=False) as src:
        src.write(b"content")
        src.flush()
        src_path = src.name

    with tempfile.NamedTemporaryFile(suffix=".onnx", delete=False) as dst:
        dst_path = dst.name

    try:
        result = compiler.compile(src_path, dst_path, {"fp16": True})
        assert result.runtime == "mock"
    finally:
        os.unlink(src_path)
        os.unlink(dst_path)


# ---------------------------------------------------------------------------
# New tests
# ---------------------------------------------------------------------------


def test_compile_output_already_exists():
    """If output path already exists, it should be overwritten."""
    compiler = MockCompiler()

    with tempfile.NamedTemporaryFile(suffix=".onnx", delete=False) as src:
        src.write(b"new model content")
        src.flush()
        src_path = src.name

    with tempfile.NamedTemporaryFile(suffix=".onnx", delete=False) as dst:
        # Pre-populate the output file with old content
        dst.write(b"old stale content that should be replaced")
        dst.flush()
        dst_path = dst.name

    try:
        result = compiler.compile(src_path, dst_path, {})
        assert result.output_path == dst_path

        # Verify old content was replaced
        with open(dst_path, "rb") as f:
            content = f.read()
        assert content == b"new model content"
        assert result.file_size == len(b"new model content")
    finally:
        os.unlink(src_path)
        os.unlink(dst_path)


def test_compile_large_file():
    """Create a 10MB temp file, verify correct copy and file size."""
    compiler = MockCompiler()
    large_content = b"X" * (10 * 1024 * 1024)  # 10 MB

    with tempfile.NamedTemporaryFile(suffix=".onnx", delete=False) as src:
        src.write(large_content)
        src.flush()
        src_path = src.name

    with tempfile.NamedTemporaryFile(suffix=".onnx", delete=False) as dst:
        dst_path = dst.name

    try:
        result = compiler.compile(src_path, dst_path, {})
        assert result.file_size == len(large_content)
        assert result.runtime == "mock"

        # Verify content integrity
        with open(dst_path, "rb") as f:
            assert f.read() == large_content
    finally:
        os.unlink(src_path)
        os.unlink(dst_path)


def test_compile_returns_compile_result():
    """Verify CompileResult fields: output_path, runtime, file_size, metadata."""
    compiler = MockCompiler()
    content = b"model data for result test"

    with tempfile.NamedTemporaryFile(suffix=".onnx", delete=False) as src:
        src.write(content)
        src.flush()
        src_path = src.name

    with tempfile.NamedTemporaryFile(suffix=".onnx", delete=False) as dst:
        dst_path = dst.name

    try:
        result = compiler.compile(src_path, dst_path, {})

        # Check it is a CompileResult instance
        assert isinstance(result, CompileResult)

        # Check all fields
        assert result.output_path == dst_path
        assert result.runtime == "mock"
        assert result.file_size == len(content)
        assert isinstance(result.metadata, dict)
        assert result.metadata["compiler"] == "mock"
        assert result.metadata["source_format"] == "onnx"
    finally:
        os.unlink(src_path)
        os.unlink(dst_path)


def test_validate_empty_file():
    """Create a 0-byte file, verify it validates (exists but empty)."""
    compiler = MockCompiler()

    with tempfile.NamedTemporaryFile(suffix=".onnx", delete=False) as f:
        # Write nothing -- 0 bytes
        f.flush()
        path = f.name

    try:
        # MockCompiler.validate checks os.path.isfile, so a 0-byte file is valid
        assert compiler.validate(path) is True
    finally:
        os.unlink(path)


def test_validate_directory():
    """Pass a directory path -- should return False since isfile checks."""
    compiler = MockCompiler()

    with tempfile.TemporaryDirectory() as tmpdir:
        # Directories are not files
        assert compiler.validate(tmpdir) is False


# ---------------------------------------------------------------------------
# Edge-case tests
# ---------------------------------------------------------------------------


def test_runtime_name_is_string():
    """runtime_name() should return a plain string."""
    compiler = MockCompiler()
    name = compiler.runtime_name()
    assert isinstance(name, str)
    assert len(name) > 0


def test_supported_formats_returns_list():
    """supported_formats() returns a non-empty list."""
    compiler = MockCompiler()
    formats = compiler.supported_formats()
    assert isinstance(formats, list)
    assert len(formats) >= 1


def test_compile_nonexistent_source():
    """Compile with nonexistent source should raise FileNotFoundError."""
    compiler = MockCompiler()
    with tempfile.NamedTemporaryFile(suffix=".onnx", delete=False) as dst:
        dst_path = dst.name

    try:
        with pytest.raises(FileNotFoundError):
            compiler.compile("/nonexistent/model.onnx", dst_path, {})
    finally:
        os.unlink(dst_path)


def test_compile_options_ignored():
    """MockCompiler ignores options — any dict should work."""
    compiler = MockCompiler()

    with tempfile.NamedTemporaryFile(suffix=".onnx", delete=False) as src:
        src.write(b"data")
        src.flush()
        src_path = src.name

    with tempfile.NamedTemporaryFile(suffix=".onnx", delete=False) as dst:
        dst_path = dst.name

    try:
        result = compiler.compile(src_path, dst_path, {"fp16": True, "int8": False, "batch": 32})
        assert result.runtime == "mock"
        assert result.metadata["compiler"] == "mock"
    finally:
        os.unlink(src_path)
        os.unlink(dst_path)


def test_compile_empty_options_dict():
    """Empty options dict should work."""
    compiler = MockCompiler()

    with tempfile.NamedTemporaryFile(suffix=".onnx", delete=False) as src:
        src.write(b"data")
        src.flush()
        src_path = src.name

    with tempfile.NamedTemporaryFile(suffix=".onnx", delete=False) as dst:
        dst_path = dst.name

    try:
        result = compiler.compile(src_path, dst_path, {})
        assert result.runtime == "mock"
    finally:
        os.unlink(src_path)
        os.unlink(dst_path)


def test_validate_symlink():
    """Validate should return True for a symlink pointing to a real file."""
    compiler = MockCompiler()

    with tempfile.NamedTemporaryFile(suffix=".onnx", delete=False) as f:
        f.write(b"model data")
        f.flush()
        real_path = f.name

    link_path = real_path + ".link"
    try:
        os.symlink(real_path, link_path)
        assert compiler.validate(link_path) is True
    finally:
        os.unlink(link_path)
        os.unlink(real_path)


def test_validate_empty_string():
    """Empty string path should return False."""
    compiler = MockCompiler()
    assert compiler.validate("") is False


def test_compile_preserves_file_content_integrity():
    """Compile should produce byte-identical output to input (mock copies)."""
    compiler = MockCompiler()
    content = bytes(range(256)) * 100  # 25.6KB of varied bytes

    with tempfile.NamedTemporaryFile(suffix=".onnx", delete=False) as src:
        src.write(content)
        src.flush()
        src_path = src.name

    with tempfile.NamedTemporaryFile(suffix=".onnx", delete=False) as dst:
        dst_path = dst.name

    try:
        result = compiler.compile(src_path, dst_path, {})
        with open(dst_path, "rb") as f:
            output = f.read()
        assert output == content
        assert result.file_size == len(content)
    finally:
        os.unlink(src_path)
        os.unlink(dst_path)


def test_compile_metadata_has_expected_keys():
    """Verify metadata contains 'compiler' and 'source_format' keys."""
    compiler = MockCompiler()

    with tempfile.NamedTemporaryFile(suffix=".onnx", delete=False) as src:
        src.write(b"x")
        src.flush()
        src_path = src.name

    with tempfile.NamedTemporaryFile(suffix=".onnx", delete=False) as dst:
        dst_path = dst.name

    try:
        result = compiler.compile(src_path, dst_path, {})
        assert "compiler" in result.metadata
        assert "source_format" in result.metadata
        assert result.metadata["compiler"] == "mock"
        assert result.metadata["source_format"] == "onnx"
    finally:
        os.unlink(src_path)
        os.unlink(dst_path)
