"use client";

import { Settings, Shield, Key, Bell, User } from 'lucide-react';

export default function SettingsPage() {
    return (
        <div className="max-w-4xl mx-auto space-y-8">
            <div>
                <h2 className="text-2xl font-bold text-white flex items-center gap-2">
                    <Settings className="text-slate-400" /> Platform Settings
                </h2>
                <p className="text-slate-400 text-sm">Configure your analyst workbench and system preferences.</p>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                {/* Sidebar Tabs (Simulator) */}
                <div className="space-y-1">
                    <SettingsTab icon={<User size={18} />} label="Account Profile" active />
                    <SettingsTab icon={<Shield size={18} />} label="Security Analysis" />
                    <SettingsTab icon={<Key size={18} />} label="API Credentials" />
                    <SettingsTab icon={<Bell size={18} />} label="Notifications" />
                </div>

                {/* Content Area */}
                <div className="md:col-span-2 space-y-6">
                    <div className="bg-slate-900 border border-slate-800 rounded-xl p-6 space-y-6">
                        <h3 className="font-bold text-white border-b border-slate-800 pb-4">Account Profile</h3>

                        <div className="space-y-4">
                            <div className="grid grid-cols-2 gap-4">
                                <div className="space-y-2">
                                    <label className="text-xs font-semibold text-slate-500 uppercase">Username</label>
                                    <input readOnly value="Analyst-1" className="w-full bg-slate-950 border border-slate-800 rounded-lg p-2 text-slate-300" />
                                </div>
                                <div className="space-y-2">
                                    <label className="text-xs font-semibold text-slate-500 uppercase">Designation</label>
                                    <input readOnly value="Senior SOC Analyst" className="w-full bg-slate-950 border border-slate-800 rounded-lg p-2 text-slate-300" />
                                </div>
                            </div>

                            <div className="space-y-2">
                                <label className="text-xs font-semibold text-slate-500 uppercase">Analysis Threshold</label>
                                <input type="range" className="w-full" min="0" max="100" defaultValue="55" />
                                <div className="flex justify-between text-[10px] text-slate-500">
                                    <span>Sensitive (Aggressive)</span>
                                    <span>Standard (55%)</span>
                                    <span>Conservative</span>
                                </div>
                            </div>
                        </div>

                        <div className="pt-4 flex justify-end">
                            <button className="bg-blue-600 hover:bg-blue-500 text-white px-4 py-2 rounded-lg text-sm font-semibold transition-colors">
                                Save Changes
                            </button>
                        </div>
                    </div>

                    <div className="bg-slate-900 border border-slate-800 rounded-xl p-6">
                        <h3 className="font-bold text-white border-b border-slate-800 pb-4">Detection Engines</h3>
                        <div className="mt-4 space-y-3">
                            <EngineToggle label="Visual AI (Golden Set)" enabled />
                            <EngineToggle label="NLP Deep Intent" enabled />
                            <EngineToggle label="Bayesian Classification" enabled />
                            <EngineToggle label="Threat Intel Proxy" enabled />
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
}

function SettingsTab({ icon, label, active = false }: { icon: React.ReactNode, label: string, active?: boolean }) {
    return (
        <div className={`flex items-center gap-3 px-4 py-3 rounded-lg cursor-pointer transition-colors ${active ? 'bg-blue-600/10 text-blue-400' : 'text-slate-400 hover:bg-slate-800'}`}>
            {icon}
            <span className="text-sm font-medium">{label}</span>
        </div>
    );
}

function EngineToggle({ label, enabled }: { label: string, enabled: boolean }) {
    return (
        <div className="flex justify-between items-center py-2">
            <span className="text-sm text-slate-300">{label}</span>
            <div className={`w-10 h-5 rounded-full relative transition-colors ${enabled ? 'bg-blue-600' : 'bg-slate-700'}`}>
                <div className={`w-3 h-3 bg-white rounded-full absolute top-1 transition-all ${enabled ? 'right-1' : 'left-1'}`} />
            </div>
        </div>
    );
}
