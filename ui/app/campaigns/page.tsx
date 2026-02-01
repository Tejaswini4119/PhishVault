"use client";

import { useEffect, useState } from 'react';
import { Shield, Users, Globe, Target } from 'lucide-react';

interface Campaign {
    id: string;
    name: string;
    threat_actor: string;
    target_sector: string;
    node_count: number;
    risk_level: string;
}

export default function CampaignsPage() {
    const [campaigns, setCampaigns] = useState<Campaign[]>([]);

    useEffect(() => {
        fetch('/api/campaigns')
            .then(res => res.json())
            .then(data => setCampaigns(data || []))
            .catch(err => console.error("Failed to fetch campaigns:", err));
    }, []);

    return (
        <div className="space-y-8">
            <div>
                <h2 className="text-3xl font-bold text-white tracking-tight">Active Campaigns</h2>
                <p className="text-slate-400 mt-2">Clusters of related malicious activity identified by the Graph Engine.</p>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-6">
                {campaigns.map((camp) => (
                    <div key={camp.id} className="bg-slate-900 border border-slate-800 rounded-xl p-6 hover:border-blue-500/50 transition-all shadow-xl">
                        <div className="flex justify-between items-start mb-4">
                            <div className="p-3 bg-blue-900/20 rounded-lg text-blue-400">
                                <Shield size={24} />
                            </div>
                            <span className={`px-3 py-1 rounded-full text-xs font-bold uppercase ${camp.risk_level === 'CRITICAL' ? 'bg-red-900/50 text-red-400' : 'bg-orange-900/50 text-orange-400'
                                }`}>
                                {camp.risk_level}
                            </span>
                        </div>

                        <h3 className="text-xl font-bold text-white mb-1">{camp.name}</h3>
                        <p className="text-slate-500 text-sm mb-6">ID: {camp.id}</p>

                        <div className="space-y-3">
                            <div className="flex items-center text-sm text-slate-300">
                                <Users size={16} className="mr-3 text-slate-500" />
                                <span>Actor: <span className="text-white font-medium">{camp.threat_actor}</span></span>
                            </div>
                            <div className="flex items-center text-sm text-slate-300">
                                <Target size={16} className="mr-3 text-slate-500" />
                                <span>Target: <span className="text-white font-medium">{camp.target_sector}</span></span>
                            </div>
                            <div className="flex items-center text-sm text-slate-300">
                                <Globe size={16} className="mr-3 text-slate-500" />
                                <span>Infrastructure: <span className="text-white font-medium">{camp.node_count} nodes</span></span>
                            </div>
                        </div>

                        <div className="mt-6 pt-6 border-t border-slate-800 flex gap-2">
                            <button className="flex-1 bg-slate-800 hover:bg-slate-700 text-white py-2 rounded-lg text-sm font-medium transition-colors">
                                View Intelligence
                            </button>
                            <button className="flex-1 bg-blue-600 hover:bg-blue-500 text-white py-2 rounded-lg text-sm font-medium transition-colors">
                                Graph View
                            </button>
                        </div>
                    </div>
                ))}
            </div>
        </div>
    );
}
