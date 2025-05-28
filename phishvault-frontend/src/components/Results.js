import React from 'react';

function Results({ data }) {
  return (
    <div className="results">
      <h2>Scan Results</h2>
      <p>Phishing Probability: {data.phishingScore}</p>
      <p>Explanation: {data.explanation}</p>
    </div>
  );
}

export default Results;
