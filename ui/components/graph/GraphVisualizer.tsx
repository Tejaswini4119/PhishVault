"use client";

import { useState, useEffect, useCallback, useRef } from 'react';
import dynamic from 'next/dynamic';
import neo4j from 'neo4j-driver';
import { Play, Loader2, Database, Share2, Table, FileText, Code as CodeIcon, Info, ChevronRight, ChevronDown } from 'lucide-react';

// Dynamically import ForceGraph2D to avoid SSR issues
const ForceGraph2D = dynamic(() => import('react-force-graph-2d'), { ssr: false });

const DRIVER_URI = "neo4j://localhost:7687";
const DRIVER_USER = "neo4j";
const DRIVER_PASSWORD = "password";

type ViewMode = 'GRAPH' | 'TABLE' | 'TEXT' | 'CODE';

export default function GraphVisualizer() {
    const [query, setQuery] = useState("MATCH (n) OPTIONAL MATCH (n)-[r]-() RETURN n, r LIMIT 50");
    const [viewMode, setViewMode] = useState<ViewMode>('GRAPH');

    // Data States
    const [graphData, setGraphData] = useState<{ nodes: any[]; links: any[] }>({ nodes: [], links: [] });
    const [rawRecords, setRawRecords] = useState<any[]>([]);
    const [resultSummary, setResultSummary] = useState<any>(null);
    const [summary, setSummary] = useState<{ nodes: number; relationships: number; labels: Record<string, number>; types: Record<string, number> }>({
        nodes: 0,
        relationships: 0,
        labels: {},
        types: {}
    });

    const [loading, setLoading] = useState(false);
    const [error, setError] = useState("");

    // Graph Refs
    const [dimensions, setDimensions] = useState({ width: 0, height: 0 });
    const graphRef = useRef<any>(null);
    const containerRef = useRef<HTMLDivElement>(null);

    // Resize Observer
    useEffect(() => {
        if (!containerRef.current) return;
        const resizeObserver = new ResizeObserver((entries) => {
            for (const entry of entries) {
                const { width, height } = entry.contentRect;
                setDimensions({ width, height });
            }
        });
        resizeObserver.observe(containerRef.current);
        return () => resizeObserver.disconnect();
    }, []);

    const runQuery = useCallback(async () => {
        setLoading(true);
        setError("");
        setResultSummary(null);

        const driver = neo4j.driver(DRIVER_URI, neo4j.auth.basic(DRIVER_USER, DRIVER_PASSWORD));
        const session = driver.session();

        try {
            const result = await session.run(query);

            // Store raw records with Neo4j types intact slightly processed for display
            const records = result.records.map(r => {
                const obj: any = {};
                r.keys.forEach(key => {
                    obj[key] = r.get(key); // Keep raw Neo4j objects
                });
                return obj;
            });
            setRawRecords(records);
            setResultSummary(result.summary);

            // Process for Graph & Summary
            const nodesMap = new Map();
            const links: any[] = [];
            const labelsCount: Record<string, number> = {};
            const typesCount: Record<string, number> = {};

            result.records.forEach(record => {
                record.keys.forEach(key => {
                    const item = record.get(key);

                    if (neo4j.isNode(item)) { // Node
                        const id = item.identity.toString();
                        if (!nodesMap.has(id)) {
                            nodesMap.set(id, {
                                id: id,
                                label: item.labels[0] || "Node",
                                ...item.properties,
                                color: getNodeColor(item.labels[0])
                            });

                            // Summary
                            item.labels.forEach((l: string) => {
                                labelsCount[l] = (labelsCount[l] || 0) + 1;
                            });
                        }
                    } else if (neo4j.isRelationship(item)) { // Relationship
                        const start = item.start.toString();
                        const end = item.end.toString();
                        links.push({
                            source: start,
                            target: end,
                            label: item.type
                        });

                        // Summary
                        typesCount[item.type] = (typesCount[item.type] || 0) + 1;
                    }
                });
            });

            setGraphData({
                nodes: Array.from(nodesMap.values()),
                links: links
            });

            setSummary({
                nodes: nodesMap.size,
                relationships: links.length,
                labels: labelsCount,
                types: typesCount
            });

        } catch (err: any) {
            console.error("Graph Error:", err);
            setError(err.message || "Failed to execute query");
        } finally {
            await session.close();
            await driver.close();
            setLoading(false);
        }
    }, [query]);

    // Initial load
    useEffect(() => {
        runQuery();
    }, []);

    return (
        <div className="flex flex-col h-[calc(100vh-140px)] bg-slate-900 border border-slate-800 rounded-xl overflow-hidden shadow-2xl">

            {/* 1. Query Bar */}
            <div className="p-3 bg-slate-950 border-b border-slate-800 flex gap-3 items-center z-20 shadow-md">
                <div className="flex-1 relative group">
                    <Database className="absolute left-3 top-2.5 text-slate-500 group-focus-within:text-blue-500 transition-colors" size={16} />
                    <input
                        type="text"
                        value={query}
                        onChange={(e) => setQuery(e.target.value)}
                        className="w-full bg-slate-900 border border-slate-700 rounded-md pl-10 pr-16 py-2 text-sm font-mono text-slate-200 focus:outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500 transition-all placeholder:text-slate-600"
                        placeholder="Enter Cypher Query..."
                    />
                    <kbd className="absolute right-3 top-2.5 text-xs text-slate-600 font-mono border border-slate-700 rounded px-1 hidden sm:inline-block">Enter</kbd>
                </div>
                <button
                    onClick={runQuery}
                    disabled={loading}
                    className="bg-blue-600 hover:bg-blue-500 text-white px-5 py-2 rounded-md flex items-center gap-2 font-medium disabled:opacity-50 disabled:cursor-not-allowed transition-all shadow-lg shadow-blue-900/20 active:scale-95"
                >
                    {loading ? <Loader2 className="animate-spin" size={16} /> : <Play size={16} fill="currentColor" />}
                    Run
                </button>
            </div>

            <div className="flex flex-1 overflow-hidden relative">

                {/* 2. Sidebar Navigation */}
                <div className="w-14 bg-slate-950 border-r border-slate-800 flex flex-col items-center py-4 gap-4 z-10">
                    <NavButton icon={<Share2 size={20} />} active={viewMode === 'GRAPH'} onClick={() => setViewMode('GRAPH')} label="Graph" />
                    <NavButton icon={<Table size={20} />} active={viewMode === 'TABLE'} onClick={() => setViewMode('TABLE')} label="Table" />
                    <NavButton icon={<FileText size={20} />} active={viewMode === 'TEXT'} onClick={() => setViewMode('TEXT')} label="Text" />
                    <NavButton icon={<CodeIcon size={20} />} active={viewMode === 'CODE'} onClick={() => setViewMode('CODE')} label="Code" />
                </div>

                {/* 3. Main Content Area */}
                <div className="flex-1 relative bg-slate-900 overflow-hidden flex flex-col" ref={containerRef}>

                    {error && (
                        <div className="absolute top-4 left-4 right-4 z-50 bg-red-500/10 border border-red-500/50 text-red-400 p-3 rounded backdrop-blur-md text-sm flex items-center gap-2 shadow-xl">
                            <Info size={16} /> {error}
                        </div>
                    )}

                    {/* View: Graph */}
                    {viewMode === 'GRAPH' && dimensions.width > 0 && (
                        <ForceGraph2D
                            ref={graphRef}
                            width={dimensions.width}
                            height={dimensions.height}
                            graphData={graphData}
                            nodeLabel="label"
                            nodeAutoColorBy="label"
                            backgroundColor="#0f172a" // Slate-900 matches sidebar tone
                            linkColor={() => "#475569"} // Slate-600
                            linkDirectionalArrowLength={3.5}
                            linkDirectionalArrowRelPos={1}
                            nodeCanvasObject={(node: any, ctx, globalScale) => {
                                const label = node.url || node.domain || node.ip || node.label || node.id;
                                const fontSize = 12 / globalScale;
                                ctx.font = `${fontSize}px Sans-Serif`;
                                const textWidth = ctx.measureText(label).width;
                                const bckgDimensions = [textWidth, fontSize].map(n => n + fontSize * 0.2);

                                ctx.fillStyle = 'rgba(15, 23, 42, 0.8)'; // Slate-900 background
                                ctx.fillRect(node.x - bckgDimensions[0] / 2, node.y - bckgDimensions[1] / 2, bckgDimensions[0], bckgDimensions[1]);

                                ctx.textAlign = 'center';
                                ctx.textBaseline = 'middle';
                                ctx.fillStyle = node.color || '#fff';
                                ctx.fillText(label, node.x, node.y);

                                ctx.beginPath();
                                ctx.arc(node.x, node.y, 3, 0, 2 * Math.PI, false);
                                ctx.fill();
                            }}
                        />
                    )}

                    {/* View: Table */}
                    {viewMode === 'TABLE' && (
                        <div className="flex-1 overflow-auto p-4 bg-slate-900">
                            {rawRecords.length > 0 ? (
                                <table className="w-full text-left border-collapse text-sm">
                                    <thead>
                                        <tr className="border-b border-slate-700">
                                            {Object.keys(rawRecords[0]).map(key => (
                                                <th key={key} className="py-2 px-4 text-slate-400 font-medium font-mono uppercase bg-slate-950/50 sticky top-0">{key}</th>
                                            ))}
                                        </tr>
                                    </thead>
                                    <tbody className="divide-y divide-slate-800 font-mono text-slate-300">
                                        {rawRecords.map((row, i) => (
                                            <tr key={i} className="hover:bg-slate-800/50 transition-colors">
                                                {Object.values(row).map((val: any, j) => (
                                                    <td key={j} className="py-2 px-4 align-top">
                                                        {renderTableCell(val)}
                                                    </td>
                                                ))}
                                            </tr>
                                        ))}
                                    </tbody>
                                </table>
                            ) : <EmptyState />}
                        </div>
                    )}

                    {/* View: Text */}
                    {viewMode === 'TEXT' && (
                        <div className="flex-1 overflow-auto p-4 bg-slate-950 font-mono text-xs text-green-400 leading-normal">
                            <pre className="whitespace-pre">{generateAsciiTable(rawRecords)}</pre>
                        </div>
                    )}

                    {/* View: Code */}
                    {viewMode === 'CODE' && (
                        <div className="flex-1 overflow-auto bg-[#1e1e1e] text-slate-300 font-mono text-sm leading-relaxed">
                            <div className="border-b border-slate-700 bg-slate-800/50 px-4 py-2 text-xs text-slate-400 flex gap-8">
                                <div>
                                    <span className="font-bold text-slate-200">Server version</span>
                                    <div className="text-slate-300">{resultSummary?.server.version || "Unknown"}</div>
                                </div>
                                <div>
                                    <span className="font-bold text-slate-200">Server address</span>
                                    <div className="text-slate-300">{resultSummary?.server.address || "localhost:7687"}</div>
                                </div>
                            </div>

                            <DetailsGroup title="Query" content={query} open />
                            <DetailsGroup title="Summary" content={JSON.stringify(resultSummary?.updateStatistics || {}, null, 2)} />
                            <DetailsGroup
                                title="Response"
                                content={JSON.stringify(formatRawRecordsForJson(rawRecords), null, 2)}
                                open
                            />
                        </div>
                    )}

                </div>

                {/* 4. Overview Panel (Collapsible) */}
                <div className="w-64 bg-slate-950 border-l border-slate-800 p-4 overflow-y-auto hidden lg:block">
                    <h3 className="text-xs font-bold text-slate-500 uppercase tracking-widest mb-4">Overview</h3>

                    <div className="space-y-6">
                        <StatGroup title="Nodes" count={summary.nodes} items={summary.labels} color="text-blue-400" />
                        <StatGroup title="Relationships" count={summary.relationships} items={summary.types} color="text-pink-400" />
                    </div>
                </div>

            </div>
            <div className="p-2 bg-slate-950 text-xs text-slate-500 border-t border-slate-800 flex justify-between select-none">
                <span className="flex gap-4">
                    <span>Server: <span className="text-green-500">{resultSummary?.server?.address || "localhost:7687"}</span></span>
                    <span>Version: <span className="text-slate-400">{resultSummary?.server?.version || "Neo4j"}</span></span>
                </span>
                <span>Took {resultSummary ? resultSummary.resultAvailableAfter + "ms" : "..."}</span>
            </div>
        </div>
    );
}

