"""Tests for the TFLiteCompiler backend with mocked imports."""
import subprocess
from unittest.mock import MagicMock, patch, mock_open

import pytest

from compiler.compilers.base import CompileResult


# ---------------------------------------------------------------------------
# Fixture: import TFLiteCompiler with _check_tflite bypassed
# ---------------------------------------------------------------------------


@pytest.fixture
def tflite_compiler():
    """Import TFLiteCompiler with _check_tflite bypassed."""
    with patch("compiler.compilers.tflite._check_tflite"):
        import importlib
        import compiler.compilers.tflite as mod
        importlib.reload(mod)
        return mod.TFLiteCompiler()


# ---------------------------------------------------------------------------
# Basic property tests
# ---------------------------------------------------------------------------


def test_runtime_name(tflite_compiler):
    assert tflite_compiler.runtime_name() == "tflite"


def test_supported_formats(tflite_compiler):
    assert tflite_compiler.supported_formats() == ["onnx"]


# ---------------------------------------------------------------------------
# Compile tests (mocked onnx, onnx_tf, tensorflow)
# ---------------------------------------------------------------------------


@patch("compiler.compilers.tflite.os.path.getsize", return_value=2048)
def test_compile_success(mock_getsize, tflite_compiler):
    """Mock onnx, onnx_tf, tensorflow imports, verify compile pipeline."""
    mock_onnx = MagicMock()
    mock_onnx_model = MagicMock()
    mock_onnx.load.return_value = mock_onnx_model

    mock_tf_rep = MagicMock()
    mock_prepare = MagicMock(return_value=mock_tf_rep)

    mock_tf = MagicMock()
    mock_converter = MagicMock()
    mock_tf.lite.TFLiteConverter.from_saved_model.return_value = mock_converter
    mock_converter.convert.return_value = b"fake tflite model bytes"

    with patch.dict("sys.modules", {
        "onnx": mock_onnx,
        "onnx_tf": MagicMock(),
        "onnx_tf.backend": MagicMock(prepare=mock_prepare),
        "tensorflow": mock_tf,
    }):
        m = mock_open()
        with patch("builtins.open", m):
            result = tflite_compiler.compile(
                "/tmp/model.onnx", "/tmp/model.tflite", {}
            )

    # Verify the ONNX model was loaded
    mock_onnx.load.assert_called_once_with("/tmp/model.onnx")

    # Verify onnx-tf prepare was called
    mock_prepare.assert_called_once_with(mock_onnx_model)

    # Verify TF SavedModel was exported
    mock_tf_rep.export_graph.assert_called_once_with("/tmp/model.tflite_savedmodel")

    # Verify TFLite conversion
    mock_tf.lite.TFLiteConverter.from_saved_model.assert_called_once_with(
        "/tmp/model.tflite_savedmodel"
    )
    mock_converter.convert.assert_called_once()

    # Verify result
    assert isinstance(result, CompileResult)
    assert result.runtime == "tflite"
    assert result.file_size == 2048
    assert result.metadata["compiler"] == "onnx-tf"


@patch("compiler.compilers.tflite.os.path.getsize", return_value=1024)
def test_compile_fp16_optimization(mock_getsize, tflite_compiler):
    """Verify DEFAULT optimizations and float16 type are set for fp16."""
    mock_onnx = MagicMock()
    mock_onnx.load.return_value = MagicMock()

    mock_tf_rep = MagicMock()
    mock_prepare = MagicMock(return_value=mock_tf_rep)

    mock_tf = MagicMock()
    mock_converter = MagicMock()
    mock_tf.lite.TFLiteConverter.from_saved_model.return_value = mock_converter
    mock_converter.convert.return_value = b"fp16 model"

    with patch.dict("sys.modules", {
        "onnx": mock_onnx,
        "onnx_tf": MagicMock(),
        "onnx_tf.backend": MagicMock(prepare=mock_prepare),
        "tensorflow": mock_tf,
    }):
        m = mock_open()
        with patch("builtins.open", m):
            result = tflite_compiler.compile(
                "/tmp/model.onnx", "/tmp/model.tflite", {"fp16": True}
            )

    # Verify optimizations were set
    assert mock_converter.optimizations == [mock_tf.lite.Optimize.DEFAULT]
    assert mock_converter.target_spec.supported_types == [mock_tf.float16]
    assert result.metadata["fp16"] is True


