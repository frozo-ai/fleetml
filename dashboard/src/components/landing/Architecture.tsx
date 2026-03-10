export default function Architecture() {
  return (
    <section id="architecture" className="py-24 border-t border-white/5">
      <div className="max-w-7xl mx-auto px-4 sm:px-6">
        <div className="text-center mb-16 scroll-reveal">
          <h2 className="text-3xl sm:text-4xl font-bold text-white mb-4">
            Three-tier architecture
          </h2>
          <p className="text-gray-400 max-w-2xl mx-auto">
            Control plane manages fleet state, gRPC + MQTT handle communication, agents run autonomously on edge devices.
          </p>
        </div>

        <div className="scroll-reveal">
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 lg:gap-0 items-start relative">
            {/* Control Plane */}
            <div className="glass p-6 lg:rounded-r-none">
              <div className="flex items-center gap-3 mb-4">
                <div className="w-10 h-10 rounded-lg bg-blue-500/20 flex items-center justify-center">
                  <svg className="w-5 h-5 text-blue-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2" />
                  </svg>
                </div>
                <h3 className="text-lg font-semibold text-white">Control Plane</h3>
              </div>
              <div className="space-y-2">
                {['REST & gRPC APIs', 'Fleet Management', 'Model Registry', 'Deployment Orchestrator', 'Monitoring & Alerts'].map((item) => (
                  <div key={item} className="flex items-center gap-2 text-sm text-gray-400">
                    <span className="w-1.5 h-1.5 rounded-full bg-blue-400/60" />
                    {item}
                  </div>
                ))}
              </div>
            </div>

            {/* Network Layer */}
            <div className="glass p-6 lg:rounded-none border-x-0 lg:border-x lg:border-white/10 relative">
              {/* Connection indicators */}
              <div className="hidden lg:block absolute left-0 top-1/2 -translate-x-1/2 -translate-y-1/2">
                <div className="w-3 h-3 rounded-full bg-purple-500 animate-pulse" />
              </div>
              <div className="hidden lg:block absolute right-0 top-1/2 translate-x-1/2 -translate-y-1/2">
                <div className="w-3 h-3 rounded-full bg-purple-500 animate-pulse" />
              </div>

              <div className="flex items-center gap-3 mb-4">
                <div className="w-10 h-10 rounded-lg bg-purple-500/20 flex items-center justify-center">
                  <svg className="w-5 h-5 text-purple-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M8.288 15.038a5.25 5.25 0 017.424 0M5.106 11.856c3.807-3.808 9.98-3.808 13.788 0M1.924 8.674c5.565-5.565 14.587-5.565 20.152 0M12 18.75h.008v.008H12v-.008z" />
                  </svg>
                </div>
                <h3 className="text-lg font-semibold text-white">Network</h3>
              </div>
              <div className="space-y-2">
                {['Encrypted by default', 'Auto-failover protocols', 'Store & forward offline', 'Sub-second latency', 'Works behind firewalls'].map((item) => (
                  <div key={item} className="flex items-center gap-2 text-sm text-gray-400">
                    <span className="w-1.5 h-1.5 rounded-full bg-purple-400/60" />
                    {item}
                  </div>
                ))}
              </div>

              {/* Animated dashes between columns (mobile: vertical, desktop: hidden — handled by dots) */}
              <div className="lg:hidden flex justify-center my-4">
                <svg width="2" height="40">
                  <line x1="1" y1="0" x2="1" y2="40" stroke="rgba(168,85,247,0.4)" strokeWidth="2" strokeDasharray="4 4" className="animate-flow-dash" />
                </svg>
              </div>
            </div>

            {/* Edge Fleet */}
            <div className="glass p-6 lg:rounded-l-none">
              <div className="flex items-center gap-3 mb-4">
                <div className="w-10 h-10 rounded-lg bg-cyan-500/20 flex items-center justify-center">
                  <svg className="w-5 h-5 text-cyan-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M9 17.25v1.007a3 3 0 01-.879 2.122L7.5 21h9l-.621-.621A3 3 0 0115 18.257V17.25m6-12V15a2.25 2.25 0 01-2.25 2.25H5.25A2.25 2.25 0 013 15V5.25m18 0A2.25 2.25 0 0018.75 3H5.25A2.25 2.25 0 003 5.25m18 0V12a2.25 2.25 0 01-2.25 2.25H5.25A2.25 2.25 0 013 12V5.25" />
                  </svg>
                </div>
                <h3 className="text-lg font-semibold text-white">Edge Fleet</h3>
              </div>
              <div className="space-y-2">
                {['Lightweight agent (~15MB)', 'Auto chip detection', 'Runs fully offline', 'Health monitoring', 'Zero-downtime updates'].map((item) => (
                  <div key={item} className="flex items-center gap-2 text-sm text-gray-400">
                    <span className="w-1.5 h-1.5 rounded-full bg-cyan-400/60" />
                    {item}
                  </div>
                ))}
              </div>
            </div>
          </div>

          {/* Flow arrows (desktop only) */}
          <div className="hidden lg:flex justify-center mt-6 gap-2">
            <svg width="100%" height="24" className="max-w-3xl">
              <defs>
                <linearGradient id="arrowGrad" x1="0%" y1="0%" x2="100%" y2="0%">
                  <stop offset="0%" stopColor="#3b82f6" stopOpacity="0.6" />
                  <stop offset="50%" stopColor="#8b5cf6" stopOpacity="0.6" />
                  <stop offset="100%" stopColor="#22d3ee" stopOpacity="0.6" />
                </linearGradient>
              </defs>
              <line x1="10%" y1="12" x2="90%" y2="12" stroke="url(#arrowGrad)" strokeWidth="2" strokeDasharray="6 4" className="animate-flow-dash" />
            </svg>
          </div>
        </div>
      </div>
    </section>
  );
}
