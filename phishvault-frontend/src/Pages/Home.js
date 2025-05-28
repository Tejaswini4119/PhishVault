jsx
import React, { useState } from 'react';
import { motion } from 'framer-motion';
import logo from '../assets/logo.png';
import LoadingSpinner from '../components/LoadingSpinner';

export default function Home() {
const [url, setUrl] = useState('');
const [isLoading, setIsLoading] = useState(false);

const handleSubmit = (e) => {
e.preventDefault();
setIsLoading(true);
setTimeout(() => {
  setIsLoading(false);
  console.log("Scan completed for:", url);
}, 3000);
};
return (
<div className="min-h-screen bg-gradient-to-b from-blue-50 to-white flex flex-col items-center justify-center px-4">
{/* Logo and Title */}
<motion.div
initial={{ opacity: 0, y: -20 }}
animate={{ opacity: 1, y: 0 }}
transition={{ duration: 0.6 }}
className="flex flex-col items-center mb-8"
>
<img src={logo} alt="PhishVault Logo" className="w-16 h-16 mb-4" />
<h1 className="text-4xl font-bold text-blue-800">PhishVault</h1>
<p className="text-lg text-gray-600 mt-2 text-center">
AI-powered phishing detection and secure preview
</p>
</motion.div>

  {/* Input Form */}
  <motion.form
    onSubmit={handleSubmit}
    className="w-full max-w-md flex flex-col gap-4"
    initial={{ opacity: 0, y: 20 }}
    animate={{ opacity: 1, y: 0 }}
    transition={{ delay: 0.4, duration: 0.6 }}
  >
    <motion.input
      whileFocus={{ scale: 1.02, borderColor: '#3B82F6' }}
      type="text"
      placeholder="Enter a suspicious URL"
      value={url}
      onChange={(e) => setUrl(e.target.value)}
      className="p-3 rounded-xl border border-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500 transition"
      required
    />
    <motion.button
      whileHover={{ scale: 1.03 }}
      whileTap={{ scale: 0.97 }}
      type="submit"
      className="bg-blue-600 text-white py-3 rounded-xl font-semibold hover:bg-blue-700 transition"
    >
      Scan URL
    </motion.button>
  </motion.form>

  {/* Show spinner when loading */}
  {isLoading && <LoadingSpinner />}
</div>
);
}