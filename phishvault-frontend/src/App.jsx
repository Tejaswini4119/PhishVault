import { Routes, Route } from 'react-router-dom';
import HomePage from './pages/HomePage';
import ReportPage from './pages/ReportPage';
import { Toaster } from 'react-hot-toast';

function App() {
return (
<>
<Routes>
<Route path="/" element={<HomePage />} />
<Route path="/report/:id" element={<ReportPage />} />
</Routes>
<Toaster position="top-center" />
</>
);
}
export default App;