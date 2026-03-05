"""Tests for the OpenVINOCompiler backend with mocked subprocess and dependencies."""
import subprocess
from unittest.mock import MagicMock, patch

import pytest

from compiler.compilers.base import CompileResult


# ---------------------------------------------------------------------------
# Fixture: import OpenVINOCompiler with _check_openvino bypassed
# ---------------------------------------------------------------------------


@pytest.fixture
def openvino_compiler():
    """Import OpenVINOCompiler with _check_openvino bypassed."""
    with patch("compiler.compilers.openvino._check_openvino"):
        import importlib
        import compiler.compilers.openvino as mod
        importlib.reload(mod)
        return mod.OpenVINOCompiler()


# ---------------------------------------------------------------------------
# Basic property tests
# ---------------------------------------------------------------------------


def test_runtime_name(openvino_compiler):
    assert openvino_compiler.runtime_name() == "openvino"


def test_supported_formats(openvino_compiler):
    assert openvino_compiler.supported_formats() == ["onnx"]


# ---------------------------------------------------------------------------
# Compile tests (mocked subprocess)
# ---------------------------------------------------------------------------


@patch("compiler.compilers.openvino.os.path.getsize")
@patch("compiler.compilers.openvino.subprocess.run")
def test_compile_success(mock_run, mock_getsize, openvino_compiler):
    """Mock subprocess.run, verify mo called correctly."""
    mock_run.return_value = MagicMock(returncode=0, stdout="", stderr="")
    # getsize is called twice: once for .xml, once for .bin
    mock_getsize.side_effect = [1024, 2048]

    result = openvino_compiler.compile(
        "/tmp/model.onnx", "/tmp/model.xml", {}
    )

    mock_run.assert_called_once()
    cmd = mock_run.call_args[0][0]
    assert cmd[0] == "mo"
    assert "--input_model" in cmd
    assert "/tmp/model.onnx" in cmd
    assert "--output_dir" in cmd
    assert "--model_name" in cmd

    assert isinstance(result, CompileResult)
    assert result.runtime == "openvino"
    assert result.file_size == 1024 + 2048  # xml + bin sizes
    assert result.metadata["compiler"] == "openvino_mo"


@patch("compiler.compilers.openvino.os.path.getsize")
@patch("compiler.compilers.openvino.subprocess.run")
def test_compile_fp16_option(mock_run, mock_getsize, openvino_compiler):
    """Verify --data_type FP16 flag is passed."""
    mock_run.return_value = MagicMock(returncode=0, stdout="", stderr="")
    mock_getsize.side_effect = [512, 1024]

    openvino_compiler.compile(
        "/tmp/model.onnx", "/tmp/model.xml", {"fp16": True}
    )

    cmd = mock_run.call_args[0][0]
    assert "--data_type" in cmd
    dt_idx = cmd.index("--data_type")
    assert cmd[dt_idx + 1] == "FP16"


@patch("compiler.compilers.openvino.os.path.getsize")
@patch("compiler.compilers.openvino.subprocess.run")
def test_compile_input_shape_option(mock_run, mock_getsize, openvino_compiler):
    """Verify --input_shape flag is passed."""
    mock_run.return_value = MagicMock(returncode=0, stdout="", stderr="")
    mock_getsize.side_effect = [512, 1024]

    openvino_compiler.compile(
        "/tmp/model.onnx", "/tmp/model.xml", {"input_shape": "[1,3,224,224]"}
    )

    cmd = mock_run.call_args[0][0]
    assert "--input_shape" in cmd
    shape_idx = cmd.index("--input_shape")
    assert cmd[shape_idx + 1] == "[1,3,224,224]"


@patch("compiler.compilers.openvino.os.path.getsize")
@patch("compiler.compilers.openvino.subprocess.run")
def test_compile_produces_xml_and_bin(mock_run, mock_getsize, openvino_compiler):
    """Mock both files exist, verify combined size in result."""
    mock_run.return_value = MagicMock(returncode=0, stdout="", stderr="")
    xml_size = 4096
    bin_size = 8192
    mock_getsize.side_effect = [xml_size, bin_size]

    result = openvino_compiler.compile(
        "/tmp/model.onnx", "/tmp/model.xml", {}
    )

    assert result.file_size == xml_size + bin_size
    assert "xml_path" in result.metadata
    assert "bin_path" in result.metadata
    assert result.metadata["xml_path"].endswith(".xml")
    assert result.metadata["bin_path"].endswith(".bin")


@patch("compiler.compilers.openvino.subprocess.run")
def test_compile_subprocess_failure(mock_run, openvino_compiler):
    """Non-zero exit, verify RuntimeError."""
    mock_run.return_value = MagicMock(
        returncode=1, stdout="", stderr="Model optimization failed"
    )

    with pytest.raises(RuntimeError, match="OpenVINO mo failed"):
        openvino_compiler.compile("/tmp/model.onnx", "/tmp/model.xml", {})


@patch("compiler.compilers.openvino.os.path.getsize")
@patch("compiler.compilers.openvino.subprocess.run")
def test_compile_bin_file_missing(mock_run, mock_getsize, openvino_compiler):
    """Subprocess succeeds but .bin file is missing (getsize raises), verify error."""
    mock_run.return_value = MagicMock(returncode=0, stdout="", stderr="")
    # First call (xml) succeeds, second call (bin) raises FileNotFoundError
    mock_getsize.side_effect = [1024, FileNotFoundError("model.bin not found")]

    with pytest.raises(FileNotFoundError):
        openvino_compiler.compile("/tmp/model.onnx", "/tmp/model.xml", {})


# ---------------------------------------------------------------------------
# Validate tests (mocked onnx)
# ---------------------------------------------------------------------------


def test_validate_valid_onnx(openvino_compiler):
    """Mock onnx.load + onnx.checker.check_model, verify True."""
    mock_onnx = MagicMock()
    mock_onnx.load.return_value = MagicMock()
    mock_onnx.checker.check_model.return_value = None

    with patch.dict("sys.modules", {"onnx": mock_onnx}):
        result = openvino_compiler.validate("/tmp/model.onnx")

    assert result is True
    mock_onnx.load.assert_called_once_with("/tmp/model.onnx")
    mock_onnx.checker.check_model.assert_called_once()


def test_validate_invalid_onnx(openvino_compiler):
    """Mock onnx.checker raises, verify False."""
    mock_onnx = MagicMock()
    mock_onnx.load.return_value = MagicMock()
    mock_onnx.checker.check_model.side_effect = Exception("Corrupt ONNX file")

    with patch.dict("sys.modules", {"onnx": mock_onnx}):
        result = openvino_compiler.validate("/tmp/bad_model.onnx")

    assert result is False
