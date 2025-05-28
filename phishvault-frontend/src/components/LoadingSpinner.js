jsx
import React from 'react';
import { motion } from 'framer-motion';

const LoadingSpinner = () => {
return (
<motion.div
className="flex items-center justify-center mt-6"
initial={{ opacity: 0 }}
animate={{ opacity: 1 }}
>
<motion.div
className="w-8 h-8 border-4 border-blue-500 border-t-transparent rounded-full animate-spin"
animate={{ rotate: 360 }}
transition={{ repeat: Infinity, duration: 1, ease: "linear" }}
/>
<span className="ml-3 text-blue-600 font-medium">Scanning...</span>
</motion.div>
);
};
export default LoadingSpinner;