from pydantic import BaseModel


class CompileRequest(BaseModel):
    model_url: str  # S3 URL of source ONNX model
    model_id: str  # UUID for S3 path construction
    target_runtime: str  # tensorrt | openvino | tflite | mock
    options: dict = {}  # Runtime-specific options (fp16, int8, etc.)


class CompileResponse(BaseModel):
    runtime: str
    artifact_url: str  # S3 URL of compiled variant
    checksum: str  # sha256 of compiled artifact
    file_size: int
    compile_time_seconds: float
    metadata: dict = {}  # Runtime-specific info (TRT version, etc.)


class RuntimeInfo(BaseModel):
    name: str
    status: str  # available | unavailable
    description: str
