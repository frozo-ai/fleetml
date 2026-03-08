import { Link } from 'react-router-dom';

const productLinks = [
  { label: 'Features', href: '#features' },
  { label: 'Pricing', href: '#pricing' },
  { label: 'How It Works', href: '#how-it-works' },
  { label: 'Architecture', href: '#architecture' },
];

const resourceLinks = [
  { label: 'Documentation', href: 'https://github.com/ashish-frozo/fleetML#readme', external: true },
  { label: 'GitHub', href: 'https://github.com/ashish-frozo/fleetML', external: true },
  { label: 'API Reference', href: 'https://github.com/ashish-frozo/fleetML/blob/main/docs/api-reference.md', external: true },
];

export default function Footer() {
  return (
    <footer className="border-t border-white/5 py-12">
      <div className="max-w-7xl mx-auto px-4 sm:px-6">
        <div className="grid grid-cols-2 md:grid-cols-4 gap-8 mb-10">
          {/* Brand */}
          <div className="col-span-2 md:col-span-1">
            <div className="flex items-center gap-2 mb-3">
              <div className="w-6 h-6 rounded bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center">
                <span className="text-white font-bold text-[10px]">F</span>
              </div>
              <span className="text-white font-bold text-sm">FleetML</span>
            </div>
            <p className="text-gray-600 text-xs leading-relaxed">
              Open-source edge MLOps platform. Deploy AI models to device fleets with one command.
            </p>
          </div>

          {/* Product */}
          <div>
            <h4 className="text-gray-400 text-xs font-semibold uppercase tracking-wider mb-3">Product</h4>
            <ul className="space-y-2">
              {productLinks.map((link) => (
                <li key={link.label}>
                  <a href={link.href} className="text-gray-500 hover:text-gray-300 transition-colors text-sm">
                    {link.label}
                  </a>
                </li>
              ))}
            </ul>
          </div>

          {/* Resources */}
          <div>
            <h4 className="text-gray-400 text-xs font-semibold uppercase tracking-wider mb-3">Resources</h4>
            <ul className="space-y-2">
              {resourceLinks.map((link) => (
                <li key={link.label}>
                  <a
                    href={link.href}
                    target={link.external ? '_blank' : undefined}
                    rel={link.external ? 'noopener noreferrer' : undefined}
                    className="text-gray-500 hover:text-gray-300 transition-colors text-sm"
                  >
                    {link.label}
                  </a>
                </li>
              ))}
            </ul>
          </div>

          {/* Account */}
          <div>
            <h4 className="text-gray-400 text-xs font-semibold uppercase tracking-wider mb-3">Account</h4>
            <ul className="space-y-2">
              <li>
                <Link to="/login" className="text-gray-500 hover:text-gray-300 transition-colors text-sm">
                  Sign In
                </Link>
              </li>
              <li>
                <Link to="/signup" className="text-gray-500 hover:text-gray-300 transition-colors text-sm">
                  Create Account
                </Link>
              </li>
              <li>
                <Link to="/dashboard" className="text-gray-500 hover:text-gray-300 transition-colors text-sm">
                  Dashboard
                </Link>
              </li>
            </ul>
          </div>
        </div>

        <div className="border-t border-white/5 pt-6 flex flex-col sm:flex-row items-center justify-between gap-4">
          <p className="text-gray-700 text-xs">
            &copy; {new Date().getFullYear()} FleetML Contributors &middot; Apache 2.0 License
          </p>
        </div>
      </div>
    </footer>
  );
}
