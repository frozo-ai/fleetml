import { useState } from 'react';

export default function InlineEmailCapture() {
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
    <section className="py-16">
      <div className="max-w-2xl mx-auto px-4 sm:px-6">
        <div className="scroll-reveal glass p-8 text-center">
          {submitted ? (
            <p className="text-green-400 text-sm">You&rsquo;re on the list. We&rsquo;ll let you know when FleetML launches.</p>
          ) : (
            <>
              <p className="text-gray-300 font-medium mb-1">
                Convinced? We&rsquo;re almost ready.
              </p>
              <p className="text-gray-500 text-sm mb-5">
                Get the launch announcement and early access.
              </p>
              <form onSubmit={handleSubmit} className="flex flex-col sm:flex-row gap-2 max-w-md mx-auto">
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
                  Join the Waitlist
                </button>
              </form>
              <p className="text-xs text-gray-600 mt-3">No spam. Just the launch announcement.</p>
            </>
          )}
        </div>
      </div>
    </section>
  );
}
