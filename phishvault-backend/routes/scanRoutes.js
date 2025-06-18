// routes/scanRoutes.js
import Scan from '../models/Scan.js';
import { runScan } from '../controllers/scanController.js';

export default async function scanRoutes(app) {
  app.post('/api/scan', runScan);

  // Uncomment and define getScanResult if needed
  // app.get('/api/scan/:id', getScanResult);

  app.get('/api/scan/:scanId', async (req, reply) => {
    const { scanId } = req.params;

    if (!scanId) {
      return reply.code(400).send({ error: 'Missing scan ID' });
    }

    try {
      const scanRecord = await Scan.findById(scanId);
      if (!scanRecord) {
        return reply.code(404).send({ error: 'Scan not found' });
      }

      return reply.code(200).send(scanRecord);
    } catch (err) {
      console.error('Get Scan Result Error:', err.message);
      return reply.code(500).send({ error: 'Failed to retrieve scan result' });
    }
  });

  app.get('/api/scans', async (req, reply) => {
    try {
      const scans = await Scan.find().sort({ timestamp: -1 });
      return reply.code(200).send(scans);
    } catch (err) {
      console.error('Get Scans Error:', err.message);
      return reply.code(500).send({ error: 'Failed to retrieve scans' });
    }
  });

  app.get('/api/scans/:verdict', async (req, reply) => {
    const { verdict } = req.params;

    if (!verdict) {
      return reply.code(400).send({ error: 'Missing verdict' });
    }

    try {
      const scans = await Scan.find({ verdict }).sort({ timestamp: -1 });
      if (scans.length === 0) {
        return reply.code(404).send({ error: 'No scans found for this verdict' });
      }
      return reply.code(200).send(scans);
    } catch (err) {
      console.error('Get Scans by Verdict Error:', err.message);
      return reply.code(500).send({ error: 'Failed to retrieve scans' });
    }
  });

  app.setErrorHandler((error, request, reply) => {
    console.error('Error:', error);
    reply.status(500).send({ error: 'Internal Server Error' });
  });
}