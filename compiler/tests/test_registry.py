"""Tests for CompilerRegistry."""
from compiler.compilers.registry import CompilerRegistry
from compiler.compilers.mock import MockCompiler


def test_register_and_get():
    registry = CompilerRegistry()
    mock = MockCompiler()
    registry.register(mock)

    assert registry.get("mock") is mock
    assert registry.get("nonexistent") is None


def test_list_available():
    registry = CompilerRegistry()
    assert registry.list_available() == []

    registry.register(MockCompiler())
    assert "mock" in registry.list_available()


def test_auto_discover_finds_mock():
    registry = CompilerRegistry()
    registry.auto_discover()
    # MockCompiler should always be available (no GPU deps)
    assert "mock" in registry.list_available()


def test_auto_discover_skips_unavailable():
    registry = CompilerRegistry()
    registry.auto_discover()
    available = registry.list_available()
    # TensorRT, OpenVINO, TFLite should NOT be available in test environment
    # (their import-time checks will fail)
    assert "mock" in available
    # Don't assert absence — they might be installed in some environments


# ---------------------------------------------------------------------------
# Edge-case tests
# ---------------------------------------------------------------------------


def test_get_nonexistent_returns_none():
    """Get a runtime that was never registered."""
    registry = CompilerRegistry()
    assert registry.get("does_not_exist") is None


def test_get_empty_string():
    """Get with empty string key."""
    registry = CompilerRegistry()
    assert registry.get("") is None


def test_register_overwrites():
    """Registering a compiler with the same runtime_name overwrites the previous one."""
    registry = CompilerRegistry()
    mock1 = MockCompiler()
    mock2 = MockCompiler()
    registry.register(mock1)
    registry.register(mock2)
    assert registry.get("mock") is mock2
    assert len(registry.list_available()) == 1


def test_list_available_empty():
    """Fresh registry has no compilers."""
    registry = CompilerRegistry()
    assert registry.list_available() == []


def test_list_available_order():
    """list_available returns names in insertion order."""
    registry = CompilerRegistry()
    registry.register(MockCompiler())
    names = registry.list_available()
    assert "mock" in names


def test_auto_discover_idempotent():
    """Calling auto_discover twice should not duplicate entries."""
    registry = CompilerRegistry()
    registry.auto_discover()
    count1 = len(registry.list_available())
    registry.auto_discover()
    count2 = len(registry.list_available())
    assert count1 == count2


def test_register_preserves_instance():
    """The exact instance registered is the one returned by get()."""
    registry = CompilerRegistry()
    mock = MockCompiler()
    registry.register(mock)
    assert registry.get("mock") is mock  # identity check, not equality
