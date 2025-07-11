import React from 'react';

export default function VerdictBadge({ verdict }) {
  const badgeColors = {
    Safe: 'bg-green-500',
    Malicious: 'bg-red-500',
    Suspicious: 'bg-yellow-500',
    Unknown: 'bg-gray-500',
  };

  const badgeColor = badgeColors[verdict] || 'bg-gray-400';

  return (
    <span className={`text-white px-2 py-1 rounded text-sm font-medium ${badgeColor}`}>
      {verdict}
    </span>
  );
}
