// services/puppeteerService.js
import puppeteer from 'puppeteer';
import path from 'path';
import fs from 'fs';
import { v4 as uuidv4 } from 'uuid';
import { fileURLToPath } from 'url';
import { dirname } from 'path';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

export default {
  async scanURL(url) {
    const browser = await puppeteer.launch({ headless: 'new' });
    const page = await browser.newPage();

    const logs = [];
    const redirects = [];
    const collectedCookies = [];

    const screenshotName = `${uuidv4()}.png`;
    const screenshotDir = path.join(__dirname, '..', 'screenshots');
    const screenshotPath = path.join(screenshotDir, screenshotName);

    // 🔧 Ensure screenshots directory exists
    if (!fs.existsSync(screenshotDir)) {
      fs.mkdirSync(screenshotDir, { recursive: true });
    }

    page.on('console', msg => logs.push(msg.text()));
    page.on('framenavigated', frame => {
      const url = frame.url();
      if (!redirects.includes(url)) redirects.push(url);
    });

    try {
      await page.goto(url, { waitUntil: 'domcontentloaded', timeout: 10000 });
      await page.screenshot({ path: screenshotPath });

      const context = page.browserContext();
      const cookiesFromContext = await context.cookies();
      collectedCookies.push(...cookiesFromContext);
    } catch (err) {
      logs.push(`[ERROR] ${err.message}`);
    }

    await browser.close();

    return {
  logs,
  redirects,
  cookies: collectedCookies,
  screenshot: `/screenshots/${screenshotName}`
};

  }
};
