export function PublishGuide() {
  return (
    <div className="max-w-prose">
      <h1 className="font-sans text-2xl font-semibold text-hm-text mb-4">
        Publish a plugin
      </h1>
      <p className="text-hm-text-passive mb-6">
        Use the registry index when building and publishing your hivemind plugin.
      </p>
      <div className="bg-hm-code-bg border border-hm-border border-l-4 border-l-hm-amber p-4 font-mono text-sm text-hm-text mb-6">
        <p className="text-hm-muted mb-2"># Configure your package (pyproject.toml)</p>
        <pre className="whitespace-pre-wrap break-words">
{`[tool.setuptools.packages.find]
[tool.hivemind]
index-url = "https://registry.hivemind.rithul.dev/simple/"`}
        </pre>
      </div>
      <p className="text-hm-muted text-sm">
        After registering and logging in, create a package and upload wheels or sdists via the dashboard or API.
        Verification (safety, pip-audit, bandit) runs automatically before a version is published.
      </p>
    </div>
  );
}
