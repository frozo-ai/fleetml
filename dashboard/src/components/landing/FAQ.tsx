import { useState } from 'react';

const faqs = [
  {
    question: 'Do I need to change my model format?',
    answer:
      'No. Export your model as ONNX (most frameworks support this natively) and FleetML handles the rest. We auto-compile to TensorRT, OpenVINO, TFLite, or ONNX Runtime based on each device\'s chip.',
  },
  {
    question: 'What happens if a device goes offline?',
    answer:
      'It keeps running the last-deployed model with zero interruption. When connectivity returns, the agent syncs buffered metrics and picks up any pending model updates automatically.',
  },
  {
    question: 'How does the canary rollout work?',
    answer:
      'You define stages (e.g., 5% \u2192 50% \u2192 100%). FleetML deploys to each stage, monitors error rates and inference latency, and automatically rolls back if thresholds are exceeded. You can also define custom policies.',
  },
  {
    question: 'Can I self-host FleetML?',
    answer:
      'Yes. FleetML is Apache 2.0 licensed. Run the control plane on your own infrastructure with Docker Compose, Kubernetes, or Railway. The cloud version adds managed hosting and support.',
  },
  {
    question: 'What devices are supported?',
    answer:
      'Any Linux device that can run a 15MB Go binary: NVIDIA Jetson (Nano, Xavier, Orin), Raspberry Pi (3/4/5), Intel NUCs, Google Coral, x86 servers, and ARM-based SBCs. The agent has no GPU driver dependency.',
  },
  {
    question: 'How is this different from AWS IoT Greengrass?',
    answer:
      'Greengrass is a general IoT platform tied to AWS. FleetML is purpose-built for ML: chip-neutral compilation, model-aware canary rollouts, zero-downtime hot-swap, inference metrics, and drift detection. No AWS lock-in.',
  },
  {
    question: 'Is there a free tier?',
    answer:
      'Yes. The free plan includes 5 devices and 1 fleet \u2014 enough to prototype and validate your edge ML workflow. No credit card required.',
  },
];

export default function FAQ() {
  const [openIndex, setOpenIndex] = useState<number | null>(null);

  return (
    <section id="faq" className="py-24 border-t border-white/5">
      <div className="max-w-3xl mx-auto px-4 sm:px-6">
        <div className="text-center mb-16 scroll-reveal">
          <h2 className="text-3xl sm:text-4xl font-bold text-white mb-4">
            Frequently asked questions
          </h2>
          <p className="text-gray-400">
            Everything you need to know about FleetML.
          </p>
        </div>

        <div className="space-y-3">
          {faqs.map((faq, i) => (
            <div
              key={i}
              className="scroll-reveal glass overflow-hidden hover:bg-white/[0.08] transition-all duration-300"
              style={{ transitionDelay: `${i * 60}ms` }}
            >
              <button
                onClick={() => setOpenIndex(openIndex === i ? null : i)}
                className="w-full flex items-center justify-between p-5 text-left"
              >
                <h3 className="text-sm font-semibold text-white pr-4">
                  {faq.question}
                </h3>
                <svg
                  className={`w-5 h-5 text-gray-500 shrink-0 transition-transform duration-300 ${
                    openIndex === i ? 'rotate-180' : ''
                  }`}
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
                </svg>
              </button>
              <div
                className={`overflow-hidden transition-all duration-300 ${
                  openIndex === i ? 'max-h-48 opacity-100' : 'max-h-0 opacity-0'
                }`}
              >
                <p className="text-gray-400 text-sm leading-relaxed px-5 pb-5">
                  {faq.answer}
                </p>
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
