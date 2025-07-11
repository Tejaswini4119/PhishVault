import puppeteerService from '../services/puppeteerService.js';
import threatScorer from '../services/threatScorer.js';
import { sanitizeUrl } from '../utils/sanitizeUrl.js';
import Scan from '../models/Scan.js';

export const runScan = async (req, reply) => {
  const { url } = req.body;

  // ✅ Validate URL
  if (!url || !/^https?:\/\//.test(url)) {
    return reply.code(400).send({ error: 'Invalid or missing URL' });
  }

  try {
    const safeUrl = sanitizeUrl(url);
    const scanData = await puppeteerService.scanURL(safeUrl);

    // ✅ Debug logs to inspect detection behavior
    console.log("============== DEBUG ==============");
    console.log("[HTML Snippet]:", scanData.html?.slice(0, 500));
    console.log("[Detected Form Action]:", scanData.html.match(/<form[^>]+action=["']?([^"'>]+)["']?/i));
    console.log("[Password Field Found?]:", /type\s*=\s*["']?password["']?/i.test(scanData.html));
    console.log("===================================");

    const scoreResult = threatScorer(scanData);

    const scanRecord = await Scan.create({
      url: safeUrl,
      screenshot: scanData.screenshot,
      redirects: scanData.redirects,
      logs: scanData.logs,
      cookies: scanData.cookies,
      verdict: scoreResult.verdict,
      score: scoreResult.score,
      notes: scoreResult.notes,
      timestamp: new Date()
    });

    return reply.code(200).send({
      scanId: scanRecord._id,
      verdict: scoreResult.verdict,
      score: scoreResult.score,
      notes: scoreResult.notes
    });
  } catch (err) {
    console.error('Scan Error:', err.message);
    return reply.code(500).send({ error: 'Scan failed' });
  }
};

export const getScanResult = async (req, reply) => {
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
};