// --- Helper Components & Functions ---

function DetailsGroup({ title, content, open = false }: { title: string, content: string, open?: boolean }) {
    const [isOpen, setIsOpen] = useState(open);
    return (
        <div className="border-b border-slate-700">
            <button
                onClick={() => setIsOpen(!isOpen)}
                className="w-full flex items-center gap-2 px-4 py-2 hover:bg-slate-800 text-slate-200 font-bold select-none"
            >
                {isOpen ? <ChevronDown size={14} /> : <ChevronRight size={14} />}
                {title}
            </button>
            {isOpen && (
                <div className="px-8 py-2 pb-4 text-slate-300">
                    <pre className="whitespace-pre-wrap">{content}</pre>
                </div>
            )}
        </div>
    );
}

function NavButton({ icon, active, onClick, label }: any) {
    return (
        <button
            onClick={onClick}
            title={label}
            className={`p-2 rounded-lg transition-all ${active ? 'bg-blue-600 text-white shadow-lg shadow-blue-900/50' : 'text-slate-500 hover:text-slate-300 hover:bg-slate-900'}`}
        >
            {icon}
        </button>
    );
}

function StatGroup({ title, count, items, color }: any) {
    return (
        <div>
            <div className="flex justify-between items-center mb-2">
                <span className="text-sm font-semibold text-slate-200">{title}</span>
                <span className="text-xs bg-slate-900 px-2 py-0.5 rounded-full text-slate-400 font-mono">{count}</span>
            </div>
            <div className="space-y-1">
                {Object.entries(items).map(([name, val]: any) => (
                    <div key={name} className="flex justify-between items-center text-xs group cursor-pointer hover:bg-slate-900 p-1 rounded">
                        <div className="flex items-center gap-2">
                            <div className={`w-2 h-2 rounded-full ${color}`}></div>
                            <span className="text-slate-400 group-hover:text-slate-200 transition-colors">{name}</span>
                        </div>
                        <span className="font-mono text-slate-500">{val}</span>
                    </div>
                ))}
            </div>
        </div>
    );
}

