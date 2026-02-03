"use client";

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { useAuth } from '@/components/AuthProvider';
import { useRouter } from 'next/navigation';

interface Scan {
    scan_id: string;
    target_url: string;
    status: string;
    timestamp: string;
    verdict: string;
    risk_score: number;
}

export default function ScansPage() {
    const [scans, setScans] = useState<Scan[]>([]);
    const [pageLoading, setPageLoading] = useState(true);
    const { token, loading: authLoading, isAuthenticated } = useAuth();
    const router = useRouter();

    useEffect(() => {
        if (!authLoading && !isAuthenticated) {
            router.push("/login");
            return;
        }

        if (isAuthenticated && token) {
            setPageLoading(true);
            fetch('/api/scans', {
                headers: {
                    "Authorization": `Bearer ${token}`
                }
            })
                .then(res => {
                    if (res.status === 401) {
                        // Token might be invalid/expired
                        router.push("/login");
                        throw new Error("Unauthorized");
                    }
                    return res.json();
                })
                .then(data => {
                    // Check if data is array
                    if (Array.isArray(data)) {
                        setScans(data);
                    } else {
                        setScans([]);
                    }
                    setPageLoading(false);
                })
                .catch(err => {
                    console.error("Failed to fetch scans:", err);
                    setPageLoading(false);
                });
        }
    }, [authLoading, isAuthenticated, token, router]);

    if (authLoading || pageLoading) return <div className="text-white p-6">Loading scans...</div>;

    return (
        <div className="space-y-6">
            <div className="flex justify-between items-center">
                <div>
                    <h2 className="text-2xl font-bold text-white">Live Scan Registry</h2>
                    <p className="text-slate-400 text-sm">Comprehensive list of all submitted URLs and files.</p>
                </div>
                <button
                    className="bg-blue-600 hover:bg-blue-500 text-white px-4 py-2 rounded-lg text-sm font-medium transition-colors"
                    onClick={() => window.location.reload()}
                >
                    Refresh Data
                </button>
            </div>

            <div className="bg-slate-900 border border-slate-800 rounded-xl overflow-hidden shadow-lg">
                <div className="overflow-x-auto">
                    <table className="w-full text-sm text-left text-slate-400">
                        <thead className="text-xs text-slate-500 uppercase bg-slate-950/50 border-b border-slate-800">
                            <tr>
                                <th className="px-6 py-4">Scan ID</th>
                                <th className="px-6 py-4">Target / Subject</th>
                                <th className="px-6 py-4">Risk Score</th>
                                <th className="px-6 py-4">Ingestion</th>
                                <th className="px-6 py-4">Timestamp</th>
                                <th className="px-6 py-4">Verdict</th>
                                <th className="px-6 py-4">Actions</th>
                            </tr>
                        </thead>
                        <tbody>
                            {scans.length === 0 ? (
                                <tr><td colSpan={7} className="px-6 py-8 text-center text-slate-500">No records found.</td></tr>
                            ) : (
                                scans.map((scan) => (
                                    <tr key={scan.scan_id} className="border-b border-slate-800 hover:bg-slate-800/50 transition-colors">
                                        <td className="px-6 py-4 font-mono text-xs text-slate-500">
                                            {scan.scan_id.substring(0, 12)}...
                                        </td>
                                        <td className="px-6 py-4">
                                            <div className="text-white font-medium truncate max-w-md" title={scan.target_url}>
                                                {scan.target_url}
                                            </div>
                                        </td>
                                        <td className="px-6 py-4">
                                            <div className="flex items-center gap-2">
                                                {(() => {
                                                    const displayScore = Math.round(scan.risk_score * 100);
                                                    return (
                                                        <>
                                                            <span className="text-sm font-bold text-white">{displayScore}</span>
                                                            <div className="w-16 h-1.5 bg-slate-800 rounded-full overflow-hidden">
                                                                <div
                                                                    className={`h-full rounded-full ${displayScore > 70 ? 'bg-red-500' : displayScore > 40 ? 'bg-yellow-500' : 'bg-green-500'}`}
                                                                    style={{ width: `${displayScore}%` }}
                                                                />
                                                            </div>
                                                        </>
                                                    );
                                                })()}
                                            </div>
                                        </td>
                                        <td className="px-6 py-4">
                                            <span className="px-2 py-1 rounded text-xs bg-slate-800 text-slate-300">
                                                API
                                            </span>
                                        </td>
                                        <td className="px-6 py-4 text-xs">
                                            {new Date(scan.timestamp).toLocaleString()}
                                        </td>
                                        <td className="px-6 py-4">
                                            <Badge verdict={scan.verdict} />
                                        </td>
                                        <td className="px-6 py-4">
                                            <Link href={`/scans/${scan.scan_id}`} className="text-blue-400 hover:text-blue-300 text-xs font-semibold">
                                                View Report
                                            </Link>
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

const Badge = ({ verdict }: { verdict: string }) => {
    let color = "bg-slate-800 text-slate-400";
    if (verdict === "MALICIOUS") color = "bg-red-900/40 text-red-400 border border-red-900";
    if (verdict === "SAFE") color = "bg-green-900/40 text-green-400 border border-green-900";
    if (verdict === "SUSPICIOUS") color = "bg-yellow-900/40 text-yellow-400 border border-yellow-900";

    return (
        <span className={`px-3 py-1 rounded-full text-xs font-bold ${color}`}>
            {verdict || "PENDING"}
        </span>
    );
};
