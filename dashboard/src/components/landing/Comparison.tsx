import { useState } from 'react';

const comparisons = [
  {
    question: '"Can\'t I just use AWS Greengrass?"',
    answer: 'Greengrass locks you into AWS, isn\'t ML-focused, and requires deep AWS expertise. FleetML is chip-neutral, ML-first, and works with any infrastructure.',
  },
  {
    question: '"We\'ll build it with SSH scripts"',
    answer: 'Custom scripts work for 5 devices. At 50+, you need canary rollouts, rollback, offline handling, and multi-chip compilation. That\'s 6\u201312 weeks of engineering FleetML gives you on day one.',
  },
  {
    question: '"What about Balena or Mender?"',
    answer: 'Great for OS and container updates. But they don\'t understand ML models \u2014 no model-aware canary, no inference metrics, no hot-swap, no ONNX compilation.',
  },
  {
    question: '"MLflow handles our model lifecycle"',
    answer: 'MLflow stops at the cloud boundary. It packages your model but doesn\'t deploy it to physical edge hardware, handle offline operation, or manage device fleets.',
  },
];

function AccordionItem({ question, answer, isOpen, onToggle, index }: {
  question: string;
  answer: string;
  isOpen: boolean;
  onToggle: () => void;
  index: number;
}) {
  return (
    <div
      className="scroll-reveal glass overflow-hidden hover:bg-white/[0.08] transition-all duration-300"
      style={{ transitionDelay: `${index * 80}ms` }}
    >
      <button
        onClick={onToggle}
        className="w-full flex items-center justify-between p-6 text-left"
      >
        <h3 className="text-base font-semibold text-white pr-4">
          {question}
        </h3>
        <svg
          className={`w-5 h-5 text-gray-500 shrink-0 transition-transform duration-300 ${isOpen ? 'rotate-180' : ''}`}
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
        </svg>
      </button>
      <div
        className={`overflow-hidden transition-all duration-300 ${
          isOpen ? 'max-h-40 opacity-100' : 'max-h-0 opacity-0'
        }`}
      >
        <p className="text-gray-400 text-sm leading-relaxed px-6 pb-6">
          {answer}
        </p>
      </div>
    </div>
  );
}

export default function Comparison() {
  const [openIndex, setOpenIndex] = useState<number | null>(null);

  return (
    <section className="py-24 border-t border-white/5">
      <div className="max-w-4xl mx-auto px-4 sm:px-6">
        <div className="text-center mb-16 scroll-reveal">
          <h2 className="text-3xl sm:text-4xl font-bold text-white mb-4">
            Why not just&hellip;
          </h2>
          <p className="text-gray-400 max-w-2xl mx-auto">
            We get it. Here&rsquo;s how FleetML compares to the alternatives you&rsquo;re probably considering.
          </p>
        </div>

        {/* Structured comparison table — extractable by AI search */}
        <div className="scroll-reveal glass overflow-hidden mb-10">
          <table className="w-full text-sm text-left">
            <thead>
              <tr className="border-b border-white/10">
                <th className="px-5 py-3 text-gray-500 font-medium">Capability</th>
                <th className="px-5 py-3 text-white font-semibold">FleetML</th>
                <th className="px-5 py-3 text-gray-400 font-medium">Greengrass</th>
                <th className="px-5 py-3 text-gray-400 font-medium">Balena</th>
                <th className="px-5 py-3 text-gray-400 font-medium">MLflow</th>
              </tr>
            </thead>
            <tbody className="text-gray-400">
              <tr className="border-b border-white/5">
                <td className="px-5 py-3 text-gray-300">Edge deployment</td>
                <td className="px-5 py-3 text-green-400">Yes</td>
                <td className="px-5 py-3 text-green-400">Yes</td>
                <td className="px-5 py-3 text-green-400">Yes</td>
                <td className="px-5 py-3 text-red-400">No</td>
              </tr>
              <tr className="border-b border-white/5">
                <td className="px-5 py-3 text-gray-300">Chip-neutral</td>
                <td className="px-5 py-3 text-green-400">Yes</td>
                <td className="px-5 py-3 text-red-400">AWS only</td>
                <td className="px-5 py-3 text-yellow-400">Partial</td>
                <td className="px-5 py-3 text-gray-600">N/A</td>
              </tr>
              <tr className="border-b border-white/5">
                <td className="px-5 py-3 text-gray-300">ML-aware canary</td>
                <td className="px-5 py-3 text-green-400">Yes</td>
                <td className="px-5 py-3 text-red-400">No</td>
                <td className="px-5 py-3 text-red-400">No</td>
                <td className="px-5 py-3 text-red-400">No</td>
              </tr>
              <tr className="border-b border-white/5">
                <td className="px-5 py-3 text-gray-300">Zero-downtime hot-swap</td>
                <td className="px-5 py-3 text-green-400">Yes</td>
                <td className="px-5 py-3 text-red-400">No</td>
                <td className="px-5 py-3 text-red-400">No</td>
                <td className="px-5 py-3 text-red-400">No</td>
              </tr>
              <tr className="border-b border-white/5">
                <td className="px-5 py-3 text-gray-300">Offline-first</td>
                <td className="px-5 py-3 text-green-400">Yes</td>
                <td className="px-5 py-3 text-yellow-400">Partial</td>
                <td className="px-5 py-3 text-yellow-400">Partial</td>
                <td className="px-5 py-3 text-red-400">No</td>
              </tr>
              <tr>
                <td className="px-5 py-3 text-gray-300">Open source</td>
                <td className="px-5 py-3 text-green-400">Apache 2.0</td>
                <td className="px-5 py-3 text-red-400">No</td>
                <td className="px-5 py-3 text-yellow-400">Partial</td>
                <td className="px-5 py-3 text-green-400">Yes</td>
              </tr>
            </tbody>
          </table>
        </div>

        <div className="space-y-3">
          {comparisons.map((item, i) => (
            <AccordionItem
              key={i}
              question={item.question}
              answer={item.answer}
              isOpen={openIndex === i}
              onToggle={() => setOpenIndex(openIndex === i ? null : i)}
              index={i}
            />
          ))}
        </div>
      </div>
    </section>
  );
}
