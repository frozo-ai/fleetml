import os
import subprocess

from compiler.compilers.base import BaseCompiler, CompileResult


def _check_tensorrt() -> None:
    """Verify trtexec is available."""
    result = subprocess.run(
        ["trtexec", "--help"], capture_output=True, timeout=10
    )
    if result.returncode != 0:
        raise RuntimeError("trtexec not found or not working")


# Fail fast at import time if TensorRT is not installed
_check_tensorrt()


class TensorRTCompiler(BaseCompiler):
    """Compiles ONNX models to TensorRT engine files using trtexec."""

    def runtime_name(self) -> str:
        return "tensorrt"

    def compile(self, input_path: str, output_path: str, options: dict) -> CompileResult:
        cmd = [
            "trtexec",
            f"--onnx={input_path}",
            f"--saveEngine={output_path}",
        ]

        if options.get("fp16"):
            cmd.append("--fp16")
        if options.get("int8"):
            cmd.append("--int8")
        if options.get("workspace_mb"):
            cmd.append(f"--workspace={options['workspace_mb']}")

        result = subprocess.run(cmd, capture_output=True, text=True, timeout=600)
        if result.returncode != 0:
            raise RuntimeError(f"trtexec failed: {result.stderr}")

        file_size = os.path.getsize(output_path)
        return CompileResult(
            output_path=output_path,
            runtime="tensorrt",
            file_size=file_size,
            metadata={
                "compiler": "trtexec",
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

    def supported_formats(self):
        return ["onnx"]
