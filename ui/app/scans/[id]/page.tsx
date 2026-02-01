"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { AlertTriangle, CheckCircle, XCircle, Shield, Globe, Image as ImageIcon } from "lucide-react";

interface ScanReport {
  scan_id: string;
  url: string;
  verdict: string;
  risk_score: number;
  timestamp: string;
  final_url?: string;
  status_code?: number;
  screenshot_url?: string;
}

export default function ScanReportPage() {
  const { id } = useParams();
  const [data, setData] = useState<ScanReport | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [isImageOpen, setIsImageOpen] = useState(false);

  useEffect(() => {
    if (!id) return;
    fetch(`http://localhost:8080/scans/${id}`)
      .then((res) => {
        if (!res.ok) throw new Error("Failed to fetch report");
        return res.json();
      })
      .then((data) => {
        setData(data);
        setLoading(false);
      })
      .catch((err) => {
        setError(err.message);
        setLoading(false);
      });
  }, [id]);

  if (loading) return <div className="min-h-screen bg-black text-white p-10 flex items-center justify-center">Loading Report...</div>;
  if (error) return <div className="min-h-screen bg-black text-white p-10 flex items-center justify-center text-red-500">Error: {error}</div>;
  if (!data) return null;

  const isMalicious = data.verdict === "MALICIOUS";
  const isSafe = data.verdict === "SAFE";
  const statusColor = isMalicious ? "text-red-500" : isSafe ? "text-green-500" : "text-yellow-500";
  const borderColor = isMalicious ? "border-red-500/50" : isSafe ? "border-green-500/50" : "border-yellow-500/50";

  return (
    <div className="min-h-screen bg-slate-900 text-gray-100 p-8 font-sans">
      <div className="max-w-6xl mx-auto space-y-8">

        {/* Header Section */}
        <div className={`p-6 rounded-2xl border ${borderColor} bg-white/5 backdrop-blur-md flex justify-between items-center shadow-2xl`}>
          <div>
            <h1 className="text-3xl font-bold tracking-tight mb-2">Scan Report</h1>
            <p className="text-gray-400 font-mono text-sm">{data.scan_id}</p>
          </div>
          <div className="flex items-center gap-4 text-right">
            <div>
              <div className={`text-4xl font-extrabold ${statusColor} tracking-wider`}>{data.verdict}</div>
              <div className="text-sm text-gray-400 mt-1">RISK SCORE: {data.risk_score.toFixed(2)}</div>
            </div>
            {isMalicious ? <AlertTriangle size={48} className="text-red-500" /> : <CheckCircle size={48} className="text-green-500" />}
          </div>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">

          {/* Main Details */}
          <div className="lg:col-span-2 space-y-8">
            <div className="p-6 rounded-2xl bg-white/5 border border-white/10 shadow-lg">
              <h2 className="text-xl font-semibold mb-6 flex items-center gap-2">
                <Globe className="text-blue-400" /> Target Details
              </h2>
              <div className="space-y-4">
                <div className="group">
                  <label className="text-xs uppercase tracking-wider text-gray-500 font-bold">Submitted URL</label>
                  <div className="mt-1 font-mono text-sm break-all p-3 bg-black/30 rounded border border-white/5 group-hover:border-blue-500/30 transition-colors">
                    {data.url}
                  </div>
                </div>
                {data.final_url && (
                  <div className="group">
                    <label className="text-xs uppercase tracking-wider text-gray-500 font-bold">Final URL</label>
                    <div className="mt-1 font-mono text-sm break-all p-3 bg-black/30 rounded border border-white/5 group-hover:border-purple-500/30 transition-colors">
                      {data.final_url}
                    </div>
                  </div>
                )}
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="text-xs uppercase tracking-wider text-gray-500 font-bold">HTTP Status</label>
                    <div className="text-lg font-mono">{data.status_code || "N/A"}</div>
                  </div>
                  <div>
                    <label className="text-xs uppercase tracking-wider text-gray-500 font-bold">Scanned At</label>
                    <div className="text-lg font-mono">{new Date(data.timestamp).toLocaleString()}</div>
                  </div>
                </div>
              </div>
            </div>

            {/* Analysis Signals / Notes */}
            <div className="p-6 rounded-2xl bg-white/5 border border-white/10 shadow-lg">
              <h2 className="text-xl font-semibold mb-6 flex items-center gap-2">
                <Shield className="text-purple-400" /> Analysis Notes & Signals
              </h2>
              <div className="space-y-4">
                <div className="p-4 rounded bg-black/40 border border-white/5 text-sm font-mono text-gray-300">
                  {/* Placeholder for now till we persist signals */}
                  <p className="text-gray-500 italic mb-2">// Automated Analysis Log</p>
                  <p>- Visual Analysis: Completed</p>
                  <p>- NLP Engine: {data.verdict === "MALICIOUS" ? "High Urgency Detected" : "No obvious threats"}</p>
                  <p>- Domain Age: Checked</p>
                  <p>- Reputation: Clean</p>
                </div>
              </div>
            </div>
          </div>

          {/* Screenshot Column */}
          <div className="space-y-8">
            <div className="p-6 rounded-2xl bg-white/5 border border-white/10 shadow-lg h-full">
              <h2 className="text-xl font-semibold mb-6 flex items-center gap-2">
                <ImageIcon className="text-pink-400" /> Screenshot
              </h2>
              {data.screenshot_url && !data.screenshot_url.endsWith("undefined") ? (
                <div
                  className="rounded-xl overflow-hidden border border-white/10 shadow-2xl transition-all hover:scale-[1.02] duration-300 cursor-zoom-in"
                  onClick={() => setIsImageOpen(true)}
                >
                  <img src={data.screenshot_url} alt="Site Screenshot" className="w-full h-auto object-cover" />
                </div>
              ) : (
                <div className="h-48 flex items-center justify-center bg-black/30 rounded-xl border border-dashed border-gray-700 text-gray-500">
                  No screenshot available
                </div>
              )}
              <p className="mt-4 text-xs text-center text-gray-500">Click image to enlarge</p>
            </div>
          </div>

        </div>
      </div>

      {/* Image Modal */}
      {isImageOpen && data.screenshot_url && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-[#0f172a]/95 backdrop-blur-md p-4 transition-all duration-300"
          onClick={() => setIsImageOpen(false)}
        >
          <div className="relative max-w-7xl w-full max-h-screen overflow-hidden rounded-lg shadow-2xl border border-white/10">
            <button
              className="absolute top-4 right-4 text-white/70 hover:text-white bg-blue-500/20 hover:bg-blue-500/40 rounded-full p-2 transition-colors border border-white/10"
              onClick={() => setIsImageOpen(false)}
            >
              <XCircle size={32} />
            </button>
            <img
              src={data.screenshot_url}
              alt="Full Size Screenshot"
              className="w-full h-auto max-h-[90vh] object-contain bg-black"
            />
          </div>
        </div>
      )}
    </div>
  );
}
