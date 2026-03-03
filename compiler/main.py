from fastapi import FastAPI
from fastapi.responses import JSONResponse

app = FastAPI(title="FleetML Compiler Service", version="0.1.0")


@app.get("/health")
def health():
    return {"status": "healthy", "version": "0.1.0"}


@app.post("/compile")
def compile_model():
    """Multi-chip model compilation — coming in v0.2.0."""
    return JSONResponse(
        status_code=501,
        content={
            "error": "Not implemented",
            "message": "Multi-chip model compilation is coming in v0.2.0. "
                       "For now, use ONNX models directly.",
            "supported_targets": ["tensorrt", "openvino", "tflite", "snpe", "hailo"],
        },
    )


@app.get("/supported-runtimes")
def supported_runtimes():
    """List supported compilation targets."""
    return {
        "runtimes": [
            {"name": "onnx", "status": "supported", "description": "Universal ONNX Runtime"},
            {"name": "tensorrt", "status": "planned_v0.2", "description": "NVIDIA TensorRT"},
            {"name": "openvino", "status": "planned_v0.2", "description": "Intel OpenVINO"},
            {"name": "tflite", "status": "planned_v0.2", "description": "TensorFlow Lite"},
            {"name": "snpe", "status": "planned_v0.2", "description": "Qualcomm SNPE"},
            {"name": "hailo", "status": "planned_v0.2", "description": "Hailo Runtime"},
        ]
    }


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8081)
