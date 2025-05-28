import React, { useState } from 'react';
import URLInput from './components/URLInput';
import Results from './components/Results';
import SandboxPreview from './components/SandboxPreview';
import './styles/App.css';

function App() {
  const [scanResults, setScanResults] = useState(null);
  const [url, setUrl] = useState('');

  const handleScan = async (inputUrl) => {
    setUrl(inputUrl);
    const response = await fetch('/api/scan-url', {
      method: 'POST',
      body: JSON.stringify({ url: inputUrl }),
      headers: {
        'Content-Type': 'application/json',
      },
    });
    const data = await response.json();
    setScanResults(data);
  };

  return (
    <div className="App">
      <h1>PhishVault: Phishing URL Detection</h1>
      <URLInput onScan={handleScan} />
      {scanResults && <Results data={scanResults} />}
      {scanResults && <SandboxPreview url={url} />}
    </div>
  );
}

export default App;
