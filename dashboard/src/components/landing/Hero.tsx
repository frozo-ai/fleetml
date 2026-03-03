import { useState, useEffect, useRef } from 'react';

const terminalLines = [
  { text: '$ fleetml deploy model.onnx --fleet production', isCommand: true },
  { text: '\u2713 Model uploaded (SHA-256 verified)', isCommand: false },
  { text: '\u2713 Deploying to 47 devices...', isCommand: false },
  { text: '\u2713 Canary 5% \u2192 50% \u2192 100% complete', isCommand: false },
  { text: '\u2713 Zero-downtime hot-swap successful', isCommand: false },
];

function Terminal() {
  const [visibleLines, setVisibleLines] = useState<number>(0);
  const [typedChars, setTypedChars] = useState<number>(0);
  const [started, setStarted] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const el = ref.current;
    if (!el) return;
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting && !started) setStarted(true);
      },
      { threshold: 0.3 }
    );
    observer.observe(el);
    return () => observer.disconnect();
  }, [started]);

  useEffect(() => {
    if (!started) return;
    if (visibleLines >= terminalLines.length) return;

    const currentLine = terminalLines[visibleLines];
    const speed = currentLine.isCommand ? 40 : 10;

    if (typedChars < currentLine.text.length) {
      const timer = setTimeout(() => setTypedChars((c) => c + 1), speed);
      return () => clearTimeout(timer);
    } else {
      const delay = currentLine.isCommand ? 600 : 300;
      const timer = setTimeout(() => {
        setVisibleLines((l) => l + 1);
        setTypedChars(0);
      }, delay);
      return () => clearTimeout(timer);
    }
  }, [started, visibleLines, typedChars]);

  return (
    <div ref={ref} className="max-w-3xl mx-auto mt-14 animate-fade-in-up" style={{ animationDelay: '0.4s' }}>
      <div className="rounded-xl overflow-hidden border border-white/10 shadow-2xl glow">
        <div className="flex items-center gap-2 px-4 py-3 bg-gray-900/80 border-b border-white/5">
          <div className="w-3 h-3 rounded-full bg-red-500/80" />
          <div className="w-3 h-3 rounded-full bg-yellow-500/80" />
          <div className="w-3 h-3 rounded-full bg-green-500/80" />
          <span className="ml-2 text-xs text-gray-500 font-medium">terminal</span>
        </div>
        <div className="bg-gray-950 p-5 font-mono text-sm leading-relaxed min-h-[180px]">
          {terminalLines.map((line, i) => {
            if (i > visibleLines) return null;
            const isCurrentLine = i === visibleLines;
            const displayText = isCurrentLine
              ? line.text.slice(0, typedChars)
              : line.text;
            return (
              <div key={i} className="flex items-start">
                <span
                  className={
                    line.isCommand
                      ? 'text-gray-300'
                      : line.text.startsWith('\u2713')
                        ? 'text-green-400'
                        : 'text-gray-400'
                  }
                >
                  {displayText}
                </span>
                {isCurrentLine && (
                  <span className="inline-block w-2 h-5 bg-gray-300 ml-0.5 animate-typing-cursor" />
                )}
              </div>
            );
          })}
          {visibleLines >= terminalLines.length && (
            <div className="flex items-start mt-0">
              <span className="text-gray-300">$ </span>
              <span className="inline-block w-2 h-5 bg-gray-300 ml-0.5 animate-typing-cursor" />
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

export default function Hero() {
  return (
    <section className="relative min-h-screen flex items-center justify-center overflow-hidden">
      {/* Background */}
      <div className="absolute inset-0 bg-gray-950" />
      <div className="absolute inset-0 dot-grid" />
      <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[800px] h-[800px] bg-gradient-radial from-purple-500/10 via-transparent to-transparent rounded-full blur-3xl" />

      <div className="relative z-10 max-w-5xl mx-auto px-4 sm:px-6 text-center pt-24 pb-16">
        {/* Badge */}
        <div className="inline-flex items-center gap-2 px-4 py-1.5 rounded-full border border-white/10 bg-white/5 text-sm text-gray-400 mb-8 animate-fade-in-up">
          <span className="w-2 h-2 rounded-full bg-green-400 animate-pulse" />
          Open Source &mdash; Apache 2.0
        </div>

        {/* Headline — pain-first */}
        <h1 className="text-4xl sm:text-5xl md:text-7xl font-extrabold text-white leading-tight tracking-tight mb-6 animate-fade-in-up" style={{ animationDelay: '0.1s' }}>
          Stop SSH-ing Into Every{' '}
          <span className="gradient-text">Edge Device</span>
          {' '}to Update Your Models
        </h1>

        {/* Subtitle — benefit-led */}
        <p className="text-lg sm:text-xl text-gray-400 max-w-2xl mx-auto mb-10 animate-fade-in-up" style={{ animationDelay: '0.2s' }}>
          One command deploys to your entire fleet &mdash; Jetson, Raspberry Pi, Intel, Coral, any chip.
          Canary rollouts, zero-downtime hot-swap, offline-first. Open source.
        </p>

        {/* CTAs */}
        <div className="flex flex-col sm:flex-row items-center justify-center gap-4 animate-fade-in-up" style={{ animationDelay: '0.3s' }}>
          <a
            href="#quickstart"
            className="btn-primary text-base"
          >
            Try the Quickstart
            <svg className="w-4 h-4 ml-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 7l5 5m0 0l-5 5m5-5H6" />
            </svg>
          </a>
          <a
            href="https://github.com/fleetml/fleetml"
            target="_blank"
            rel="noopener noreferrer"
            className="btn-outline text-base"
          >
            <svg className="w-5 h-5 mr-2" fill="currentColor" viewBox="0 0 24 24">
              <path fillRule="evenodd" d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z" clipRule="evenodd" />
            </svg>
            Star on GitHub
          </a>
        </div>

        {/* Terminal — embedded in hero as proof */}
        <Terminal />
      </div>
    </section>
  );
}
