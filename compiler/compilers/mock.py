import os
import shutil

from compiler.compilers.base import BaseCompiler, CompileResult


class MockCompiler(BaseCompiler):
    """Mock compiler that copies the ONNX file as-is. Always available, used for testing."""

    def runtime_name(self) -> str:
        return "mock"

    def compile(self, input_path: str, output_path: str, options: dict) -> CompileResult:
        shutil.copy2(input_path, output_path)
        file_size = os.path.getsize(output_path)
        return CompileResult(
            output_path=output_path,
            runtime="mock",
            file_size=file_size,
            metadata={"compiler": "mock", "source_format": "onnx"},
        )

    def validate(self, model_path: str) -> bool:
        return os.path.isfile(model_path)

    def supported_formats(self) -> list[str]:
        return ["onnx"]
