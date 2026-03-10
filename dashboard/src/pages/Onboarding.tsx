import { useState, useEffect } from 'react';
import { useAuth } from '../contexts/AuthContext';
import { api } from '../api/client';

function CopyButton({ text }: { text: string }) {
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    await navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <button
      onClick={handleCopy}
      className="absolute top-2 right-2 px-2 py-1 rounded text-xs bg-gray-700 text-gray-300 hover:bg-gray-600 transition-colors"
    >
      {copied ? 'Copied!' : 'Copy'}
    </button>
  );
}

function CodeBlock({ code }: { code: string }) {
  return (
    <div className="relative group">
      <pre className="bg-gray-900 text-gray-100 rounded-lg p-4 text-sm font-mono overflow-x-auto whitespace-pre-wrap break-all">
        <span className="text-gray-500 select-none">$ </span>
        {code}
      </pre>
      <CopyButton text={code} />
    </div>
  );
}

function APIKeyDisplay({ apiKey, onRegenerate }: { apiKey: string; onRegenerate: () => void }) {
  const [visible, setVisible] = useState(false);
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    await navigator.clipboard.writeText(apiKey);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const masked = apiKey ? apiKey.slice(0, 8) + '••••••••••••••••••••' : '—';

  return (
    <div className="bg-gray-50 border border-gray-200 rounded-lg p-4">
      <div className="flex items-center justify-between mb-2">
        <span className="text-sm font-medium text-gray-700">Your API Key</span>
        <div className="flex items-center gap-2">
          <button
            onClick={() => setVisible(!visible)}
            className="text-xs text-gray-500 hover:text-gray-700"
          >
            {visible ? 'Hide' : 'Reveal'}
          </button>
          <button
            onClick={handleCopy}
            className="text-xs text-blue-600 hover:text-blue-800 font-medium"
          >
            {copied ? 'Copied!' : 'Copy'}
          </button>
        </div>
      </div>
      <code className="block bg-white border border-gray-200 rounded px-3 py-2 text-sm font-mono text-gray-800 break-all">
        {visible ? apiKey : masked}
      </code>
      <div className="mt-2 flex items-center justify-between">
        <p className="text-xs text-gray-400">
          Keep this key secret. It authenticates your devices to your account.
        </p>
        <button
          onClick={onRegenerate}
          className="text-xs text-red-500 hover:text-red-700"
        >
          Regenerate
        </button>
      </div>
    </div>
  );
}

