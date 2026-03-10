import { Link } from 'react-router-dom';

export default function OpenSourceCTA() {
  return (
    <section className="py-24 border-t border-white/5 relative overflow-hidden">
      {/* Background glow */}
      <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[600px] h-[600px] bg-gradient-radial from-purple-500/10 via-transparent to-transparent rounded-full blur-3xl" />

      <div className="relative max-w-3xl mx-auto px-4 sm:px-6 text-center">
        <div className="scroll-reveal">
          <h2 className="text-3xl sm:text-4xl font-bold text-white mb-4">
            Your models deserve a better deployment story.
          </h2>
          <p className="text-gray-400 mb-8 max-w-xl mx-auto">
            Go from training to production edge deployment in under 5 minutes. Free tier includes 5 devices &mdash; no credit card required.
          </p>

          <Link
            to="/signup"
            className="btn-primary text-base"
          >
            Start Free &mdash; 5 Devices Included
            <svg className="w-4 h-4 ml-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 7l5 5m0 0l-5 5m5-5H6" />
            </svg>
          </Link>
        </div>
      </div>
    </section>
  );
}
