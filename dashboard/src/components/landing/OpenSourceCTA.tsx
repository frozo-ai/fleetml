import { useState } from 'react';

export default function OpenSourceCTA() {
  const [email, setEmail] = useState('');
  const [submitted, setSubmitted] = useState(false);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (email.trim()) {
      setSubmitted(true);
      setEmail('');
    }
  };

  return (
    <section className="py-24 border-t border-white/5 relative overflow-hidden">
      {/* Background glow */}
      <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[600px] h-[600px] bg-gradient-radial from-purple-500/10 via-transparent to-transparent rounded-full blur-3xl" />

      <div className="relative max-w-3xl mx-auto px-4 sm:px-6 text-center">
        <div className="scroll-reveal">
          <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full border border-white/10 bg-white/5 text-xs text-gray-400 mb-6">
            Apache 2.0 License
          </div>

          <h2 className="text-3xl sm:text-4xl font-bold text-white mb-4">
            Built in the open
          </h2>
          <p className="text-gray-400 mb-8 max-w-xl mx-auto">
            Join the developers building the future of edge ML. Contribute code, file issues, or fork it and make it your own.
          </p>

          <div className="flex flex-col sm:flex-row items-center justify-center gap-4 mb-12">
            <a
              href="https://github.com/fleetml/fleetml"
              target="_blank"
              rel="noopener noreferrer"
              className="btn-primary text-base"
            >
              <svg className="w-5 h-5 mr-2" fill="currentColor" viewBox="0 0 24 24">
                <path fillRule="evenodd" d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z" clipRule="evenodd" />
              </svg>
              Star on GitHub
            </a>
            <a
              href="https://github.com/fleetml/fleetml#readme"
              target="_blank"
              rel="noopener noreferrer"
              className="btn-outline text-base"
            >
              Read the Docs
            </a>
          </div>

          {/* Email capture */}
          <div className="glass p-6 max-w-md mx-auto">
            <p className="text-sm text-gray-300 font-medium mb-3">
              Get notified when we launch
            </p>
            {submitted ? (
              <p className="text-green-400 text-sm">You're on the list. We'll be in touch.</p>
            ) : (
              <form onSubmit={handleSubmit} className="flex gap-2">
                <input
                  type="email"
                  required
                  placeholder="you@company.com"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  className="flex-1 px-4 py-2.5 rounded-lg bg-white/5 border border-white/10 text-white text-sm placeholder-gray-500 focus:outline-none focus:border-purple-500/50 focus:ring-1 focus:ring-purple-500/30"
                />
                <button
                  type="submit"
                  className="btn-primary text-sm !px-5 !py-2.5 shrink-0"
                >
                  Notify Me
                </button>
              </form>
            )}
            <p className="text-xs text-gray-600 mt-2">No spam. Just the launch announcement.</p>
          </div>
        </div>
      </div>
    </section>
  );
}
