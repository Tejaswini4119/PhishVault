"use client";

import { useState } from 'react';
import { Mail, Send, ShieldCheck, AlertCircle } from 'lucide-react';
import { useAuth } from '@/components/AuthProvider';

export default function EmailPage() {
    const [emailContent, setEmailContent] = useState('');
    const [status, setStatus] = useState<string | null>(null);
    const { token } = useAuth();

    const handleAnalyze = async () => {
        if (!emailContent || !token) return;
        setStatus('Analyzing...');
        try {
            const res = await fetch('/api/submit-email', {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ email_raw: emailContent })
            });
            if (res.ok) {
                setStatus('Analysis in progress. Check Live Scans for results.');
                setEmailContent('');
            } else {
                setStatus('Failed to submit email.');
            }
        } catch (err) {
            setStatus('Error: ' + err);
        }
    };

    return (
        <div className="max-w-4xl mx-auto space-y-6">
            <div>
                <h2 className="text-2xl font-bold text-white flex items-center gap-2">
                    <Mail className="text-blue-400" /> Email Verifier
                </h2>
                <p className="text-slate-400 text-sm">Paste raw email content (EML) for deep forensic analysis.</p>
            </div>

            <div className="bg-slate-900 border border-slate-800 rounded-xl p-6 space-y-4">
                <textarea
                    value={emailContent}
                    onChange={(e) => setEmailContent(e.target.value)}
                    placeholder="Paste raw email headers and body here..."
                    className="w-full h-96 bg-slate-950 border border-slate-800 rounded-lg p-4 text-slate-300 font-mono text-sm focus:border-blue-500 outline-none resize-none"
                />

                <div className="flex justify-between items-center">
                    <p className="text-xs text-slate-500">
                        Supports: SPF/DKIM verification, Display Name spoofing detection, Link extraction.
                    </p>
                    <button
                        onClick={handleAnalyze}
                        disabled={!emailContent}
                        className="flex items-center gap-2 bg-blue-600 hover:bg-blue-500 disabled:opacity-50 disabled:cursor-not-allowed text-white px-6 py-2 rounded-lg font-semibold transition-colors"
                    >
                        <Send size={18} /> Run Analysis
                    </button>
                </div>
            </div>

            {status && (
                <div className={`p-4 rounded-lg flex items-center gap-3 ${status.includes('Error') || status.includes('Failed') ? 'bg-red-500/10 text-red-400 border border-red-500/20' : 'bg-green-500/10 text-green-400 border border-green-500/20'}`}>
                    {status.includes('Error') || status.includes('Failed') ? <AlertCircle size={20} /> : <ShieldCheck size={20} />}
                    <span className="text-sm font-medium">{status}</span>
                </div>
            )}
        </div>
    );
}
