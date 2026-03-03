const features = [
  {
    icon: '\u26A1',
    title: 'Zero-Downtime Hot-Swap',
    description: 'FleetML swaps models on running devices without losing a single inference. Your fleet keeps serving while updates roll out.',
  },
  {
    icon: '\uD83D\uDCE1',
    title: 'Offline-First',
    description: 'FleetML devices keep running when the network goes down. All data syncs automatically when connectivity returns.',
  },
  {
    icon: '\uD83D\uDEE1\uFE0F',
    title: 'Canary Deployments',
    description: 'Roll out to 5% of your fleet first, then 50%, then 100%. Automatic rollback if error rates spike.',
  },
  {
    icon: '\uD83E\uDDE9',
    title: 'Chip-Neutral',
    description: 'Upload one ONNX model and FleetML compiles it for each device\u2019s chip automatically \u2014 TensorRT, OpenVINO, TFLite, and more.',
  },
  {
    icon: '\uD83D\uDCCA',
    title: 'Fleet Monitoring',
    description: 'CPU, GPU, memory, temperature, and inference metrics from every device in one dashboard. Know what\u2019s running where.',
  },
  {
    icon: '\u2328\uFE0F',
    title: 'CLI-First Workflow',
    description: (<>Add <code className="px-1.5 py-0.5 rounded bg-white/10 text-gray-300 font-mono text-xs">fleetml deploy</code> to your CI/CD pipeline. One line in your GitHub Action, and your fleet stays updated.</>),
  },
];

export default function Features() {
  return (
    <section id="features" className="py-24 relative">
      <div className="max-w-7xl mx-auto px-4 sm:px-6">
        <div className="text-center mb-16 scroll-reveal">
          <h2 className="text-3xl sm:text-4xl font-bold text-white mb-4">
            Everything you need for edge ML
          </h2>
          <p className="text-gray-400 max-w-2xl mx-auto">
            Built for production edge deployments &mdash; from a single Raspberry Pi to thousands of heterogeneous devices.
          </p>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
          {features.map((feature, i) => (
            <div
              key={feature.title}
              className="scroll-reveal glass p-6 hover:bg-white/[0.08] transition-all duration-300 group"
              style={{ transitionDelay: `${i * 80}ms` }}
            >
              <span className="text-3xl mb-4 block">{feature.icon}</span>
              <h3 className="text-lg font-semibold text-white mb-2 group-hover:gradient-text transition-all">
                {feature.title}
              </h3>
              <p className="text-gray-400 text-sm leading-relaxed">
                {feature.description}
              </p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
