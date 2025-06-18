// routes/reportRoutes.js

import {
  getReportById,
  getAllReports,
  createReport,
  deleteReport,
  updateReport,
  getReportByUrl,
  getReportsByVerdict,
  getReportsByDateRange,
  getReportSummary
} from '../controllers/reportController.js';

export default async function reportRoutes(app) {
  app.post('/api/reports', createReport);                        // Create new report
  app.get('/api/report/:id', getReportById);                     // Get report by ID
  app.get('/api/reports', getAllReports);                        // Get all reports
  app.delete('/api/report/:id', deleteReport);                   // Delete report by ID
  app.put('/api/report/:id', updateReport);                      // Update report
  app.get('/api/report/url/:url', getReportByUrl);               // Get report by URL
  app.get('/api/reports/verdict/:verdict', getReportsByVerdict); // Filter by verdict
  app.get('/api/reports/date', getReportsByDateRange);           // Filter by date range
  app.get('/api/summary', getReportSummary);                     // Summary of all reports
}
