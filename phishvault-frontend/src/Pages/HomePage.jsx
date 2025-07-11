import { useState } from 'react';
import UrlSubmissionForm from '../components/UrlSubmissionForm.jsx';
import { useNavigate } from 'react-router-dom';

export default function HomePage() {
  const [showReport, setShowReport] = useState(false);
  const [reportId, setReportId] = useState('');
  const navigate = useNavigate();

  return (
    <main className="relative flex flex-col items-center justify-center min-h-[80vh] overflow-hidden">
      {/* Animated gradient blob */}
      <div className="absolute top-[-10%] left-1/2 -translate-x-1/2 w-[700px] h-[700px] z-0 pointer-events-none">
        <div className="w-full h-full rounded-full bg-gradient-to-br from-accent to-accent2 opacity-30 blur-3xl animate-pulse-slow"></div>
      </div>
      <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[500px] h-[500px] z-0 pointer-events-none">
        <div className="w-full h-full rounded-full bg-gradient-to-br from-accent2 to-accent opacity-20 blur-3xl animate-pulse-slow"></div>
      </div>
      {/* Navigation buttons */}
      <div className="flex gap-4 mb-8 z-10">
        <button
          onClick={() => setShowReport(false)}
          className={`px-6 py-2 rounded-xl font-bold transition ${!showReport ? "bg-accent2 text-white" : "bg-dark-card text-accent2 border border-accent2"}`}
        >
          Scan URL
        </button>
        <button
          onClick={() => setShowReport(true)}
          className={`px-6 py-2 rounded-xl font-bold transition ${showReport ? "bg-accent2 text-white" : "bg-dark-card text-accent2 border border-accent2"}`}
        >
          View Report
        </button>
      </div>
      <div className="bg-dark-card/90 backdrop-blur-2xl rounded-3xl shadow-2xl p-10 w-full max-w-lg z-10 animate-fade-in-up border border-accent/30">
        {!showReport ? (
          <UrlSubmissionForm />
        ) : (
          <form
            onSubmit={e => {
              e.preventDefault();
              if (reportId) navigate(`/report/${reportId}`);
            }}
            className="flex flex-col gap-4"
          >
            <input
              type="text"
              placeholder="Enter Report ID"
              value={reportId}
              onChange={e => setReportId(e.target.value)}
              className="border border-accent/40 rounded-xl px-5 py-3 bg-dark-bg/80 text-white focus:outline-none focus:ring-2 focus:ring-accent2 text-lg transition-all duration-200 shadow-inner"
              required
            />
            <button
              type="submit"
              className="bg-accent2 text-white font-bold py-3 rounded-xl transition-all duration-300 shadow-xl hover:scale-105"
            >
              Go to Report
            </button>
          </form>
        )}
      </div>
      <p className="mt-10 text-blue-200 text-center z-10 animate-fade-in-up">
        Your privacy is protected. URLs are never stored.
      </p>
    </main>
  );
}