function EmptyState() {
    return (
        <div className="flex flex-col items-center justify-center h-full text-slate-500">
            <Database size={48} className="mb-4 opacity-20" />
            <p>No records found</p>
        </div>
    );
}

// Format Helper for JSON View
function formatRawRecordsForJson(records: any[]) {
    // Convert Neo4j objects to clean JSON structure expected by user
    return records.map(row => {
        const cleanRow: any = {};
        Object.keys(row).forEach(key => {
            const val = row[key];
            if (neo4j.isNode(val)) {
                cleanRow[key] = {
                    identity: val.identity.toNumber(),
                    labels: val.labels,
                    properties: val.properties,
                    elementId: val.elementId
                };
            } else if (neo4j.isRelationship(val)) {
                cleanRow[key] = {
                    identity: val.identity.toNumber(),
                    start: val.start.toNumber(),
                    end: val.end.toNumber(),
                    type: val.type,
                    properties: val.properties,
                    elementId: val.elementId,
                    startNodeElementId: val.startNodeElementId,
                    endNodeElementId: val.endNodeElementId
                }
            } else {
                cleanRow[key] = val;
            }
        });
        return cleanRow;
    });
}


// Helper to render complex objects in table to look like Neo4j Browser
function renderTableCell(val: any) {
    if (val && typeof val === 'object') {
        if (neo4j.isNode(val)) {
            return (
                <div className="font-mono text-xs">
                    <div className="text-slate-500">&#123;</div>
                    <div className="pl-4">
                        <span className="text-blue-400">"identity"</span>: <span className="text-yellow-300">{val.identity.toString()}</span>,
                    </div>
                    <div className="pl-4">
                        <span className="text-blue-400">"labels"</span>: <span className="text-green-400">[{val.labels.map((l: string) => `"${l}"`).join(', ')}]</span>,
                    </div>
                    <div className="pl-4">
                        <span className="text-blue-400">"properties"</span>: &#123;
                        {Object.entries(val.properties).map(([k, v]: any) => (
                            <div key={k} className="pl-4">
                                <span className="text-sky-300">"{k}"</span>: <span className="text-orange-300">{JSON.stringify(v)}</span>,
                            </div>
                        ))}
                        &#125;
                    </div>
                    <div className="text-slate-500">&#125;</div>
                </div>
            );
        }
        if (neo4j.isRelationship(val)) {
            return (
                <div className="font-mono text-xs">
                    <div className="text-slate-500">&#123;</div>
                    <div className="pl-4">
                        <span className="text-blue-400">"identity"</span>: <span className="text-yellow-300">{val.identity.toString()}</span>,
                    </div>
                    <div className="pl-4">
                        <span className="text-blue-400">"type"</span>: <span className="text-pink-400">"{val.type}"</span>,
                    </div>
                    <div className="pl-4">
                        <span className="text-blue-400">"start"</span>: <span className="text-yellow-300">{val.start.toString()}</span>,
                    </div>
                    <div className="pl-4">
                        <span className="text-blue-400">"end"</span>: <span className="text-yellow-300">{val.end.toString()}</span>,
                    </div>
                    <div className="pl-4">
                        <span className="text-blue-400">"properties"</span>: &#123;
                        {Object.entries(val.properties).map(([k, v]: any) => (
                            <div key={k} className="pl-4">
                                <span className="text-sky-300">"{k}"</span>: <span className="text-orange-300">{JSON.stringify(v)}</span>,
                            </div>
                        ))}
                        &#125;
                    </div>
                    <div className="text-slate-500">&#125;</div>
                </div>
            );
        }
        return JSON.stringify(val);
    }
    return String(val);
}

