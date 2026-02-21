"use client";

import { useState } from 'react';
import { Mail, Upload, ShieldCheck, AlertCircle } from 'lucide-react';
import { useAuth } from '@/components/AuthProvider';

export default function EmailPage() {
    const [file, setFile] = useState<File | null>(null);
    const [status, setStatus] = useState<string | null>(null);
    const { token } = useAuth();

    const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        if (e.target.files && e.target.files[0]) {
            setFile(e.target.files[0]);
        }
    };

    const handleAnalyze = async () => {
        if (!file || !token) return;
        setStatus('Uploading and Analyzing...');

        const formData = new FormData();
        formData.append('file', file);

        try {
            console.log('Sending email analysis request to direct backend...', { token: token?.substring(0, 10) + '...' });
            const res = await fetch('http://127.0.0.1:8080/submit-email', {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${token}`
                },
                body: formData
            });
            console.log('Response Status:', res.status);
            if (res.ok) {
                setStatus('Analysis in progress. Check Live Scans for results.');
                setFile(null);
            } else {
                let errorData = '';
                try {
                    errorData = await res.text();
                } catch (e) {
                    errorData = 'Could not read error body';
                }
                console.error('Email submission failed:', res.status, errorData);
                setStatus(`Failed: Server returned ${res.status}. Error: ${errorData.substring(0, 50)}`);
            }
        } catch (err: any) {
            console.error('Email submission network/fetch error:', err);
            setStatus(`Network Error: ${err.message || err}`);
        }
    };

    return (
        <div className="max-w-4xl mx-auto space-y-6">
            <div>
                <h2 className="text-2xl font-bold text-white flex items-center gap-2">
                    <Mail className="text-blue-400" /> Email Verifier
                </h2>
                <p className="text-slate-400 text-sm">Upload raw email content (.eml) for deep forensic analysis.</p>
            </div>

            <div className="bg-slate-900 border border-slate-800 rounded-xl p-12 flex flex-col items-center justify-center space-y-6 border-dashed hover:bg-slate-800/50 transition-colors cursor-pointer relative">
                <input
                    type="file"
                    onChange={handleFileChange}
                    accept=".eml,message/rfc822"
                    className="absolute inset-0 opacity-0 cursor-pointer"
                />
                <div className="bg-slate-800 p-4 rounded-full text-slate-400">
                    <Upload size={40} />
                </div>
                <div className="text-center">
                    <p className="text-lg font-semibold text-white">
                        {file ? file.name : 'Click or drag .eml file to upload'}
                    </p>
                    <p className="text-slate-500 text-sm mt-1">
                        Supports: SPF/DKIM verification, Display Name spoofing detection, Link extraction.
                    </p>
                </div>
            </div>

            <div className="flex justify-end">
                <button
                    onClick={handleAnalyze}
                    disabled={!file}
                    className="bg-blue-600 hover:bg-blue-500 disabled:opacity-50 disabled:cursor-not-allowed text-white px-8 py-3 rounded-lg font-semibold transition-colors flex items-center gap-2"
                >
                    <Mail size={20} /> Run Analysis
                </button>
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
