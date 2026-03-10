import { useState, useEffect, useRef } from 'react';
import { Link } from 'react-router-dom';

const terminalLines = [
  { text: '$ fleetml deploy defect-detector.onnx --fleet production --canary 5,50,100', isCommand: true },
  { text: '\u2713 Model uploaded (SHA-256 verified)', isCommand: false },
  { text: '\u2713 Compiled: TensorRT (Jetson), TFLite (RPi), OpenVINO (Intel)', isCommand: false },
  { text: '\u2713 Canary 5% \u2192 12 devices healthy', isCommand: false },
  { text: '\u2713 Canary 50% \u2192 120 devices healthy', isCommand: false },
  { text: '\u2713 Rolling out to 100%... 240/240 complete', isCommand: false },
  { text: '\u2713 Zero-downtime hot-swap successful. 0 dropped inferences.', isCommand: false },
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
    const speed = currentLine.isCommand ? 25 : 8;

    if (typedChars < currentLine.text.length) {
      const timer = setTimeout(() => setTypedChars((c) => c + 1), speed);
      return () => clearTimeout(timer);
    } else {
      const delay = currentLine.isCommand ? 800 : 250;
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
          <span className="ml-2 text-xs text-gray-500 font-medium">terminal &mdash; 240 devices, 3 chip types</span>
        </div>
        <div className="bg-gray-950 p-5 font-mono text-sm leading-relaxed min-h-[220px]">
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
          Open Source &mdash; Apache 2.0 &bull; Free tier available
        </div>

        {/* Headline */}
        <h1 className="text-4xl sm:text-5xl md:text-7xl font-extrabold text-white leading-tight tracking-tight mb-6 animate-fade-in-up" style={{ animationDelay: '0.1s' }}>
          Deploy AI Models to{' '}
          <span className="gradient-text">Edge Devices.</span>
          <br />
          One Command.
        </h1>

        {/* Subtitle */}
        <p className="text-lg sm:text-xl text-gray-400 max-w-2xl mx-auto mb-6 animate-fade-in-up" style={{ animationDelay: '0.2s' }}>
          FleetML ships your models to hundreds of devices across any chip type &mdash; with canary rollouts,
          zero-downtime updates, and offline-first reliability. Stop SSH-ing into devices. Start deploying.
        </p>

        {/* AI-extractable definition */}
        <p className="text-sm text-gray-500 max-w-2xl mx-auto mb-10 animate-fade-in-up" style={{ animationDelay: '0.25s' }}>
          FleetML is an open-source, chip-neutral edge MLOps platform that compiles one ONNX model
          for every chip type automatically, deploys with canary rollouts and auto-rollback, and
          operates offline-first so devices keep running without connectivity.
        </p>

        {/* CTAs */}
        <div className="flex flex-col sm:flex-row items-center justify-center gap-4 animate-fade-in-up" style={{ animationDelay: '0.3s' }}>
          <Link
            to="/signup"
            className="btn-primary text-base"
          >
            Start Free &mdash; 5 Devices Included
            <svg className="w-4 h-4 ml-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 7l5 5m0 0l-5 5m5-5H6" />
            </svg>
          </Link>
          <a
            href="#how-it-works"
            className="btn-outline text-base"
          >
            See How It Works
            <svg className="w-4 h-4 ml-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 14l-7 7m0 0l-7-7m7 7V3" />
            </svg>
          </a>
        </div>

        {/* Terminal */}
        <Terminal />
      </div>
    </section>
  );
}
