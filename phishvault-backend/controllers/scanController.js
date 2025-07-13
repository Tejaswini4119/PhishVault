import puppeteerService from '../services/puppeteerService.js';
import threatScorer from '../services/threatScorer.js';
import { sanitizeUrl } from '../utils/sanitizeUrl.js';
import Scan from '../models/Scan.js';

export const runScan = async (req, reply) => {
  const { url } = req.body;

  if (!url || !/^https?:\/\//.test(url)) {
    return reply.code(400).send({ error: 'Invalid or missing URL' });
  }

  try {
    const safeUrl = sanitizeUrl(url);
    const scanData = await puppeteerService.scanURL(safeUrl);

    // Debug Logs
    console.log("============== SCAN DEBUG LOGS ==============");
    console.log("[Final URL]:", scanData.url);
    console.log("[Redirect Count]:", scanData.redirects.length);
    console.log("[HTML Snippet]:", scanData.html?.slice(0, 500));
    console.log("[Form Action Match]:", scanData.html.match(/<form[^>]+action=[\"']?([^\"'>]+)[\"']?/i));
    console.log("[Password Field Found?]:", /type\s*=\s*["']?password["']?/i.test(scanData.html));
    console.log("[JS Logs Detected]:", scanData.logs.slice(0, 3));
    console.log("[Cookies Count]:", scanData.cookies.length);
    console.log("================================================");

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
      details: scoreResult.details,
      timestamp: new Date()
    });

    return reply.code(200).send({
      scanId: scanRecord._id,
      verdict: scoreResult.verdict,
      score: scoreResult.score,
      notes: scoreResult.notes,
      details: scoreResult.details
    });
  } catch (err) {
    console.error('❌ Scan Error:', err.message);
    return reply.code(500).send({ error: 'Scan failed. Possibly due to browser crash or blocked resources.' });
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
