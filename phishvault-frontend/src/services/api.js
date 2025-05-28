import axios from 'axios';

export const scanURL = async (url) => {
  const response = await axios.post('/api/scan-url', { url });
  return response.data;
};
export const getScanResults = async (scanId) => {
  const response = await axios.get(`/api/scan-results/${scanId}`);
  return response.data;
};
export const getAllScans = async () => {
  const response = await axios.get('/api/scans');
  return response.data;
};
export const deleteScan = async (scanId) => {
  const response = await axios.delete(`/api/scans/${scanId}`);
  return response.data;
};
export const updateScan = async (scanId, data) => {
  const response = await axios.put(`/api/scans/${scanId}`, data);
  return response.data;
};
export const createScan = async (data) => {
  const response = await axios.post('/api/scans', data);
  return response.data;
};
