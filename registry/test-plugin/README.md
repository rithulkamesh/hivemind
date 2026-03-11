# Hivemind Plugin Demo

Minimal hivemind plugin for testing the registry. Provides one tool: `demo_echo`.

## Build

```bash
pip install build
python -m build
```

Wheel and sdist are in `dist/`. From repo root you can run `just test-plugin-build`.

## Install from registry (after uploading or seeding)

```bash
pip install --index-url http://localhost:3000/simple/ hivemind-plugin-demo
```

Or from repo root: `just test-plugin-install` (registry must be up and the package must exist).

## Use

Once installed, hivemind will discover the plugin via the `hivemind.plugins` entry point and register the `demo_echo` tool.
