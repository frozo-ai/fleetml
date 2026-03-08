import { Link } from 'react-router-dom';

export default function InlineEmailCapture() {
  return (
    <section className="py-16">
      <div className="max-w-2xl mx-auto px-4 sm:px-6">
        <div className="scroll-reveal glass p-8 text-center">
          <p className="text-gray-300 font-medium mb-1">
            Your models deserve a better deployment story.
          </p>
          <p className="text-gray-500 text-sm mb-5">
            Get started in under 5 minutes. Free tier includes 5 devices.
          </p>
          <Link
            to="/signup"
            className="btn-primary text-sm inline-flex items-center"
          >
            Create Free Account
            <svg className="w-4 h-4 ml-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 7l5 5m0 0l-5 5m5-5H6" />
            </svg>
          </Link>
          <p className="text-xs text-gray-600 mt-3">No credit card required.</p>
        </div>
      </div>
    </section>
  );
}
