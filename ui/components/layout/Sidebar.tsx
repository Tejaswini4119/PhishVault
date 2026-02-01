import Link from 'next/link';
import { Home, Activity, Shield, Settings, Sliders } from 'lucide-react';

const Sidebar = () => {
  return (
    <aside className="w-64 bg-slate-900 text-white h-screen fixed left-0 top-0 overflow-y-auto border-r border-slate-800">
      <div className="p-6 border-b border-slate-800">
        <h1 className="text-xl font-bold text-blue-400 tracking-wider">PHISHVAULT</h1>
        <p className="text-xs text-slate-500 mt-1">Advanced Threat Intel</p>
      </div>

      <nav className="p-4 space-y-2">
        <NavLink href="/" icon={<Home size={20} />} label="Dashboard" />
        <NavLink href="/scans" icon={<Activity size={20} />} label="Live Scans" />
        <NavLink href="/campaigns" icon={<Shield size={20} />} label="Campaigns" />
        <NavLink href="/graph" icon={<Sliders size={20} />} label="Graph Explorer" />
        <NavLink href="/settings" icon={<Settings size={20} />} label="Settings" />
      </nav>

      <div className="absolute bottom-0 w-full p-4 border-t border-slate-800">
        <div className="flex items-center gap-3">
          <div className="w-8 h-8 rounded-full bg-blue-500 flex items-center justify-center font-bold">
            A
          </div>
          <div>
            <p className="text-sm font-medium">Analyst</p>
            <p className="text-xs text-slate-400">SOC Team</p>
          </div>
        </div>
      </div>
    </aside>
  );
};

const NavLink = ({ href, icon, label }: { href: string; icon: React.ReactNode; label: string }) => (
  <Link 
    href={href} 
    className="flex items-center gap-3 px-4 py-3 text-slate-300 hover:bg-slate-800 hover:text-blue-400 rounded-lg transition-colors"
  >
    {icon}
    <span className="text-sm font-medium">{label}</span>
  </Link>
);

export default Sidebar;
