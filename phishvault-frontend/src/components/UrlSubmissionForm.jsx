import React, { useState } from 'react';
import { Link } from 'react-router-dom';

export default function UrlSubmissionForm() {
  const [url, setUrl] = useState('');
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState(null);
  const [error, setError] = useState('');
  const [reportId, setReportId] = useState(null);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setResult(null);
    setError('');
    setReportId(null);
    try {
      const res = await fetch('http://localhost:4002/api/scan', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ url })
      });
      const data = await res.json();
      if (res.ok) {
        setResult(data);
        // Assume backend returns { verdict, id }
        setReportId(data.id || data._id); // Use the correct property for your backend
      } else {
        setError(data.error || "Scan failed.");
      }
    } catch {
      setError("Network error.");
    }
    setLoading(false);
  };

  return (
    <form onSubmit={handleSubmit} className="flex flex-col gap-5">
      <input
        type="url"
        placeholder="Enter URL to scan"
        value={url}
        onChange={e => setUrl(e.target.value)}
        required
        className="border border-accent/40 rounded-xl px-5 py-3 bg-dark-bg/80 text-white focus:outline-none focus:ring-2 focus:ring-accent2 text-lg transition-all duration-200 shadow-inner"
      />
      <button
        type="submit"
        disabled={loading}
        className="relative bg-gradient-to-r from-accent2 to-accent hover:from-accent to-accent2 text-white font-bold py-3 rounded-xl transition-all duration-300 shadow-xl hover:scale-105 disabled:opacity-60"
      >
        {loading ? (
          <span className="flex items-center justify-center">
            <svg className="animate-spin h-5 w-5 mr-2 text-white" viewBox="0 0 24 24">
              <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
              <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8z" />
            </svg>
            Scanning...
          </span>
        ) : "Scan"}
      </button>
      {result && (
        <div className={`mt-2 p-4 rounded-xl text-white animate-fade-in-up font-semibold text-lg shadow-xl flex items-center gap-2 ${result.verdict === "phishing" ? "bg-red-600" : "bg-green-600"}`}>
          {result.verdict === "phishing" ? "🚨" : "✅"}
          <span><strong>Verdict:</strong> {result.verdict}</span>
        </div>
      )}
      {error && (
        <div className="mt-2 p-4 rounded-xl bg-red-600 text-white animate-fade-in-up shadow-xl">
          {error}
        </div>
      )}
      {reportId && (
        <Link
          to={`/report/${reportId}`}
          className="inline-block mt-4 px-6 py-3 bg-accent2 text-white font-bold rounded-xl shadow-lg hover:bg-accent transition"
        >
          View Full Report
        </Link>
      )}
    </form>
  );
}