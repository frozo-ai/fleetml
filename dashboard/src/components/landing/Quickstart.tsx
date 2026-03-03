export default function Quickstart() {
  return (
    <section id="quickstart" className="py-24 border-t border-white/5 relative overflow-hidden">
      <div className="absolute top-0 left-1/2 -translate-x-1/2 w-[600px] h-[300px] bg-gradient-radial from-blue-500/5 via-transparent to-transparent rounded-full blur-3xl" />

      <div className="relative max-w-3xl mx-auto px-4 sm:px-6">
        <div className="text-center mb-12 scroll-reveal">
          <h2 className="text-3xl sm:text-4xl font-bold text-white mb-4">
            Deploy your first model in 5 minutes
          </h2>
          <p className="text-gray-400 max-w-xl mx-auto">
            Three commands. One model on one device. Then scale to your entire fleet.
          </p>
        </div>

        <div className="scroll-reveal space-y-4">
          {/* Step 1 */}
          <div className="rounded-xl overflow-hidden border border-white/10">
            <div className="px-4 py-2.5 bg-white/5 border-b border-white/5 flex items-center gap-3">
              <span className="w-6 h-6 rounded-full bg-blue-500/20 border border-blue-500/30 flex items-center justify-center text-xs font-bold text-blue-400">1</span>
              <span className="text-sm text-gray-300 font-medium">Install the CLI</span>
            </div>
            <pre className="p-4 bg-gray-950 font-mono text-sm overflow-x-auto">
              <code>
                <span className="text-gray-500">$</span>{' '}
                <span className="text-gray-300">curl -sSL https://get.fleetml.io | sh</span>
              </code>
            </pre>
          </div>

          {/* Step 2 */}
          <div className="rounded-xl overflow-hidden border border-white/10">
            <div className="px-4 py-2.5 bg-white/5 border-b border-white/5 flex items-center gap-3">
              <span className="w-6 h-6 rounded-full bg-purple-500/20 border border-purple-500/30 flex items-center justify-center text-xs font-bold text-purple-400">2</span>
              <span className="text-sm text-gray-300 font-medium">Deploy your first model</span>
            </div>
            <pre className="p-4 bg-gray-950 font-mono text-sm overflow-x-auto">
              <code>
                <span className="text-gray-500">$</span>{' '}
                <span className="text-gray-300">fleetml deploy model.onnx --device local</span>
              </code>
            </pre>
          </div>

          {/* Step 3 */}
          <div className="rounded-xl overflow-hidden border border-white/10">
            <div className="px-4 py-2.5 bg-white/5 border-b border-white/5 flex items-center gap-3">
              <span className="w-6 h-6 rounded-full bg-cyan-500/20 border border-cyan-500/30 flex items-center justify-center text-xs font-bold text-cyan-400">3</span>
              <span className="text-sm text-gray-300 font-medium">Now scale to your fleet</span>
            </div>
            <pre className="p-4 bg-gray-950 font-mono text-sm overflow-x-auto">
              <code>
                <span className="text-gray-500">$</span>{' '}
                <span className="text-gray-300">fleetml deploy model.onnx --fleet production</span>
              </code>
            </pre>
          </div>
        </div>

        <div className="text-center mt-10 scroll-reveal">
          <a
            href="https://github.com/fleetml/fleetml"
            target="_blank"
            rel="noopener noreferrer"
            className="btn-primary text-base"
          >
            View Full Quickstart Guide
            <svg className="w-4 h-4 ml-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 7l5 5m0 0l-5 5m5-5H6" />
            </svg>
          </a>
        </div>
      </div>
    </section>
  );
}
