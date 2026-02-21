"use client";

import { useEffect, useState } from 'react';
import StatsCard from "@/components/dashboard/StatsCard";
import { Activity, Disc, AlertTriangle, Database } from 'lucide-react';
import { useAuth } from '@/components/AuthProvider';
import { useRouter } from 'next/navigation';

interface Scan {
  scan_id: string;
  target_url: string;
  status: string;
  timestamp: string;
  verdict: string;
}

interface Stats {
  active_scans: number;
  campaigns: number;
  threats_blocked: number;
  graph_nodes: number;
}

export default function Home() {
  const [scans, setScans] = useState<Scan[]>([]);
  const [stats, setStats] = useState<Stats>({
    active_scans: 0,
    campaigns: 0,
    threats_blocked: 0,
    graph_nodes: 0
  });
  const { token, loading, isAuthenticated } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (!loading && !isAuthenticated) {
      router.push("/login");
      return;
    }

    if (!token) return;

    // Poll for updates every 5 seconds
    const fetchData = async () => {
      try {
        const headers = { "Authorization": `Bearer ${token}` };
        const [scansRes, statsRes] = await Promise.all([
          fetch('/api/scans', { headers }),
          fetch('/api/stats', { headers })
        ]);

        if (scansRes.ok) {
          const data = await scansRes.json();
          // Ensure data is array
          setScans(Array.isArray(data) ? data : []);
        } else if (scansRes.status === 401) {
          router.push("/login"); // Token expired
        }

        if (statsRes.ok) {
          const data = await statsRes.json();
          setStats(data);
        }
      } catch (error) {
        console.error("Failed to fetch dashboard data:", error);
      }
    };

    fetchData();
    const interval = setInterval(fetchData, 5000);
    return () => clearInterval(interval);
  }, [loading, isAuthenticated, token, router]);

  if (loading) return <div className="text-white p-6">Loading dashboard...</div>;

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div>
        <h2 className="text-2xl font-bold text-white">Security Dashboard</h2>
        <p className="text-slate-400 text-sm">Real-time threat intelligence overview.</p>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <StatsCard
          title="Active Scans"
          value={stats.active_scans}
          change="Live"
          trend="neutral"
          icon={<Activity size={24} />}
        />
        <StatsCard
          title="Campaigns"
          value={stats.campaigns}
          change="mock"
          trend="up"
          icon={<Disc size={24} />}
        />
        <StatsCard
          title="Threats Blocked"
          value={stats.threats_blocked}
          change="AI Est."
          trend="up"
          icon={<AlertTriangle size={24} />}
        />
        <StatsCard
          title="Graph Nodes"
          value={stats.graph_nodes}
          change="Synced"
          trend="neutral"
          icon={<Database size={24} />}
        />
      </div>

      {/* Quick Analysis Tools */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <ToolCard
          title="Email Verifier"
          description="Analyze email headers and bodies for forensic threats."
          icon={<div className="bg-blue-500/20 p-3 rounded-lg text-blue-400"><Disc size={28} /></div>}
          href="/email"
        />
        <ToolCard
          title="File Analyzer"
          description="Static analysis for malicious attachments and embedded URLs."
          icon={<div className="bg-purple-500/20 p-3 rounded-lg text-purple-400"><Activity size={28} /></div>}
          href="/file"
        />
      </div>

      {/* Recent Activity Table */}
      <div className="bg-slate-900 border border-slate-800 rounded-xl overflow-hidden">
        <div className="p-6 border-b border-slate-800 flex justify-between items-center">
          <h3 className="font-semibold text-white">Recent Scans</h3>
          <button className="text-sm text-blue-400 hover:text-blue-300">View All</button>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full text-sm text-left text-slate-400">
            <thead className="text-xs text-slate-500 uppercase bg-slate-950/50">
              <tr>
                <th className="px-6 py-3">Scan ID</th>
                <th className="px-6 py-3">Target URL</th>
                <th className="px-6 py-3">Timestamp</th>
                <th className="px-6 py-3">Verdict</th>
              </tr>
            </thead>
            <tbody>
              {scans.length === 0 ? (
                <tr><td colSpan={4} className="px-6 py-4 text-center">No scans recorded yet.</td></tr>
              ) : (
                scans.map((scan) => (
                  <tr key={scan.scan_id} className="border-b border-slate-800 hover:bg-slate-800/50">
                    <td className="px-6 py-4 font-mono text-xs">{scan.scan_id.substring(0, 8)}...</td>
                    <td className="px-6 py-4 text-white">{scan.target_url}</td>
                    <td className="px-6 py-4">{new Date(scan.timestamp).toLocaleString()}</td>
                    <td className="px-6 py-4">
                      <span className={`font-bold ${scan.verdict === 'MALICIOUS' ? 'text-red-400' : 'text-green-400'}`}>
                        {scan.verdict || 'PENDING'}
                      </span>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}

function ToolCard({ title, description, icon, href }: { title: string; description: string; icon: React.ReactNode; href: string }) {
  const router = useRouter();
  return (
    <div
      onClick={() => router.push(href)}
      className="bg-slate-900 border border-slate-800 p-6 rounded-xl hover:border-blue-500/50 cursor-pointer transition-all group"
    >
      <div className="flex items-start gap-4">
        {icon}
        <div>
          <h3 className="font-bold text-white group-hover:text-blue-400 transition-colors">{title}</h3>
          <p className="text-sm text-slate-400 mt-1">{description}</p>
        </div>
      </div>
    </div>
  );
}
