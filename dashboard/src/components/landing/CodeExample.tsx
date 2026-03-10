export default function CodeExample() {
  return (
    <section className="py-24 border-t border-white/5">
      <div className="max-w-7xl mx-auto px-4 sm:px-6">
        <div className="text-center mb-16 scroll-reveal">
          <h2 className="text-3xl sm:text-4xl font-bold text-white mb-4">
            Deploy in one command, configure in ten lines
          </h2>
          <p className="text-gray-400 max-w-2xl mx-auto">
            A familiar CLI and declarative YAML. No PhD required.
          </p>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 scroll-reveal">
          {/* CLI Example */}
          <div className="rounded-xl overflow-hidden border border-white/10">
            <div className="px-4 py-2.5 bg-white/5 border-b border-white/5 flex items-center justify-between">
              <span className="text-xs text-gray-500 font-medium">CLI — Deploy Command</span>
              <span className="text-xs text-gray-600">bash</span>
            </div>
            <pre className="p-5 bg-gray-950 text-sm font-mono leading-relaxed overflow-x-auto">
              <code>
                <span className="code-comment"># Deploy a model to your production fleet</span>{'\n'}
                <span className="text-gray-300">$ </span>
                <span className="code-keyword">fleetml deploy</span>
                <span className="text-gray-300"> resnet50.onnx </span>
                <span className="code-flag">\</span>{'\n'}
                <span className="code-flag">  --fleet</span>
                <span className="code-string"> production</span>
                <span className="code-flag"> \</span>{'\n'}
                <span className="code-flag">  --canary</span>
                <span className="code-value"> 5,50,100</span>
                <span className="code-flag"> \</span>{'\n'}
                <span className="code-flag">  --rollback-on</span>
                <span className="code-value"> error-rate&gt;0.01</span>
                <span className="code-flag"> \</span>{'\n'}
                <span className="code-flag">  --timeout</span>
                <span className="code-value"> 300s</span>{'\n'}
                {'\n'}
                <span className="code-comment"># Check deployment status</span>{'\n'}
                <span className="text-gray-300">$ </span>
                <span className="code-keyword">fleetml status</span>
                <span className="code-flag"> --fleet</span>
                <span className="code-string"> production</span>
              </code>
            </pre>
          </div>

          {/* YAML Config */}
          <div className="rounded-xl overflow-hidden border border-white/10">
            <div className="px-4 py-2.5 bg-white/5 border-b border-white/5 flex items-center justify-between">
              <span className="text-xs text-gray-500 font-medium">Agent Config</span>
              <span className="text-xs text-gray-600">yaml</span>
            </div>
            <pre className="p-5 bg-gray-950 text-sm font-mono leading-relaxed overflow-x-auto">
              <code>
                <span className="code-comment"># fleetml-agent.yaml</span>{'\n'}
                <span className="code-keyword">server</span><span className="text-gray-500">:</span>{'\n'}
                <span className="code-flag">  endpoint</span><span className="text-gray-500">:</span>
                <span className="code-string"> "grpc://your-server:50051"</span>{'\n'}
                <span className="code-flag">  tls</span><span className="text-gray-500">:</span>
                <span className="code-value"> true</span>{'\n'}
                {'\n'}
                <span className="code-keyword">agent</span><span className="text-gray-500">:</span>{'\n'}
                <span className="code-flag">  id</span><span className="text-gray-500">:</span>
                <span className="code-string"> "edge-node-047"</span>{'\n'}
                <span className="code-flag">  fleet</span><span className="text-gray-500">:</span>
                <span className="code-string"> "production"</span>{'\n'}
                <span className="code-flag">  heartbeat_interval</span><span className="text-gray-500">:</span>
                <span className="code-value"> 30s</span>{'\n'}
                {'\n'}
                <span className="code-keyword">runtime</span><span className="text-gray-500">:</span>{'\n'}
                <span className="code-flag">  backend</span><span className="text-gray-500">:</span>
                <span className="code-string"> "onnxruntime"</span>{'\n'}
                <span className="code-flag">  max_models</span><span className="text-gray-500">:</span>
                <span className="code-value"> 3</span>
              </code>
            </pre>
          </div>
        </div>
      </div>
    </section>
  );
}
