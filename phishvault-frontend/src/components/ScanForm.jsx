import React, { useState } from "react";
import axios from "axios";
import { useNavigate } from "react-router-dom";

export default function ScanForm() {
const [url, setUrl] = useState("");
const [loading, setLoading] = useState(false);
const [error, setError] = useState(null);
const navigate = useNavigate();

const handleSubmit = async (e) => {
e.preventDefault();
setLoading(true);
setError(null);
try {
  const res = await axios.post("http://localhost:4002/api/scans", { url });
  navigate(`/report/${res.data.scanId}`);
} catch (err) {
  setError("Failed to scan the URL. Please try again.");
} finally {
  setLoading(false);
}
};

return (
<form onSubmit={handleSubmit} className="w-full max-w-md bg-white p-6 rounded shadow">
<input
type="text"
placeholder="Enter URL to scan"
value={url}
onChange={(e) => setUrl(e.target.value)}
className="w-full p-3 border rounded mb-4"
required
/>
<button type="submit" className="w-full bg-blue-600 text-white py-2 rounded hover:bg-blue-700" disabled={loading} >
{loading ? "Scanning..." : "Scan Now"}
</button>
{error && <p className="text-red-500 mt-2">{error}</p>}
</form>
);
}
