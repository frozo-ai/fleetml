from abc import ABC, abstractmethod
from dataclasses import dataclass, field
from typing import List


@dataclass
class CompileResult:
    """Result of a compilation operation."""
    output_path: str
    runtime: str
    file_size: int
    metadata: dict = field(default_factory=dict)


class BaseCompiler(ABC):
    @abstractmethod
    def runtime_name(self) -> str:
        """Return the runtime identifier (e.g. 'tensorrt', 'openvino')."""
        pass

    @abstractmethod
    def compile(self, input_path: str, output_path: str, options: dict) -> CompileResult:
        """Compile model from ONNX to target runtime. Returns CompileResult."""
        pass

    @abstractmethod
    def validate(self, model_path: str) -> bool:
        """Validate that the input model is a valid ONNX file."""
        pass

    @abstractmethod
    def supported_formats(self) -> List[str]:
        """Return supported input formats (e.g. ['onnx'])."""
        pass
