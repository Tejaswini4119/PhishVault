"use client";

import { useState } from 'react';
import { Search, Bell, User, Plus, X, Loader2 } from 'lucide-react';

const Header = () => {
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [url, setUrl] = useState('');
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [submitStatus, setSubmitStatus] = useState<'idle' | 'success' | 'error'>('idle');

    const handleNewScan = async (e: React.FormEvent) => {
        e.preventDefault();
        setIsSubmitting(true);
        setSubmitStatus('idle');

        try {
            const res = await fetch('/api/submit', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ url }),
            });
            if (res.ok) {
                setSubmitStatus('success');
                setUrl('');
                // Close modal after 1.5s
                setTimeout(() => {
                    setIsModalOpen(false);
                    setSubmitStatus('idle');
                }, 1500);
            } else {
                setSubmitStatus('error');
            }
        } catch (error) {
            console.error(error);
            setSubmitStatus('error');
        } finally {
            setIsSubmitting(false);
        }
    };

    return (
        <>
            <header className="h-16 bg-slate-900/50 backdrop-blur-md border-b border-slate-800 flex items-center justify-between px-6 fixed top-0 right-0 w-[calc(100%-16rem)] z-10 transition-all">
                {/* Search Bar */}
                <div className="relative w-96">
                    <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-slate-400" size={18} />
                    <input
                        type="text"
                        placeholder="Search scans, domains, or IP addresses..."
                        className="w-full bg-slate-800 text-slate-200 pl-10 pr-4 py-2 rounded-lg border border-slate-700 focus:outline-none focus:border-blue-500 text-sm"
                    />
                </div>

                {/* Right Actions */}
                <div className="flex items-center gap-4">
                    <button
                        onClick={() => setIsModalOpen(true)}
                        className="flex items-center gap-2 bg-blue-600 hover:bg-blue-500 text-white px-4 py-2 rounded-lg text-sm font-medium transition-colors"
                    >
                        <Plus size={16} />
                        New Scan
                    </button>

                    <div className="w-px h-6 bg-slate-700 mx-2"></div>

                    <button className="relative p-2 text-slate-400 hover:text-white transition-colors">
                        <Bell size={20} />
                        <span className="absolute top-2 right-2 w-2 h-2 bg-red-500 rounded-full animate-pulse"></span>
                    </button>

                    <button className="p-1 rounded-full bg-slate-800 border border-slate-700">
                        <User size={20} className="text-slate-300" />
                    </button>
                </div>
            </header>

            {/* Simple Modal */}
            {isModalOpen && (
                <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm">
                    <div className="bg-slate-900 border border-slate-800 rounded-xl w-full max-w-md p-6 shadow-2xl relative">
                        <button
                            onClick={() => setIsModalOpen(false)}
                            className="absolute top-4 right-4 text-slate-500 hover:text-white"
                        >
                            <X size={20} />
                        </button>

                        <h3 className="text-xl font-bold text-white mb-2">Submit New URL</h3>
                        <p className="text-slate-400 text-sm mb-6">Enter a suspicious URL to trigger the analysis pipeline.</p>

                        <form onSubmit={handleNewScan} className="space-y-4">
                            <div>
                                <label className="block text-xs font-medium text-slate-500 mb-1 uppercase">Target URL</label>
                                <input
                                    type="url"
                                    required
                                    value={url}
                                    onChange={(e) => setUrl(e.target.value)}
                                    placeholder="https://suspect-site.com/login"
                                    className="w-full bg-slate-950 text-white px-4 py-3 rounded-lg border border-slate-800 focus:border-blue-500 focus:outline-none"
                                />
                            </div>

                            {submitStatus === 'success' && (
                                <div className="p-3 bg-green-900/20 text-green-400 text-sm rounded-lg flex items-center">
                                    ✅ Submitted successfully! Scan pending...
                                </div>
                            )}

                            {submitStatus === 'error' && (
                                <div className="p-3 bg-red-900/20 text-red-400 text-sm rounded-lg flex items-center">
                                    ❌ Failed to submit. Check backend connection.
                                </div>
                            )}

                            <div className="flex justify-end pt-2">
                                <button
                                    type="button"
                                    onClick={() => setIsModalOpen(false)}
                                    className="mr-3 px-4 py-2 text-slate-400 hover:text-white text-sm font-medium"
                                >
                                    Cancel
                                </button>
                                <button
                                    type="submit"
                                    disabled={isSubmitting}
                                    className="bg-blue-600 hover:bg-blue-500 text-white px-6 py-2 rounded-lg text-sm font-medium flex items-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed"
                                >
                                    {isSubmitting ? <><Loader2 size={16} className="animate-spin" /> Analyzing...</> : 'Analyze Now'}
                                </button>
                            </div>
                        </form>
                    </div>
                </div>
            )}
        </>
    );
};

export default Header;
