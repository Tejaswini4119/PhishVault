import Scan from '../models/Scan.js';

export const createReport = async (req, reply) => {
  const { url, verdict, notes } = req.body;

  if (!url || !verdict) {
    return reply.code(400).send({ error: 'Missing required fields (url, verdict)' });
  }

  try {
    const newScan = new Scan({
      url,
      verdict,
      notes,
      timestamp: new Date()
    });
    const savedScan = await newScan.save();
    return reply.code(201).send(savedScan);
  } catch (err) {
    console.error('Create Report Error:', err.message);
    return reply.code(500).send({ error: 'Failed to create report' });
  }
};

export const getReportById = async (req, reply) => {
  const { id } = req.params;
  try {
    const report = await Scan.findById(id);
    if (!report) return reply.code(404).send({ error: 'Scan not found' });
    return reply.code(200).send(report);
  } catch (err) {
    return reply.code(500).send({ error: 'Internal server error' });
  }
};

export const getAllReports = async (req, reply) => {
  try {
    const reports = await Scan.find().sort({ timestamp: -1 });
    return reply.code(200).send(reports);
  } catch (err) {
    return reply.code(500).send({ error: 'Internal server error' });
  }
};

export const deleteReport = async (req, reply) => {
  const { id } = req.params;
  try {
    const result = await Scan.findByIdAndDelete(id);
    if (!result) return reply.code(404).send({ error: 'Scan not found' });
    return reply.code(200).send({ message: 'Scan deleted successfully' });
  } catch (err) {
    return reply.code(500).send({ error: 'Internal server error' });
  }
};

export const updateReport = async (req, reply) => {
  const { id } = req.params;
  const { notes, verdict } = req.body;

  if (!notes || !verdict) {
    return reply.code(400).send({ error: 'Missing notes or verdict' });
  }

  try {
    const updatedReport = await Scan.findByIdAndUpdate(
      id,
      { notes, verdict },
      { new: true }
    );
    if (!updatedReport) return reply.code(404).send({ error: 'Scan not found' });
    return reply.code(200).send(updatedReport);
  } catch (err) {
    return reply.code(500).send({ error: 'Internal server error' });
  }
};

export const getReportByUrl = async (req, reply) => {
  const { url } = req.params;

  if (!url) {
    return reply.code(400).send({ error: 'Missing URL' });
  }

  try {
    const report = await Scan.find({ url }).sort({ timestamp: -1 }).limit(1);
    if (!report || report.length === 0) {
      return reply.code(404).send({ error: 'No report found for this URL' });
    }
    return reply.code(200).send(report[0]);
  } catch (err) {
    console.error('Get Report by URL Error:', err.message);
    return reply.code(500).send({ error: 'Failed to retrieve report' });
  }
};

export const getReportsByVerdict = async (req, reply) => {
  const { verdict } = req.params;

  if (!verdict) {
    return reply.code(400).send({ error: 'Missing verdict' });
  }

  try {
    const reports = await Scan.find({ verdict }).sort({ timestamp: -1 });
    if (reports.length === 0) {
      return reply.code(404).send({ error: 'No reports found for this verdict' });
    }
    return reply.code(200).send(reports);
  } catch (err) {
    console.error('Get Reports by Verdict Error:', err.message);
    return reply.code(500).send({ error: 'Failed to retrieve reports' });
  }
};

export const getReportsByDateRange = async (req, reply) => {
  const { startDate, endDate } = req.query;

  if (!startDate || !endDate) {
    return reply.code(400).send({ error: 'Missing start or end date' });
  }

  try {
    const reports = await Scan.find({
      timestamp: {
        $gte: new Date(startDate),
        $lte: new Date(endDate)
      }
    }).sort({ timestamp: -1 });

    if (reports.length === 0) {
      return reply.code(404).send({ error: 'No reports found for this date range' });
    }
    return reply.code(200).send(reports);
  } catch (err) {
    console.error('Get Reports by Date Range Error:', err.message);
    return reply.code(500).send({ error: 'Failed to retrieve reports' });
  }
};

export const getReportSummary = async (req, reply) => {
  try {
    const totalReports = await Scan.countDocuments();
    const safeReports = await Scan.countDocuments({ verdict: 'Safe' });
    const suspiciousReports = await Scan.countDocuments({ verdict: 'Suspicious' });
    const maliciousReports = await Scan.countDocuments({ verdict: 'Malicious' });

    return reply.code(200).send({
      totalReports,
      safeReports,
      suspiciousReports,
      maliciousReports
    });
  } catch (err) {
    console.error('Get Report Summary Error:', err.message);
    return reply.code(500).send({ error: 'Failed to retrieve report summary' });
  }
};
