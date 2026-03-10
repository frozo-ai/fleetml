const platforms = [
  'NVIDIA Jetson',
  'Raspberry Pi',
  'Intel NUC',
  'Google Coral',
  'Qualcomm',
  'Hailo',
  'x86',
  'ARM',
];

export default function LogoCloud() {
  return (
    <section className="py-16 border-t border-white/5">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 text-center">
        <h2 className="text-lg font-semibold text-gray-300 mb-8 scroll-reveal">
          Works with every edge chip
        </h2>
        <div className="flex flex-wrap items-center justify-center gap-4 scroll-reveal">
          {platforms.map((name) => (
            <span
              key={name}
              className="px-5 py-2.5 rounded-lg border border-white/10 bg-white/[0.03] text-gray-400 text-sm font-medium hover:text-white hover:border-white/20 hover:bg-white/5 transition-all duration-300"
            >
              {name}
            </span>
          ))}
        </div>
      </div>
    </section>
  );
}
