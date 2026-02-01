"use client";

import { useEffect, useState } from 'react';
import { Info, Share2, ShieldAlert } from 'lucide-react';

interface Campaign {
    campaign_id: string;
    threat_actor: string;
    affected_assets: number;
    risk_level: string;
}

export default function GraphPage() {
    const [campaigns, setCampaigns] = useState<Campaign[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState("");

    useEffect(() => {
        fetch('/api/campaigns')
            .then(res => {
                if (!res.ok) throw new Error("Failed to connect to Graph API");
                return res.json();
            })
            .then(data => {
                setCampaigns(data || []);
                setLoading(false);
            })
            .catch(err => {
                console.error(err);
                setError("Graph Database (Neo4j) is unreachable.");
                setLoading(false);
            });
    }, []);

    return (
        <div className="h-[calc(100vh-8rem)] flex flex-col">
            <div className="mb-6">
                <h2 className="text-2xl font-bold text-white">Campaign Intelligence (Graph)</h2>
                <p className="text-slate-400 text-sm">Active threat campaigns identified by the graph engine.</p>
            </div>

            {loading ? (
                <div className="text-slate-400">Querying Graph Database...</div>
            ) : error ? (
                <div className="p-4 border border-red-900/50 bg-red-900/20 text-red-400 rounded-lg">
                    {error}
                </div>
            ) : campaigns.length === 0 ? (
                <div className="p-8 border border-slate-800 rounded-xl text-center text-slate-500">
                    <Share2 className="mx-auto mb-4 opacity-50" size={48} />
                    <p>No active campaigns found in the graph.</p>
                </div>
            ) : (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                    {campaigns.map(c => (
                        <div key={c.campaign_id} className="bg-slate-900 border border-slate-800 p-4 rounded-xl hover:border-blue-500/50 transition-colors">
                            <div className="flex items-start justify-between mb-4">
                                <div className="bg-blue-900/30 p-2 rounded-lg">
                                    <ShieldAlert className="text-blue-400" size={24} />
                                </div>
                                <span className="text-xs font-mono text-slate-500">{c.campaign_id.substring(0, 8)}</span>
                            </div>
                            <h3 className="text-lg font-bold text-white mb-1">
                                {c.threat_actor || "Unknown Actor"}
                            </h3>
                            <div className="flex justify-between items-center mt-4 text-sm">
                                <span className="text-slate-400">{c.affected_assets} Assets</span>
                                <span className="text-red-400 font-bold">{c.risk_level}</span>
                            </div>
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
}