// Pseudo-ASCII Table Gen (Text View)
function generateAsciiTable(data: any[]) {
    if (!data.length) return "No data";

    // Convert rows to their string representations first
    const rows = data.map(row => {
        const strRow: any = {};
        Object.keys(row).forEach(k => {
            const val = row[k];
            if (neo4j.isNode(val)) {
                // Approximate props display: {key: "value", ...}
                const props = JSON.stringify(val.properties)
                    .replace(/^{/, '{')
                    .replace(/}$/, '}');
                strRow[k] = `(:${val.labels[0]} ${props})`;
            } else if (neo4j.isRelationship(val)) {
                strRow[k] = `[:${val.type}]`;
            } else {
                strRow[k] = JSON.stringify(val);
            }
        });
        return strRow;
    });

    const keys = Object.keys(rows[0]);
    // Calculate max width per column (cap at 60 chars to prevent massive overflow)
    const widths = keys.reduce((acc: any, key) => {
        const maxContent = Math.max(
            key.length,
            ...rows.map(r => Math.min((r[key] || "").length, 60)) // Cap width
        );
        acc[key] = maxContent;
        return acc;
    }, {});

    // Box Drawing Characters
    const TLC = '╒'; const TRC = '╕'; const H_DOUBLE = '═'; const T_DOWN_DOUBLE = '╤';
    const V_SINGLE = '│';
    const MLC = '╞'; const MRC = '╡'; const CROSS_DOUBLE = '╪';
    const BLC = '└'; const BRC = '┘'; const T_UP_SINGLE = '┴'; const H_SINGLE = '─'; const CROSS_SINGLE = '┼';

    // Top Border
    let output = TLC + keys.map(k => H_DOUBLE.repeat(widths[k] + 2)).join(T_DOWN_DOUBLE) + TRC + "\n";

    // Header
    output += V_SINGLE + keys.map(k => ` ${k.padEnd(widths[k])} `).join(V_SINGLE) + V_SINGLE + "\n";

    // Header Separator
    output += MLC + keys.map(k => H_DOUBLE.repeat(widths[k] + 2)).join(CROSS_DOUBLE) + MRC + "\n";

    // Rows
    rows.forEach((row, i) => {
        output += V_SINGLE + keys.map(k => {
            let val = row[k] || "";
            if (val.length > 60) val = val.substring(0, 57) + "..."; // Truncate for display
            return ` ${val.padEnd(widths[k])} `;
        }).join(V_SINGLE) + V_SINGLE + "\n";

        // Internal Separator (Single line) - Optional, similar to Neo4j
        if (i < rows.length - 1) {
            output += '├' + keys.map(k => H_SINGLE.repeat(widths[k] + 2)).join(CROSS_SINGLE) + '┤' + "\n";
        }
    });

    // Bottom Border
    output += BLC + keys.map(k => H_SINGLE.repeat(widths[k] + 2)).join(T_UP_SINGLE) + BRC;

    return output;
}

function getNodeColor(label: string) {
    switch (label) {
        case 'ScanArtifact': return '#ef4444';
        case 'Domain': return '#3b82f6';
        case 'IP': return '#10b981';
        case 'Artifact': return '#f59e0b';
        default: return '#a8a29e';
    }
}
