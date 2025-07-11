import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';

export default function Dashboard() {
  const [scans, setScans] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetch('http://localhost:4002/api/scans')
      .then(res => res.json())
      .then(data => {
        setScans(data);
        setLoading(false);
      });
  }, []);

  return (
    <div className="max-w-4xl mx-auto mt-12 p-6 bg-dark-card/90 rounded-2xl shadow-2xl">
      <h2 className="text-3xl font-bold text-accent mb-6">Scan History</h2>
      {loading ? (
        <div className="text-blue-200">Loading...</div>
      ) : scans.length === 0 ? (
        <div className="text-blue-200">No scans found.</div>
      ) : (
        <table className="w-full text-left text-white">
          <thead>
            <tr className="text-accent2 border-b border-accent/20">
              <th className="py-2">URL</th>
              <th className="py-2">Verdict</th>
              <th className="py-2">Date</th>
              <th className="py-2">Report</th>
            </tr>
          </thead>
          <tbody>
            {scans.map(scan => (
              <tr key={scan._id} className="border-b border-dark-bg/30 hover:bg-dark-bg/40 transition">
                <td className="py-2 break-all text-blue-100">{scan.url}</td>
               <td className={`py-2 font-bold ${
  ["phishing", "malicious", "dangerous"].includes(scan.verdict?.toLowerCase())
    ? "text-red-400"
    : scan.verdict?.toLowerCase() === "suspicious"
    ? "text-yellow-400"
    : "text-green-400"
}`}>
  {scan.verdict}
</td>
                <td className="py-2 text-blue-100">{new Date(scan.timestamp).toLocaleString()}</td>
                <td className="py-2">
                  <Link
                    to={`/report/${scan._id}`}
                    className="text-accent2 underline hover:text-accent"
                  >
                    View
                  </Link>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}