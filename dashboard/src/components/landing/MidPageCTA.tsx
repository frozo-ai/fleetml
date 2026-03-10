import { Link } from 'react-router-dom';

export default function MidPageCTA() {
  return (
    <section className="py-12">
      <div className="max-w-3xl mx-auto px-4 sm:px-6 text-center scroll-reveal">
        <p className="text-lg text-gray-300 font-medium mb-5">
          Ready to stop SSH-ing into devices?
        </p>
        <div className="flex flex-col sm:flex-row items-center justify-center gap-3">
          <Link to="/signup" className="btn-primary text-sm">
            Start Free &mdash; 5 Devices
            <svg className="w-4 h-4 ml-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 7l5 5m0 0l-5 5m5-5H6" />
            </svg>
          </Link>
          <a href="#pricing" className="btn-outline text-sm">
            View Pricing
          </a>
        </div>
      </div>
    </section>
  );
}
