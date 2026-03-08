export default function Problem() {
  return (
    <section className="py-24 border-t border-white/5">
      <div className="max-w-5xl mx-auto px-4 sm:px-6">
        <div className="text-center mb-16 scroll-reveal">
          <h2 className="text-3xl sm:text-4xl font-bold text-white mb-4">
            Deploying ML to edge devices is a nightmare
          </h2>
          <p className="text-gray-400 max-w-2xl mx-auto">
            You trained the model. It works in the cloud. Now you need it running on
            hundreds of physical devices &mdash; and that&rsquo;s where everything breaks.
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          {[
            {
              icon: (
                <svg className="w-8 h-8 text-red-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
                </svg>
              ),
              title: 'SSH-ing into every device',
              description:
                'Manually copying models, restarting services, praying nothing breaks. Fine for 3 devices. Impossible at 300.',
            },
            {
              icon: (
                <svg className="w-8 h-8 text-red-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z" />
                </svg>
              ),
              title: 'Different chips, different runtimes',
              description:
                'Jetson needs TensorRT. Raspberry Pi needs TFLite. Intel needs OpenVINO. You end up maintaining 3+ deployment pipelines.',
            },
            {
              icon: (
                <svg className="w-8 h-8 text-red-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                </svg>
              ),
              title: 'Downtime and broken updates',
              description:
                'Push a bad model and your entire production line goes dark. No rollback, no canary, no safety net. Just fingers crossed.',
            },
          ].map((item, i) => (
            <div
              key={item.title}
              className="scroll-reveal glass p-6 border-red-500/10 hover:border-red-500/20 transition-all duration-300"
              style={{ transitionDelay: `${i * 100}ms` }}
            >
              <div className="mb-4">{item.icon}</div>
              <h3 className="text-lg font-semibold text-white mb-2">{item.title}</h3>
              <p className="text-gray-400 text-sm leading-relaxed">{item.description}</p>
            </div>
          ))}
        </div>

        <div className="mt-12 scroll-reveal glass p-6 text-center">
          <p className="text-gray-300 text-base">
            Teams spend <span className="text-white font-semibold">6&ndash;12 weeks</span> building
            custom deployment tooling &mdash; SSH scripts, chip-specific pipelines, monitoring dashboards.
            Then they spend every week maintaining it.
          </p>
          <p className="text-gray-500 text-sm mt-3">
            FleetML replaces all of that with one command.
          </p>
        </div>
      </div>
    </section>
  );
}