export default function OnboardingPage() {
  const { user } = useAuth();
  const [apiKey, setApiKey] = useState('');
  const [apiKeyLoading, setApiKeyLoading] = useState(true);
  const [completedSteps, setCompletedSteps] = useState<Set<string>>(new Set());

  useEffect(() => {
    api.getAPIKey()
      .then((data) => setApiKey(data.api_key))
      .catch(() => {})
      .finally(() => setApiKeyLoading(false));
  }, []);

  const handleRegenerate = async () => {
    if (!confirm('Regenerate API key? All existing agents will need to be reconfigured.')) return;
    try {
      const data = await api.regenerateAPIKey();
      setApiKey(data.api_key);
    } catch {
      alert('Failed to regenerate API key');
    }
  };

  const toggleStep = (id: string) => {
    setCompletedSteps((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  };

  const steps = [
    {
      id: 'cli',
      title: 'Install the FleetML CLI',
      description: 'The CLI is how you deploy models, manage fleets, and check device status from your terminal.',
      content: (
        <div className="space-y-3">
          <CodeBlock code="curl -sSL https://raw.githubusercontent.com/ashish-frozo/fleetML/main/scripts/install.sh | bash" />
          <p className="text-xs text-gray-400">Verify installation:</p>
          <CodeBlock code="fleetml version" />
        </div>
      ),
    },
    {
      id: 'connect',
      title: 'Connect the CLI to your account',
      description: 'This links your local CLI to FleetML Cloud so you can manage your fleet.',
      content: (
        <div className="space-y-3">
          <CodeBlock code="fleetml init --cloud" />
          <p className="text-xs text-gray-400">
            Enter the email and password you used to sign up. The CLI will save your credentials locally.
          </p>
        </div>
      ),
    },
    {
      id: 'apikey',
      title: 'Copy your API key',
      description: 'Your devices use this key to authenticate with your FleetML account. You\'ll need it in the next step.',
      content: (
        <div className="space-y-3">
          {apiKeyLoading ? (
            <div className="text-gray-400 text-sm">Loading API key...</div>
          ) : apiKey ? (
            <APIKeyDisplay apiKey={apiKey} onRegenerate={handleRegenerate} />
          ) : (
            <div className="text-gray-400 text-sm">
              No API key found. This may be because your account was created before API keys were introduced.
              <button onClick={handleRegenerate} className="ml-2 text-blue-600 hover:underline">Generate one</button>
            </div>
          )}
        </div>
      ),
    },
    {
      id: 'agent',
      title: 'Install the agent on your edge device',
      description: 'SSH into your edge device (Jetson, Raspberry Pi, Intel NUC, etc.) and run:',
      content: (
        <div className="space-y-3">
          <CodeBlock code="curl -sSL https://raw.githubusercontent.com/ashish-frozo/fleetML/main/scripts/install-agent.sh | sh" />
          <p className="text-xs text-gray-400 font-medium mt-2">Then configure it with your API key:</p>
          <CodeBlock code={apiKey
            ? `export FLEETML_API_KEY="${apiKey}"\nexport FLEETML_SERVER="server-production-91d4.up.railway.app:50051"\nfleetml-agent`
            : 'export FLEETML_API_KEY="your-api-key-here"\nexport FLEETML_SERVER="server-production-91d4.up.railway.app:50051"\nfleetml-agent'
          } />
          <p className="text-xs text-gray-400">
            The agent is a lightweight (~15MB) binary that runs on the device. It connects to FleetML Cloud, receives model updates, and reports health metrics.
          </p>
        </div>
      ),
    },
    {
      id: 'deploy',
      title: 'Deploy your first model',
      description: 'Push an ONNX model to your fleet. FleetML handles compilation, distribution, and zero-downtime swap.',
      content: (
        <div className="space-y-3">
          <p className="text-xs text-gray-400 font-medium">Deploy with canary rollout:</p>
          <CodeBlock code="fleetml deploy model.onnx --fleet default --canary 5,50,100" />
          <p className="text-xs text-gray-400 font-medium">Check status:</p>
          <CodeBlock code="fleetml status --fleet default" />
          <p className="text-xs text-gray-400">
            Don't have an ONNX model yet? Export from PyTorch with torch.onnx.export() or from TensorFlow with tf2onnx.
          </p>
        </div>
      ),
    },
  ];

  return (
    <div className="max-w-3xl">
      <div className="mb-8">
        <h2 className="text-2xl font-bold text-gray-900">
          Welcome{user?.name ? `, ${user.name}` : ''}!
        </h2>
        <p className="text-gray-500 mt-1">
          Follow these steps to deploy your first model to an edge device.
        </p>
      </div>

      {/* Progress */}
      <div className="flex items-center gap-2 mb-8">
        {steps.map((s, i) => (
          <div key={s.id} className="flex items-center gap-2">
            <div
              className={`w-8 h-8 rounded-full flex items-center justify-center text-sm font-bold ${
                completedSteps.has(s.id)
                  ? 'bg-green-500 text-white'
                  : 'bg-gray-200 text-gray-500'
              }`}
            >
              {completedSteps.has(s.id) ? (
                <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2.5} d="M5 13l4 4L19 7" />
                </svg>
              ) : (
                i + 1
              )}
            </div>
            {i < steps.length - 1 && (
              <div className={`w-8 h-0.5 ${completedSteps.has(s.id) ? 'bg-green-400' : 'bg-gray-200'}`} />
            )}
          </div>
        ))}
        <span className="ml-3 text-sm text-gray-400">
          {completedSteps.size}/{steps.length} complete
        </span>
      </div>

      {/* Steps */}
      <div className="space-y-6">
        {steps.map((step, i) => {
          const done = completedSteps.has(step.id);
          return (
            <div
              key={step.id}
              className={`bg-white rounded-lg border p-6 transition-all ${
                done ? 'border-green-200 bg-green-50/30' : 'border-gray-200'
              }`}
            >
              <div className="flex items-start justify-between mb-3">
                <div>
                  <h3 className="text-lg font-semibold text-gray-900 flex items-center gap-2">
                    <span className="text-sm text-gray-400">Step {i + 1}</span>
                    {step.title}
                  </h3>
                  <p className="text-sm text-gray-500 mt-1">{step.description}</p>
                </div>
                <button
                  onClick={() => toggleStep(step.id)}
                  className={`flex-shrink-0 ml-4 px-3 py-1 rounded-full text-xs font-medium transition-colors ${
                    done
                      ? 'bg-green-100 text-green-700 hover:bg-green-200'
                      : 'bg-gray-100 text-gray-500 hover:bg-gray-200'
                  }`}
                >
                  {done ? 'Done' : 'Mark done'}
                </button>
              </div>

              {step.content}
            </div>
          );
        })}
      </div>

      {/* Help */}
      <div className="mt-8 p-4 bg-blue-50 border border-blue-100 rounded-lg">
        <p className="text-sm text-blue-800">
          <strong>Need help?</strong> If you run into any issues, check the{' '}
          <a
            href="https://github.com/ashish-frozo/fleetML/issues"
            target="_blank"
            rel="noopener noreferrer"
            className="underline hover:no-underline"
          >
            GitHub Issues
          </a>{' '}
          page or email support@fleetml.dev.
        </p>
      </div>
    </div>
  );
}
