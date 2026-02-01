"use client";

import { useState } from 'react';
import GraphVisualizer from '@/components/graph/GraphVisualizer';
import { Rocket, Share2 } from 'lucide-react';

export default function GraphPage() {
    const [isLaunched, setIsLaunched] = useState(false);

    if (isLaunched) {
        return (
            <div className="h-[calc(100vh-8rem)] flex flex-col overflow-hidden">
                <GraphVisualizer />
            </div>
        );
    }

    return (
        <div className="h-[calc(100vh-8rem)] flex flex-col items-center justify-center p-8 bg-slate-900/50 rounded-xl border border-slate-800/50 m-4">
            <div className="max-w-md text-center space-y-8">
                <div className="flex justify-center">
                    <div className="w-24 h-24 bg-blue-500/10 rounded-full flex items-center justify-center border border-blue-500/20 shadow-xl shadow-blue-500/5">
                        <Share2 size={48} className="text-blue-400" />
                    </div>
                </div>

                <div className="space-y-4">
                    <h2 className="text-3xl font-bold text-white tracking-tight">Graph Explorer</h2>
                    <p className="text-slate-400 leading-relaxed">
                        Visualize threat infrastructure, traverse node relationships, and inspect artifacts using the powerful Neo4j engine.
                    </p>
                </div>

                <div className="pt-4">
                    <button
                        onClick={() => setIsLaunched(true)}
                        className="group relative inline-flex items-center justify-center px-8 py-3 font-semibold text-white transition-all duration-200 bg-blue-600 rounded-full hover:bg-blue-500 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-600 shadow-lg shadow-blue-600/30 hover:shadow-blue-600/50 hover:-translate-y-0.5"
                    >
                        <span className="mr-2">Launch Explorer</span>
                        <Rocket size={18} className="group-hover:translate-x-1 transition-transform" />
                    </button>
                    <p className="mt-4 text-xs text-slate-500 font-mono">
                        Powered by Neo4j 5.12.0
                    </p>
                </div>
            </div>
        </div>
    );
}
