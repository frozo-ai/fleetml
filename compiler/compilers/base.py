from abc import ABC, abstractmethod


class BaseCompiler(ABC):
    @abstractmethod
    def compile(self, input_path: str, output_path: str, options: dict) -> dict:
        """Compile model. Returns metadata dict."""
        pass

    @abstractmethod
    def validate(self, model_path: str) -> bool:
        """Validate compiled model."""
        pass

    @abstractmethod
    def supported_formats(self) -> list[str]:
        """Return supported input formats."""
        pass