@patch("compiler.compilers.tflite.os.path.getsize", return_value=1024)
def test_compile_int8_optimization(mock_getsize, tflite_compiler):
    """Verify DEFAULT optimizations for int8 quantization."""
    mock_onnx = MagicMock()
    mock_onnx.load.return_value = MagicMock()

    mock_tf_rep = MagicMock()
    mock_prepare = MagicMock(return_value=mock_tf_rep)

    mock_tf = MagicMock()
    mock_converter = MagicMock()
    mock_tf.lite.TFLiteConverter.from_saved_model.return_value = mock_converter
    mock_converter.convert.return_value = b"int8 model"

    with patch.dict("sys.modules", {
        "onnx": mock_onnx,
        "onnx_tf": MagicMock(),
        "onnx_tf.backend": MagicMock(prepare=mock_prepare),
        "tensorflow": mock_tf,
    }):
        m = mock_open()
        with patch("builtins.open", m):
            result = tflite_compiler.compile(
                "/tmp/model.onnx", "/tmp/model.tflite", {"int8": True}
            )

    # Verify DEFAULT optimizations were set (int8 path)
    assert mock_converter.optimizations == [mock_tf.lite.Optimize.DEFAULT]
    assert result.metadata["int8"] is True


def test_compile_onnx_load_failure(tflite_compiler):
    """onnx.load() raises, verify error propagated."""
    mock_onnx = MagicMock()
    mock_onnx.load.side_effect = Exception("Corrupt ONNX file")

    with patch.dict("sys.modules", {
        "onnx": mock_onnx,
        "onnx_tf": MagicMock(),
        "onnx_tf.backend": MagicMock(),
        "tensorflow": MagicMock(),
    }):
        with pytest.raises(Exception, match="Corrupt ONNX file"):
            tflite_compiler.compile(
                "/tmp/bad.onnx", "/tmp/model.tflite", {}
            )


def test_compile_tf_conversion_failure(tflite_compiler):
    """TFLite converter fails, verify error propagated."""
    mock_onnx = MagicMock()
    mock_onnx.load.return_value = MagicMock()

    mock_tf_rep = MagicMock()
    mock_prepare = MagicMock(return_value=mock_tf_rep)

    mock_tf = MagicMock()
    mock_converter = MagicMock()
    mock_tf.lite.TFLiteConverter.from_saved_model.return_value = mock_converter
    mock_converter.convert.side_effect = RuntimeError("TFLite conversion failed")

    with patch.dict("sys.modules", {
        "onnx": mock_onnx,
        "onnx_tf": MagicMock(),
        "onnx_tf.backend": MagicMock(prepare=mock_prepare),
        "tensorflow": mock_tf,
    }):
        with pytest.raises(RuntimeError, match="TFLite conversion failed"):
            tflite_compiler.compile(
                "/tmp/model.onnx", "/tmp/model.tflite", {}
            )


# ---------------------------------------------------------------------------
# Validate tests (mocked onnx)
# ---------------------------------------------------------------------------


def test_validate_valid_onnx(tflite_compiler):
    """Mock onnx.load + onnx.checker.check_model, verify True."""
    mock_onnx = MagicMock()
    mock_onnx.load.return_value = MagicMock()
    mock_onnx.checker.check_model.return_value = None

    with patch.dict("sys.modules", {"onnx": mock_onnx}):
        result = tflite_compiler.validate("/tmp/model.onnx")

    assert result is True
    mock_onnx.load.assert_called_once_with("/tmp/model.onnx")
    mock_onnx.checker.check_model.assert_called_once()


def test_validate_invalid_onnx(tflite_compiler):
    """Mock onnx.checker raises, verify False."""
    mock_onnx = MagicMock()
    mock_onnx.load.return_value = MagicMock()
    mock_onnx.checker.check_model.side_effect = Exception("Bad ONNX")

    with patch.dict("sys.modules", {"onnx": mock_onnx}):
        result = tflite_compiler.validate("/tmp/bad.onnx")

    assert result is False
