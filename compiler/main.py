import hashlib
import logging
import os
import tempfile
import time

from fastapi import FastAPI, HTTPException
from fastapi.responses import JSONResponse

from compiler.compilers.registry import CompilerRegistry
from compiler.schemas import CompileRequest, CompileResponse, RuntimeInfo
from compiler.storage import S3Client

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = FastAPI(title="FleetML Compiler Service", version="0.2.0")

# Initialize compiler registry with auto-discovery
registry = CompilerRegistry()
registry.auto_discover()

# Initialize S3 client
s3 = S3Client()

logger.info("Available compilers: %s", registry.list_available())


@app.get("/health")
def health():
    return {
        "status": "healthy",
        "version": "0.2.0",
        "available_runtimes": registry.list_available(),
    }


@app.get("/supported-runtimes")
def supported_runtimes():
    available = set(registry.list_available())
    all_runtimes = [
        ("mock", "Mock compiler for testing"),
        ("tensorrt", "NVIDIA TensorRT"),
        ("openvino", "Intel OpenVINO"),
        ("tflite", "TensorFlow Lite"),
    ]
    runtimes = [
        RuntimeInfo(
            name=name,
            status="available" if name in available else "unavailable",
            description=desc,
        )
        for name, desc in all_runtimes
    ]
    return {"runtimes": [r.model_dump() for r in runtimes]}


@app.post("/compile", response_model=CompileResponse)
def compile_model(req: CompileRequest):
    compiler = registry.get(req.target_runtime)
    if compiler is None:
        available = registry.list_available()
        raise HTTPException(
            status_code=400,
            detail=f"Unsupported runtime '{req.target_runtime}'. Available: {available}",
        )

    with tempfile.TemporaryDirectory() as tmpdir:
        # 1. Download source ONNX from S3
        input_path = os.path.join(tmpdir, "model.onnx")
        try:
            s3.download(req.model_url, input_path)
        except Exception as e:
            raise HTTPException(status_code=400, detail=f"Failed to download model: {e}")

        # 2. Validate ONNX file
        if not compiler.validate(input_path):
            raise HTTPException(status_code=400, detail="Invalid ONNX model file")

        # 3. Compile to target runtime
        ext = _runtime_extension(req.target_runtime)
        output_path = os.path.join(tmpdir, f"model.{ext}")

        start = time.monotonic()
        try:
            result = compiler.compile(input_path, output_path, req.options)
        except Exception as e:
            logger.error("Compilation failed: %s", e)
            raise HTTPException(status_code=500, detail=f"Compilation failed: {e}")
        compile_time = time.monotonic() - start

        # 4. Compute checksum of compiled artifact
        checksum = _sha256(result.output_path)

        # 5. Upload compiled artifact to S3
        s3_key = f"{req.model_id}/compiled/{req.target_runtime}/model.{ext}"
        try:
            artifact_url = s3.upload(result.output_path, s3_key)
        except Exception as e:
            raise HTTPException(status_code=500, detail=f"Failed to upload compiled artifact: {e}")

        return CompileResponse(
            runtime=req.target_runtime,
            artifact_url=artifact_url,
            checksum=f"sha256:{checksum}",
            file_size=result.file_size,
            compile_time_seconds=round(compile_time, 3),
            metadata=result.metadata,
        )


def _runtime_extension(runtime: str) -> str:
    extensions = {
        "mock": "onnx",
        "tensorrt": "trt",
        "openvino": "xml",
        "tflite": "tflite",
    }
    return extensions.get(runtime, "bin")


def _sha256(file_path: str) -> str:
    h = hashlib.sha256()
    with open(file_path, "rb") as f:
        for chunk in iter(lambda: f.read(8192), b""):
            h.update(chunk)
    return h.hexdigest()


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8081)
