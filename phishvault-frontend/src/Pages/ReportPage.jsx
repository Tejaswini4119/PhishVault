// src/pages/ReportPage.js
import React, { useEffect, useState } from "react";
import { useParams } from "react-router-dom";

const ReportPage = () => {
  const { id } = useParams();
  const [report, setReport] = useState(null);
  const [error, setError] = useState("");

  useEffect(() => {
    const fetchReport = async () => {
      try {
        const res = await fetch(`http://localhost:4002/api/scan/${id}`);
        if (!res.ok) throw new Error("Report not found");
        const data = await res.json();
        setReport(data);
      } catch (err) {
        setError("Failed to load report.");
      }
    };
    fetchReport();
  }, [id]);

  if (error) return <p className="text-red-500 font-semibold text-center">{error}</p>;
  if (!report) return <p className="text-gray-300 text-center">Loading report...</p>;

  return (
    <div className="min-h-screen bg-gray-950 text-white p-8">
      <div className="max-w-2xl mx-auto bg-gray-900 border border-cyan-500 p-6 rounded-lg shadow-lg">
        <h2 className="text-2xl font-bold mb-4 text-cyan-400">🧾 Full Report</h2>
        <p><strong className="text-gray-300">URL:</strong> {report.url}</p>
        <p><strong className="text-gray-300">Report ID:</strong> {report._id}</p>
        <p>
  <strong className="text-gray-300">Verdict:</strong>{" "}
  <span className={`font-semibold ${
    ["phishing", "malicious", "dangerous"].includes(report.verdict?.toLowerCase())
      ? "text-red-400"
      : report.verdict?.toLowerCase() === "suspicious"
      ? "text-yellow-400"
      : "text-green-400"
  }`}>
    {report.verdict}
  </span>
</p>

        <p><strong className="text-gray-300">Threat Score:</strong> {report.score ?? 'N/A'}</p>
        <p><strong className="text-gray-300">Details:</strong> {report.details || 'No additional details available.'}</p>
        <p><strong className="text-gray-300">Timestamp:</strong> {report.timestamp}</p>

        {report.screenshot && (
  <div className="mt-6">
    <h3 className="text-lg font-semibold text-cyan-300 mb-2">Captured Screenshot:</h3>
    <img
      src={`http://localhost:4002${report.screenshot}`}
      alt="Captured Screenshot"
      className="w-full border border-gray-700 rounded-lg"
    />
  </div>
)}
      </div>
    </div>
  );
};

export default ReportPage;
