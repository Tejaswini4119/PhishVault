import { Routes, Route } from 'react-router-dom';
import HomePage from './Pages/HomePage.jsx';
import Header from './components/Header.jsx';
import ReportPage from './Pages/ReportPage.jsx';
import Dashboard from './Pages/DashboardPage.jsx';
import './index.css';

function App() {
  return (
    <div className="min-h-screen bg-dark-bg">
      <Header />
      <Routes>
        <Route path="/" element={<HomePage />} />
        <Route path="/dashboard" element={<Dashboard />} />
        <Route path="/report/:id" element={<ReportPage />} />
        {/* Add more routes as needed */}
      </Routes>
    </div>
  );
}

export default App;