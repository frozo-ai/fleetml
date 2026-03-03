import os
import subprocess

from compiler.compilers.base import BaseCompiler, CompileResult


def _check_tflite() -> None:
    """Verify onnx-tf or tflite_convert is available."""
    # Check for onnx-tf (ONNX → TF → TFLite pipeline)
    try:
        import onnx_tf  # noqa: F401
        return
    except ImportError:
        pass
    # Fall back to checking tflite_convert CLI
    result = subprocess.run(
        ["tflite_convert", "--help"], capture_output=True, timeout=10
    )
    if result.returncode != 0:
        raise RuntimeError("Neither onnx-tf nor tflite_convert found")


# Fail fast at import time if TFLite tooling is not installed
_check_tflite()


class TFLiteCompiler(BaseCompiler):
    """Compiles ONNX models to TFLite format via onnx-tf."""

    def runtime_name(self) -> str:
        return "tflite"

    def compile(self, input_path: str, output_path: str, options: dict) -> CompileResult:
        import onnx
        from onnx_tf.backend import prepare
        import tensorflow as tf

        # ONNX -> TF SavedModel
        onnx_model = onnx.load(input_path)
        tf_rep = prepare(onnx_model)
        saved_model_dir = output_path + "_savedmodel"
        tf_rep.export_graph(saved_model_dir)

        # TF SavedModel -> TFLite
        converter = tf.lite.TFLiteConverter.from_saved_model(saved_model_dir)
        if options.get("fp16"):
            converter.optimizations = [tf.lite.Optimize.DEFAULT]
            converter.target_spec.supported_types = [tf.float16]
        if options.get("int8"):
            converter.optimizations = [tf.lite.Optimize.DEFAULT]

        tflite_model = converter.convert()
        with open(output_path, "wb") as f:
            f.write(tflite_model)

        file_size = os.path.getsize(output_path)
        return CompileResult(
            output_path=output_path,
            runtime="tflite",
            file_size=file_size,
            metadata={
                "compiler": "onnx-tf",
                "fp16": options.get("fp16", False),
                "int8": options.get("int8", False),
            },
        )

    def validate(self, model_path: str) -> bool:
        try:
            import onnx
            model = onnx.load(model_path)
            onnx.checker.check_model(model)
            return True
        except Exception:
            return False

    def supported_formats(self) -> list[str]:
        return ["onnx"]
