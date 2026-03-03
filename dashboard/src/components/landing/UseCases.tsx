const useCases = [
  {
    industry: 'Manufacturing',
    title: 'Factory defect detection',
    description: 'Deploy YOLOv8 to 50+ cameras on your assembly line with FleetML. Update the model when defect patterns change — zero production stoppage, under 30 seconds per device.',
    chips: ['NVIDIA Jetson', 'x86'],
    color: 'blue',
  },
  {
    industry: 'Retail',
    title: 'Shelf monitoring at scale',
    description: 'Push inventory detection models to 200 stores running Raspberry Pi cameras with FleetML. Canary roll out store-by-store with automatic rollback.',
    chips: ['Raspberry Pi', 'Google Coral'],
    color: 'purple',
  },
  {
    industry: 'Agriculture',
    title: 'Crop disease detection',
    description: 'Deploy to 500 solar-powered field sensors with spotty connectivity using FleetML. Offline-first architecture means your models keep running without internet.',
    chips: ['ARM', 'Intel NUC'],
    color: 'cyan',
  },
  {
    industry: 'Smart Cities',
    title: 'Traffic and pedestrian analytics',
    description: 'Update models across thousands of intersection cameras with FleetML. Zero-downtime hot-swap means zero dropped inferences during updates.',
    chips: ['NVIDIA Jetson', 'Hailo'],
    color: 'green',
  },
];

const colorMap: Record<string, { badge: string; dot: string }> = {
  blue: { badge: 'bg-blue-500/10 text-blue-400 border-blue-500/20', dot: 'bg-blue-400' },
  purple: { badge: 'bg-purple-500/10 text-purple-400 border-purple-500/20', dot: 'bg-purple-400' },
  cyan: { badge: 'bg-cyan-500/10 text-cyan-400 border-cyan-500/20', dot: 'bg-cyan-400' },
  green: { badge: 'bg-green-500/10 text-green-400 border-green-500/20', dot: 'bg-green-400' },
};

export default function UseCases() {
  return (
    <section className="py-24 border-t border-white/5">
      <div className="max-w-7xl mx-auto px-4 sm:px-6">
        <div className="text-center mb-16 scroll-reveal">
          <h2 className="text-3xl sm:text-4xl font-bold text-white mb-4">
            Built for your fleet
          </h2>
          <p className="text-gray-400 max-w-2xl mx-auto">
            Wherever you deploy ML models on edge hardware, FleetML handles the hard parts.
          </p>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">
          {useCases.map((uc, i) => {
            const colors = colorMap[uc.color];
            return (
              <div
                key={uc.industry}
                className="scroll-reveal glass p-6 hover:bg-white/[0.08] transition-all duration-300"
                style={{ transitionDelay: `${i * 100}ms` }}
              >
                <span className={`inline-block px-2.5 py-1 rounded text-xs font-medium border mb-3 ${colors.badge}`}>
                  {uc.industry}
                </span>
                <h3 className="text-lg font-semibold text-white mb-2">{uc.title}</h3>
                <p className="text-gray-400 text-sm leading-relaxed mb-4">{uc.description}</p>
                <div className="flex flex-wrap gap-2">
                  {uc.chips.map((chip) => (
                    <span key={chip} className="flex items-center gap-1.5 text-xs text-gray-500">
                      <span className={`w-1.5 h-1.5 rounded-full ${colors.dot}`} />
                      {chip}
                    </span>
                  ))}
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </section>
  );
}
