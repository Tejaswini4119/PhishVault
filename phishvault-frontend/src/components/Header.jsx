import { Link } from 'react-router-dom';
export default function Header() {
  return (
    <header className="sticky top-0 z-50 bg-dark-card/80 backdrop-blur-md shadow-lg py-5 px-4 border-b border-accent/20">
      <div className="container mx-auto flex flex-col items-center">
        <h1 className="text-4xl font-black text-accent tracking-widest animate-fade-in-down drop-shadow-lg">PhishVault</h1>
        <p className="text-blue-200 mt-1 animate-fade-in-up font-medium">Phishing Detection & Reporting Platform</p>
      </div>
        <nav className="mt-4">
            <Link to="/" className="text-accent2 hover:underline">Home</Link>
            <Link to="/dashboard" className="text-accent2 hover:underline ml-4">Dashboard</Link>
        </nav>
    </header>
  );
}