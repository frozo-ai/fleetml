import os
import subprocess

from compiler.compilers.base import BaseCompiler, CompileResult


def _check_openvino() -> None:
    """Verify OpenVINO Model Optimizer is available."""
    result = subprocess.run(
        ["mo", "--help"], capture_output=True, timeout=10
    )
    if result.returncode != 0:
        raise RuntimeError("OpenVINO Model Optimizer (mo) not found")


# Fail fast at import time if OpenVINO is not installed
_check_openvino()


class OpenVINOCompiler(BaseCompiler):
    """Compiles ONNX models to OpenVINO IR format using Model Optimizer."""

    def runtime_name(self) -> str:
        return "openvino"

    def compile(self, input_path: str, output_path: str, options: dict) -> CompileResult:
        output_dir = os.path.dirname(output_path)
        output_name = os.path.splitext(os.path.basename(output_path))[0]

        cmd = [
            "mo",
            "--input_model", input_path,
            "--output_dir", output_dir,
            "--model_name", output_name,
        ]

        if options.get("fp16"):
            cmd.extend(["--data_type", "FP16"])
        if options.get("input_shape"):
            cmd.extend(["--input_shape", options["input_shape"]])

        result = subprocess.run(cmd, capture_output=True, text=True, timeout=600)
        if result.returncode != 0:
            raise RuntimeError(f"OpenVINO mo failed: {result.stderr}")

        # OpenVINO produces .xml + .bin; use the .xml as the main artifact
        xml_path = os.path.join(output_dir, f"{output_name}.xml")
        bin_path = os.path.join(output_dir, f"{output_name}.bin")
        file_size = os.path.getsize(xml_path) + os.path.getsize(bin_path)

        return CompileResult(
            output_path=xml_path,
            runtime="openvino",
            file_size=file_size,
            metadata={
                "compiler": "openvino_mo",
                "xml_path": xml_path,
                "bin_path": bin_path,
                "fp16": options.get("fp16", False),
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
