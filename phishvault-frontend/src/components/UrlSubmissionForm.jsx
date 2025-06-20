import React, { useState } from 'react';
import axios from 'axios';
import { useNavigate } from 'react-router-dom';
import toast from 'react-hot-toast';

export default function UrlSubmissionForm() {
const [url, setUrl] = useState('');
const [loading, setLoading] = useState(false);
const navigate = useNavigate();

const handleSubmit = async (e) => {
e.preventDefault();
setLoading(true);
try {
  const res = await axios.post('http://localhost:4002/scan', { url });
  toast.success("Scan initiated successfully!");
  navigate(`/report/${res.data.scanId}`);
} catch (err) {
  if (err.response?.status === 400) toast.error("Invalid URL submitted");
  else toast.error("Something went wrong. Try again.");
} finally {
  setLoading(false);
}
};

return (
<form onSubmit={handleSubmit} className="w-full max-w-md bg-white p-6 rounded shadow">
<input
type="text"
placeholder="https://example.com"
value={url}
onChange={(e) => setUrl(e.target.value)}
className="w-full p-3 border border-gray-300 rounded mb-4"
required
/>
<button type="submit" className="w-full bg-blue-600 text-white py-2 rounded hover:bg-blue-700" disabled={loading} >
{loading ? 'Scanning…' : 'Scan URL'}
</button>
</form>
);
}