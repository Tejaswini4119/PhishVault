import React, { useEffect, useState } from 'react';

function SandboxPreview({ url }) {
  const [imageUrl, setImageUrl] = useState(null);

  useEffect(() => {
    const fetchImage = async () => {
      const response = await fetch(`/api/sandbox-preview?url=${url}`);
      const imageBlob = await response.blob();
      const imageObjectURL = URL.createObjectURL(imageBlob);
      setImageUrl(imageObjectURL);
    };

    if (url) {
      fetchImage();
    }
  }, [url]);

  return (
    <div className="sandbox-preview">
      {imageUrl ? <img src={imageUrl} alt="Preview" /> : <p>Loading preview...</p>}
    </div>
  );
}

export default SandboxPreview;
