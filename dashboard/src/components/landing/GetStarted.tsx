import { Link } from 'react-router-dom';

const steps = [
  {
    step: '1',
    title: 'Create your account',
    description: 'Sign up for free. No credit card needed. You\'ll get an API key to connect your devices.',
    code: null,
    cta: { label: 'Create Free Account', to: '/signup' },
  },
  {
    step: '2',
    title: 'Install the CLI',
    description: 'One command installs the FleetML CLI on macOS or Linux.',
    code: 'curl -sSL https://raw.githubusercontent.com/ashish-frozo/fleetML/main/scripts/install.sh | bash',
  },
  {
    step: '3',
    title: 'Connect the CLI to your account',
    description: 'Authenticate with the email and password you signed up with.',
    code: 'fleetml init --cloud',
  },
  {
    step: '4',
    title: 'Copy your API key from the dashboard',
    description: 'Your API key links devices to your account. Find it in the Get Started page after logging in.',
    code: null,
    note: 'Your API key looks like: flml_a1b2c3d4e5f6... — keep it secret.',
  },
  {
    step: '5',
    title: 'Install the agent on your edge device',
    description: 'SSH into your device (Jetson, RPi, Intel NUC, etc.) and run the installer, then start the agent with your API key.',
    code: 'curl -sSL https://raw.githubusercontent.com/ashish-frozo/fleetML/main/scripts/install-agent.sh | sh',
    code2: 'FLEETML_API_KEY=flml_your_key FLEETML_SERVER=your-server:50051 fleetml-agent',
  },
  {
    step: '6',
    title: 'Deploy your first model',
    description: 'Push an ONNX model to your fleet with canary rollout. FleetML handles compilation and distribution.',
    code: 'fleetml deploy model.onnx --fleet default --canary 5,50,100',
  },
];

export default function GetStarted() {
  return (
    <section id="get-started" className="py-24 border-t border-white/5">
      <div className="max-w-3xl mx-auto px-4 sm:px-6">
        <div className="text-center mb-16 scroll-reveal">
          <h2 className="text-3xl sm:text-4xl font-bold text-white mb-4">
            Get started in 5 minutes
          </h2>
          <p className="text-gray-400 max-w-2xl mx-auto">
            From zero to your first edge deployment &mdash; no infrastructure setup required.
          </p>
        </div>

        <div className="space-y-6">
          {steps.map((s, i) => (
            <div
              key={s.step}
              className="scroll-reveal flex gap-4"
              style={{ transitionDelay: `${i * 80}ms` }}
            >
              {/* Step number */}
              <div className="flex-shrink-0 w-8 h-8 rounded-full bg-gradient-to-br from-blue-500/20 to-purple-500/20 border border-white/10 flex items-center justify-center text-sm font-bold text-gray-400">
                {s.step}
              </div>

              <div className="flex-1 min-w-0">
                <h3 className="text-base font-semibold text-white mb-1">{s.title}</h3>
                <p className="text-gray-500 text-sm mb-3">{s.description}</p>

                {s.code && (
                  <div className="rounded-lg overflow-hidden border border-white/10 bg-gray-950">
                    <div className="px-4 py-3 font-mono text-sm text-gray-300 overflow-x-auto">
                      <span className="text-gray-500 select-none">$ </span>
                      {s.code}
                    </div>
                  </div>
                )}

                {(s as any).code2 && (
                  <div className="rounded-lg overflow-hidden border border-white/10 bg-gray-950 mt-2">
                    <div className="px-4 py-3 font-mono text-sm text-gray-300 overflow-x-auto">
                      <span className="text-gray-500 select-none">$ </span>
                      {(s as any).code2}
                    </div>
                  </div>
                )}

                {(s as any).note && (
                  <p className="text-xs text-gray-600 mt-2 italic">{(s as any).note}</p>
                )}

                {s.cta && (
                  <Link
                    to={s.cta.to}
                    className="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-gradient-to-r from-blue-600 to-purple-600 text-white text-sm font-medium hover:from-blue-500 hover:to-purple-500 transition-all"
                  >
                    {s.cta.label}
                    <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 7l5 5m0 0l-5 5m5-5H6" />
                    </svg>
                  </Link>
                )}
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
