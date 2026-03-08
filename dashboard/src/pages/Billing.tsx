import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useAuth } from '../contexts/AuthContext';
import { api } from '../api/client';

const plans = [
  {
    name: 'Free',
    key: 'free',
    price: '$0',
    period: 'forever',
    features: [
      '5 devices',
      '1 fleet',
      '3-day log retention',
      'Community support',
    ],
    cta: 'Current Plan',
  },
  {
    name: 'Starter',
    key: 'starter',
    price: '$49',
    period: '/month',
    features: [
      '25 devices',
      '3 fleets',
      '7-day log retention',
      'Email support',
      'Model compilation',
    ],
    cta: 'Upgrade',
    popular: true,
  },
  {
    name: 'Pro',
    key: 'pro',
    price: '$199',
    period: '/month',
    features: [
      '100 devices',
      'Unlimited fleets',
      '30-day log retention',
      'Priority support',
      'A/B testing',
      'Drift detection',
      'Policy engine',
    ],
    cta: 'Upgrade',
  },
  {
    name: 'Enterprise',
    key: 'enterprise',
    price: 'Custom',
    period: '',
    features: [
      'Unlimited devices',
      'Unlimited fleets',
      '90-day log retention',
      'Dedicated support + SLA',
      'SSO / SAML',
      'Audit log',
      'On-premise option',
    ],
    cta: 'Contact Sales',
  },
];

export default function BillingPage() {
  const { organization } = useAuth();
  const [upgrading, setUpgrading] = useState<string | null>(null);
  const currentPlan = organization?.plan || 'free';

  const { data: subscription } = useQuery({
    queryKey: ['subscription'],
    queryFn: () => api.getSubscription(),
  });

  const handleUpgrade = async (plan: string) => {
    if (plan === 'enterprise') {
      window.open('mailto:sales@fleetml.dev?subject=FleetML Enterprise', '_blank');
      return;
    }
    setUpgrading(plan);
    try {
      const result = await api.createCheckout(plan);
      window.location.href = result.checkout_url;
    } catch (err: any) {
      alert(err.message || 'Failed to create checkout session');
    } finally {
      setUpgrading(null);
    }
  };

  // Check for success redirect
  const params = new URLSearchParams(window.location.search);
  const isSuccess = params.get('success') === 'true';

  return (
    <div>
      <h2 className="text-2xl font-bold text-gray-900 mb-2">Billing</h2>
      <p className="text-gray-600 mb-8">
        Manage your subscription and billing.
        {organization && (
          <span className="ml-2 inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
            {organization.name}
          </span>
        )}
      </p>

      {isSuccess && (
        <div className="bg-green-50 border border-green-200 text-green-800 px-4 py-3 rounded-lg mb-6">
          Payment successful! Your plan will be upgraded shortly.
        </div>
      )}

      {/* Current Plan Summary */}
      <div className="bg-white rounded-lg shadow-sm border p-6 mb-8">
        <div className="flex items-center justify-between">
          <div>
            <h3 className="text-lg font-semibold text-gray-900">Current Plan</h3>
            <p className="text-sm text-gray-600 mt-1">
              <span className="font-medium capitalize">{currentPlan}</span>
              {subscription?.status && (
                <span className={`ml-2 inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${
                  subscription.status === 'active' ? 'bg-green-100 text-green-700' :
                  subscription.status === 'on_hold' ? 'bg-yellow-100 text-yellow-700' :
                  'bg-gray-100 text-gray-700'
                }`}>
                  {subscription.status}
                </span>
              )}
            </p>
          </div>
          <div className="text-right text-sm text-gray-500">
            <div>Devices: {organization?.device_limit === -1 ? 'Unlimited' : organization?.device_limit || 5}</div>
            <div>Fleets: {organization?.fleet_limit === -1 ? 'Unlimited' : organization?.fleet_limit || 1}</div>
          </div>
        </div>
      </div>

      {/* Plan Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {plans.map((plan) => {
          const isCurrent = plan.key === currentPlan;
          const isDowngrade = plans.findIndex(p => p.key === currentPlan) > plans.findIndex(p => p.key === plan.key);

          return (
            <div
              key={plan.key}
              className={`bg-white rounded-lg border-2 p-6 relative ${
                plan.popular ? 'border-blue-500' : isCurrent ? 'border-gray-900' : 'border-gray-200'
              }`}
            >
              {plan.popular && (
                <div className="absolute -top-3 left-1/2 -translate-x-1/2 bg-blue-500 text-white text-xs font-medium px-3 py-1 rounded-full">
                  Most Popular
                </div>
              )}

              <h3 className="text-lg font-semibold text-gray-900">{plan.name}</h3>
              <div className="mt-2 mb-4">
                <span className="text-3xl font-bold text-gray-900">{plan.price}</span>
                <span className="text-gray-500 text-sm">{plan.period}</span>
              </div>

              <ul className="space-y-2 mb-6 text-sm">
                {plan.features.map((feature) => (
                  <li key={feature} className="flex items-start">
                    <span className="text-green-500 mr-2 mt-0.5">&#10003;</span>
                    <span className="text-gray-700">{feature}</span>
                  </li>
                ))}
              </ul>

              <button
                onClick={() => !isCurrent && !isDowngrade && handleUpgrade(plan.key)}
                disabled={isCurrent || isDowngrade || upgrading === plan.key}
                className={`w-full py-2 rounded-lg text-sm font-medium ${
                  isCurrent
                    ? 'bg-gray-100 text-gray-500 cursor-default'
                    : isDowngrade
                    ? 'bg-gray-50 text-gray-400 cursor-not-allowed'
                    : 'bg-gray-900 text-white hover:bg-gray-800 disabled:opacity-50'
                }`}
              >
                {isCurrent ? 'Current Plan' : isDowngrade ? 'Downgrade' : upgrading === plan.key ? 'Redirecting...' : plan.cta}
              </button>
            </div>
          );
        })}
      </div>
    </div>
  );
}
