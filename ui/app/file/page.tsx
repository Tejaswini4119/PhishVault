"use client";

import { useState } from 'react';
import { Upload, FileSearch, ShieldCheck, AlertCircle } from 'lucide-react';
import { useAuth } from '@/components/AuthProvider';

export default function FilePage() {
    const [file, setFile] = useState<File | null>(null);
    const [status, setStatus] = useState<string | null>(null);
    const { token } = useAuth();

    const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        if (e.target.files && e.target.files[0]) {
            setFile(e.target.files[0]);
        }
    };

    const handleUpload = async () => {
        if (!file || !token) return;
        setStatus('Uploading and Analyzing...');

        const formData = new FormData();
        formData.append('file', file);

        try {
            const res = await fetch('/api/submit-file', {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${token}`
                },
                body: formData
            });
            if (res.ok) {
                setStatus('File submitted successfully. Static analysis results will appear in Live Scans.');
                setFile(null);
            } else {
                setStatus('Failed to upload file.');
            }
        } catch (err) {
            setStatus('Error: ' + err);
        }
    };

    return (
        <div className="max-w-4xl mx-auto space-y-6">
            <div>
                <h2 className="text-2xl font-bold text-white flex items-center gap-2">
                    <FileSearch className="text-purple-400" /> File Analyzer
                </h2>
                <p className="text-slate-400 text-sm">Upload suspicious attachments for static analysis and URL extraction.</p>
            </div>

            <div className="bg-slate-900 border border-slate-800 rounded-xl p-12 flex flex-col items-center justify-center space-y-6 border-dashed hover:bg-slate-800/50 transition-colors cursor-pointer relative">
                <input
                    type="file"
                    onChange={handleFileChange}
                    className="absolute inset-0 opacity-0 cursor-pointer"
                />
                <div className="bg-slate-800 p-4 rounded-full text-slate-400">
                    <Upload size={40} />
                </div>
                <div className="text-center">
                    <p className="text-lg font-semibold text-white">
                        {file ? file.name : 'Click or drag file to upload'}
                    </p>
                    <p className="text-slate-500 text-sm mt-1">
                        Maximum file size: 10MB (PDF, DOCX, ZIP, EXE)
                    </p>
                </div>
            </div>

            <div className="flex justify-end">
                <button
                    onClick={handleUpload}
                    disabled={!file}
                    className="bg-purple-600 hover:bg-purple-500 disabled:opacity-50 disabled:cursor-not-allowed text-white px-8 py-3 rounded-lg font-semibold transition-colors flex items-center gap-2"
                >
                    <FileSearch size={20} /> Analyze File
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
