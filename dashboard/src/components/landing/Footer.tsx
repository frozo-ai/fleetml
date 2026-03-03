const links = [
  { label: 'GitHub', href: 'https://github.com/fleetml/fleetml' },
  { label: 'Docs', href: 'https://github.com/fleetml/fleetml#readme' },
  { label: 'Dashboard', href: '/dashboard' },
];

export default function Footer() {
  return (
    <footer className="border-t border-white/5 py-8">
      <div className="max-w-7xl mx-auto px-4 sm:px-6">
        <div className="flex flex-col sm:flex-row items-center justify-between gap-4">
          <div className="flex items-center gap-2">
            <div className="w-6 h-6 rounded bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center">
              <span className="text-white font-bold text-[10px]">F</span>
            </div>
            <span className="text-gray-500 text-sm">FleetML</span>
            <span className="text-gray-700 text-xs ml-2">Apache 2.0</span>
          </div>

          <div className="flex items-center gap-6">
            {links.map((link) => (
              <a
                key={link.label}
                href={link.href}
                target={link.href.startsWith('http') ? '_blank' : undefined}
                rel={link.href.startsWith('http') ? 'noopener noreferrer' : undefined}
                className="text-gray-500 hover:text-gray-300 transition-colors text-sm"
              >
                {link.label}
              </a>
            ))}
          </div>

          <div className="text-center sm:text-right">
            <p className="text-gray-700 text-xs">
              &copy; {new Date().getFullYear()} FleetML Contributors
            </p>
            <p className="text-gray-800 text-[10px] mt-1">
              Last updated March 2026
            </p>
          </div>
        </div>
      </div>
    </footer>
  );
}
