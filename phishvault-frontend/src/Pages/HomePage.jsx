import React from 'react';
import UrlSubmissionForm from '../components/UrlSubmissionForm';

export default function HomePage() {
return (
<div className="min-h-screen bg-gray-100 flex flex-col items-center justify-center p-6">
<h1 className="text-3xl font-bold mb-2 text-gray-800">PhishVault</h1>
<p className="text-gray-600 mb-4 text-center max-w-md">
Enter a suspicious link and get a verdict with detailed scan results.
</p>
<UrlSubmissionForm />
</div>
);
}