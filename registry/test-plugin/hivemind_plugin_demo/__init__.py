"""Minimal hivemind plugin: one demo tool for testing the registry."""

from hivemind.tools.base import Tool
from hivemind.tools.registry import register


class DemoEchoTool(Tool):
    name = "demo_echo"
    description = "Echo a message (demo plugin for registry testing)"
    input_schema = {
        "type": "object",
        "properties": {
            "message": {"type": "string", "description": "Message to echo"},
        },
        "optional": ["message"],
    }
    category = "system"

    def run(self, **kwargs) -> str:
        return str(kwargs.get("message", "(no message)"))


def load():
    """Plugin entry point: register tools and return Tool instances."""
    tool = DemoEchoTool()
    register(tool)
    return [tool]
