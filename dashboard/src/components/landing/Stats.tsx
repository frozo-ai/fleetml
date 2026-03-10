import { useState, useEffect, useRef } from 'react';

const stats = [
  { value: 15, prefix: '<', suffix: 'MB', label: 'Agent binary size' },
  { value: 30, prefix: '<', suffix: 's', label: 'Deploy to one device' },
  { value: 100, prefix: '', suffix: '+', label: 'Devices in under 2 min' },
  { value: 100, prefix: '', suffix: '%', label: 'Inference continuity during swap' },
];

function CountUp({ target, duration = 1500 }: { target: number; duration?: number }) {
  const [count, setCount] = useState(0);
  const [started, setStarted] = useState(false);
  const ref = useRef<HTMLSpanElement>(null);

  useEffect(() => {
    const el = ref.current;
    if (!el) return;

    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) setStarted(true);
      },
      { threshold: 0.5 }
    );
    observer.observe(el);
    return () => observer.disconnect();
  }, []);

  useEffect(() => {
    if (!started) return;

    const steps = 30;
    const increment = target / steps;
    const stepTime = duration / steps;
    let current = 0;

    const timer = setInterval(() => {
      current += increment;
      if (current >= target) {
        setCount(target);
        clearInterval(timer);
      } else {
        setCount(Math.floor(current));
      }
    }, stepTime);

    return () => clearInterval(timer);
  }, [started, target, duration]);

  return <span ref={ref}>{count}</span>;
}

export default function Stats() {
  return (
    <section className="py-20 border-t border-white/5">
      <div className="max-w-5xl mx-auto px-4 sm:px-6">
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-8 scroll-reveal">
          {stats.map((stat) => (
            <div key={stat.label} className="text-center">
              <div className="text-4xl sm:text-5xl font-extrabold text-white mb-2">
                {stat.prefix}
                <CountUp target={stat.value} />
                {stat.suffix}
              </div>
              <p className="text-sm text-gray-500">{stat.label}</p>
            </div>
          ))}
        </div>
        <p className="text-xs text-gray-700 text-center mt-6 scroll-reveal">
          Target performance specs. Measured on NVIDIA Jetson Nano + Raspberry Pi 4 test fleet.
        </p>
      </div>
    </section>
  );
}
