"""Tests for the TensorRTCompiler backend with mocked subprocess and dependencies."""
import subprocess
from unittest.mock import MagicMock, patch, mock_open

import pytest

from compiler.compilers.base import CompileResult


# ---------------------------------------------------------------------------
# We must patch _check_tensorrt at import time because the module runs it
# during import.  We import the class inside each test after patching.
# ---------------------------------------------------------------------------


@pytest.fixture
def tensorrt_compiler():
    """Import TensorRTCompiler with _check_tensorrt bypassed."""
    with patch("compiler.compilers.tensorrt._check_tensorrt"):
        # Force re-import to get a fresh module with the check bypassed
        import importlib
        import compiler.compilers.tensorrt as mod
        importlib.reload(mod)
        return mod.TensorRTCompiler()


def _get_tensorrt_class():
    """Helper to get TensorRTCompiler class with check bypassed."""
    with patch("compiler.compilers.tensorrt._check_tensorrt"):
        import importlib
        import compiler.compilers.tensorrt as mod
        importlib.reload(mod)
        return mod.TensorRTCompiler


# ---------------------------------------------------------------------------
# Basic property tests
# ---------------------------------------------------------------------------


def test_runtime_name(tensorrt_compiler):
    assert tensorrt_compiler.runtime_name() == "tensorrt"


def test_supported_formats(tensorrt_compiler):
    assert tensorrt_compiler.supported_formats() == ["onnx"]


# ---------------------------------------------------------------------------
# Compile tests (mocked subprocess)
# ---------------------------------------------------------------------------


@patch("compiler.compilers.tensorrt.os.path.getsize", return_value=4096)
@patch("compiler.compilers.tensorrt.subprocess.run")
def test_compile_success(mock_run, mock_getsize, tensorrt_compiler):
    """Mock subprocess.run, verify trtexec called with correct args."""
    mock_run.return_value = MagicMock(returncode=0, stdout="success", stderr="")

    result = tensorrt_compiler.compile("/tmp/model.onnx", "/tmp/model.trt", {})

    # Verify subprocess was called
    mock_run.assert_called_once()
    call_args = mock_run.call_args
    cmd = call_args[0][0]
    assert cmd[0] == "trtexec"
    assert "--onnx=/tmp/model.onnx" in cmd
    assert "--saveEngine=/tmp/model.trt" in cmd

    # Verify result
    assert isinstance(result, CompileResult)
    assert result.runtime == "tensorrt"
    assert result.file_size == 4096
    assert result.output_path == "/tmp/model.trt"


@patch("compiler.compilers.tensorrt.os.path.getsize", return_value=2048)
@patch("compiler.compilers.tensorrt.subprocess.run")
def test_compile_fp16_option(mock_run, mock_getsize, tensorrt_compiler):
    """Verify --fp16 flag is passed when fp16 option is set."""
    mock_run.return_value = MagicMock(returncode=0, stdout="", stderr="")

    tensorrt_compiler.compile("/tmp/model.onnx", "/tmp/model.trt", {"fp16": True})

    cmd = mock_run.call_args[0][0]
    assert "--fp16" in cmd


@patch("compiler.compilers.tensorrt.os.path.getsize", return_value=2048)
@patch("compiler.compilers.tensorrt.subprocess.run")
def test_compile_int8_option(mock_run, mock_getsize, tensorrt_compiler):
    """Verify --int8 flag is passed when int8 option is set."""
    mock_run.return_value = MagicMock(returncode=0, stdout="", stderr="")

    tensorrt_compiler.compile("/tmp/model.onnx", "/tmp/model.trt", {"int8": True})

    cmd = mock_run.call_args[0][0]
    assert "--int8" in cmd


@patch("compiler.compilers.tensorrt.os.path.getsize", return_value=2048)
@patch("compiler.compilers.tensorrt.subprocess.run")
def test_compile_workspace_option(mock_run, mock_getsize, tensorrt_compiler):
    """Verify --workspace flag is passed when workspace_mb option is set."""
    mock_run.return_value = MagicMock(returncode=0, stdout="", stderr="")

    tensorrt_compiler.compile(
        "/tmp/model.onnx", "/tmp/model.trt", {"workspace_mb": 2048}
    )

    cmd = mock_run.call_args[0][0]
    assert "--workspace=2048" in cmd


@patch("compiler.compilers.tensorrt.subprocess.run")
def test_compile_subprocess_failure(mock_run, tensorrt_compiler):
    """Subprocess returns non-zero, verify RuntimeError raised."""
    mock_run.return_value = MagicMock(
        returncode=1, stdout="", stderr="CUDA out of memory"
    )

    with pytest.raises(RuntimeError, match="trtexec failed"):
        tensorrt_compiler.compile("/tmp/model.onnx", "/tmp/model.trt", {})


@patch("compiler.compilers.tensorrt.subprocess.run")
def test_compile_timeout(mock_run, tensorrt_compiler):
    """subprocess.TimeoutExpired raised, verify it propagates."""
    mock_run.side_effect = subprocess.TimeoutExpired(cmd="trtexec", timeout=600)

    with pytest.raises(subprocess.TimeoutExpired):
        tensorrt_compiler.compile("/tmp/model.onnx", "/tmp/model.trt", {})


# ---------------------------------------------------------------------------
# Validate tests (mocked onnx)
# ---------------------------------------------------------------------------


def test_validate_valid_onnx(tensorrt_compiler):
    """Mock onnx.load + onnx.checker.check_model, verify True."""
    mock_onnx = MagicMock()
    mock_onnx.load.return_value = MagicMock()  # fake model
    mock_onnx.checker.check_model.return_value = None  # no error

    with patch.dict("sys.modules", {"onnx": mock_onnx}):
        result = tensorrt_compiler.validate("/tmp/model.onnx")

    assert result is True
    mock_onnx.load.assert_called_once_with("/tmp/model.onnx")
    mock_onnx.checker.check_model.assert_called_once()


def test_validate_invalid_onnx(tensorrt_compiler):
    """Mock onnx.checker raises, verify False."""
    mock_onnx = MagicMock()
    mock_onnx.load.return_value = MagicMock()
    mock_onnx.checker.check_model.side_effect = Exception("Invalid model")

    with patch.dict("sys.modules", {"onnx": mock_onnx}):
        result = tensorrt_compiler.validate("/tmp/bad_model.onnx")

    assert result is False


def test_validate_missing_file(tensorrt_compiler):
    """File does not exist and onnx.load raises, verify False."""
    mock_onnx = MagicMock()
    mock_onnx.load.side_effect = FileNotFoundError("No such file")

    with patch.dict("sys.modules", {"onnx": mock_onnx}):
        result = tensorrt_compiler.validate("/tmp/missing.onnx")

    assert result is False
