"use client";

import { Info } from 'lucide-react';

export default function GraphPage() {
    return (
        <div className="h-[calc(100vh-8rem)] flex flex-col">
            <div className="mb-6">
                <h2 className="text-2xl font-bold text-white">Graph Explorer</h2>
                <p className="text-slate-400 text-sm">Interactive visualization of threat infrastructure (Powered by Neo4j).</p>
            </div>

            <div className="flex-1 bg-slate-900 border border-slate-800 rounded-xl flex items-center justify-center relative overflow-hidden group">
                <div className="absolute inset-0 bg-[url('https://assets.website-files.com/634681057b887c6f4830fae2/6367dd988b4446b1f3c3ee1c_Graph-p-1080.png')] bg-cover bg-center opacity-20 group-hover:opacity-30 transition-opacity"></div>

                <div className="text-center p-8 bg-slate-950/80 backdrop-blur-sm rounded-2xl border border-slate-800 border-dashed max-w-lg z-10">
                    <Info size={48} className="mx-auto text-blue-500 mb-4" />
                    <h3 className="text-xl font-bold text-white mb-2">Visualization Engine Loading...</h3>
                    <p className="text-slate-400 mb-6">
                        The interactive graph canvas requires WebGL and direct connection to Neo4j Bolt protocol.
                        <br /><br />
                        <span className="text-xs font-mono bg-slate-900 px-2 py-1 rounded border border-slate-800">
                            neo4j://localhost:7687
                        </span>
                    </p>
                    <button className="bg-blue-600 hover:bg-blue-500 text-white px-6 py-3 rounded-lg font-medium transition-colors">
                        Launch Full Screen
                    </button>
                    <p className="text-xs text-slate-500 mt-4">Feature enabled in Analyst Workbench Pro</p>
                </div>
            </div>
        </div>
    );
}
