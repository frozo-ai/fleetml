import { Link } from 'react-router-dom';

const plans = [
  {
    name: 'Free',
    price: '$0',
    period: 'forever',
    description: 'For prototyping and small projects.',
    features: [
      'Up to 5 devices',
      '1 fleet',
      'ONNX Runtime',
      'Community support',
      '3-day log retention',
    ],
    cta: 'Start Free',
    ctaLink: '/signup',
    highlighted: false,
  },
  {
    name: 'Starter',
    price: '$49',
    period: '/month',
    description: 'For teams shipping to production.',
    features: [
      'Up to 25 devices',
      '3 fleets',
      'All chip compilers',
      'Canary deployments',
      'A/B testing',
      '30-day log retention',
      'Email support',
    ],
    cta: 'Start Free Trial',
    ctaLink: '/signup',
    highlighted: true,
  },
  {
    name: 'Pro',
    price: '$199',
    period: '/month',
    description: 'For scaling edge fleets.',
    features: [
      'Up to 100 devices',
      'Unlimited fleets',
      'All chip compilers',
      'Drift detection',
      'Policy engine',
      'Priority support',
      '90-day log retention',
    ],
    cta: 'Start Free Trial',
    ctaLink: '/signup',
    highlighted: false,
  },
  {
    name: 'Enterprise',
    price: 'Custom',
    period: '',
    description: 'For large-scale deployments.',
    features: [
      'Unlimited devices',
      'Unlimited fleets',
      'Custom SLAs',
      'Dedicated support',
      'On-prem option',
      'SSO / SAML',
      'Custom retention',
    ],
    cta: 'Contact Sales',
    ctaLink: 'mailto:sales@fleetml.dev',
    highlighted: false,
  },
];

export default function Pricing() {
  return (
    <section id="pricing" className="py-24 border-t border-white/5">
      <div className="max-w-7xl mx-auto px-4 sm:px-6">
        <div className="text-center mb-16 scroll-reveal">
          <h2 className="text-3xl sm:text-4xl font-bold text-white mb-4">
            Simple, transparent pricing
          </h2>
          <p className="text-gray-400 max-w-2xl mx-auto">
            Start free. Scale as your fleet grows. No surprise fees, no per-inference charges.
          </p>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
          {plans.map((plan, i) => (
            <div
              key={plan.name}
              className={`scroll-reveal flex flex-col p-6 rounded-xl border transition-all duration-300 ${
                plan.highlighted
                  ? 'border-purple-500/40 bg-purple-500/5 ring-1 ring-purple-500/20'
                  : 'border-white/10 bg-white/[0.03] hover:bg-white/[0.06]'
              }`}
              style={{ transitionDelay: `${i * 80}ms` }}
            >
              {plan.highlighted && (
                <span className="inline-block self-start px-2.5 py-0.5 rounded text-xs font-medium bg-purple-500/20 text-purple-300 border border-purple-500/30 mb-4">
                  Most Popular
                </span>
              )}
              <h3 className="text-lg font-bold text-white">{plan.name}</h3>
              <div className="mt-2 mb-1">
                <span className="text-3xl font-extrabold text-white">{plan.price}</span>
                {plan.period && (
                  <span className="text-gray-500 text-sm ml-1">{plan.period}</span>
                )}
              </div>
              <p className="text-gray-500 text-sm mb-6">{plan.description}</p>

              <ul className="space-y-2.5 mb-8 flex-1">
                {plan.features.map((feature) => (
                  <li key={feature} className="flex items-start gap-2 text-sm text-gray-300">
                    <svg className="w-4 h-4 text-green-400 shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                    </svg>
                    {feature}
                  </li>
                ))}
              </ul>

              {plan.ctaLink.startsWith('mailto') ? (
                <a
                  href={plan.ctaLink}
                  className="btn-outline text-sm text-center w-full"
                >
                  {plan.cta}
                </a>
              ) : (
                <Link
                  to={plan.ctaLink}
                  className={`text-sm text-center w-full ${
                    plan.highlighted ? 'btn-primary' : 'btn-outline'
                  }`}
                >
                  {plan.cta}
                </Link>
              )}
            </div>
          ))}
        </div>

        <p className="text-center text-gray-600 text-xs mt-8 scroll-reveal">
          All plans include the open-source core. Self-host for free anytime &mdash; Apache 2.0 license.
        </p>
      </div>
    </section>
  );
}
