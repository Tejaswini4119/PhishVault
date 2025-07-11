import React from "react";
import { Link } from 'react-router-dom';

export default function Header() {
return (
<header className="bg-white shadow px-6 py-4">
<div className="max-w-6xl mx-auto flex justify-between items-center">
<Link to="/" className="text-2xl font-bold text-blue-600">PhishVault</Link>
<span className = "text-sm text-gray-500">AI Phishing Detection </span>
</div>
</header>
);
}