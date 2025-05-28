import React, { useState } from 'react';

function URLInput({ onScan }) {
  const [url, setUrl] = useState('');

  const handleChange = (e) => {
    setUrl(e.target.value);
  };

  const handleSubmit = (e) => {
    e.preventDefault();
    onScan(url);
  };

  return (
    <form onSubmit={handleSubmit}>
      <input
        type="text"
        value={url}
        onChange={handleChange}
        placeholder="Enter URL to scan"
      />
      <button type="submit">Scan</button>
    </form>
  );
}

export default URLInput;
