import logging
from typing import Dict, List, Optional

from compiler.compilers.base import BaseCompiler

logger = logging.getLogger(__name__)


class CompilerRegistry:
    """Discovers and manages compiler backends."""

    def __init__(self):
        self._compilers: Dict[str, BaseCompiler] = {}

    def register(self, compiler: BaseCompiler) -> None:
        name = compiler.runtime_name()
        self._compilers[name] = compiler
        logger.info("Registered compiler: %s", name)

    def get(self, runtime: str) -> Optional[BaseCompiler]:
        return self._compilers.get(runtime)

    def list_available(self) -> List[str]:
        return list(self._compilers.keys())

    def auto_discover(self) -> None:
        """Try importing each compiler backend; skip if dependencies are missing."""
        backends = [
            ("compiler.compilers.mock", "MockCompiler"),
            ("compiler.compilers.tensorrt", "TensorRTCompiler"),
            ("compiler.compilers.openvino", "OpenVINOCompiler"),
            ("compiler.compilers.tflite", "TFLiteCompiler"),
        ]
        for module_path, class_name in backends:
            try:
                module = __import__(module_path, fromlist=[class_name])
                cls = getattr(module, class_name)
                instance = cls()
                self.register(instance)
            except Exception as e:
                logger.info("Skipping %s: %s", class_name, e)
