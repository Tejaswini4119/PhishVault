import React, { useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import axios from "axios";
import VerdictBadge from "../components/VerdictBadge";
import Loader from "../components/Loader";

export default function ReportPage() {
const { id } = useParams();
const [report, setReport] = useState(null);
const [loading, setLoading] = useState(true);
const [error, setError] = useState(false);

useEffect(() => {
axios
.get(http://localhost:4002/api/report/${id})
.then((res) => {
setReport(res.data);
setLoading(false);
})
.catch((err) => {
setError(true);
setLoading(false);
});
}, [id]);

if (loading) return <Loader />;
if (error) return <div className="p-8 text-red-500">Failed to load report.</div>;

return (
<div className="min-h-screen bg-gray-50 p-6">
<h2 className="text-2xl font-semibold mb-4">Scan Report</h2>
<div className="bg-white p-4 rounded shadow">
<p><strong>URL:</strong> {report.url}</p>
<p className="mt-2 flex items-center gap-2">
<strong>Verdict:</strong>
<VerdictBadge verdict={report.verdict} />
</p>
<img
src={http://localhost:4002/${report.screenshot}}
alt="Screenshot"
className="w-full mt-4 border"
/>
<div className="mt-4">
<h3 className="font-semibold">Redirects:</h3>
<ul className="list-disc list-inside text-sm text-gray-600">
{report.redirects.map((url, index) => (
<li key={index}>{url}</li>
))}
</ul>
</div>
<div className="mt-4">
<h3 className="font-semibold">Console Logs:</h3>
<ul className="list-disc list-inside text-sm text-gray-600">
{report.consoleLogs.map((log, index) => (
<li key={index}>{log}</li>
))}
</ul>
</div>
<div className="mt-4">
<h3 className="font-semibold">Cookies:</h3>
<ul className="list-disc list-inside text-sm text-gray-600">
{report.cookies.map((cookie, index) => (
<li key={index}>
{cookie.name} = {cookie.value}
</li>
))}
</ul>
</div>
</div>
</div>
);
}