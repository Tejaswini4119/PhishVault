// services/puppeteerService.js

import puppeteer from 'puppeteer';
import fs from 'fs';
import path from 'path';
import { v4 as uuid } from 'uuid';

export default {
  async scanURL(url) {
    const browser = await puppeteer.launch({ headless: 'new' });
    const page = await browser.newPage();

    const logs = [], redirects = [], cookies = [];
    const scanId = uuid();
    const screenshotPath = path.join('screenshots', `${scanId}.png`);

    page.on('console', msg => logs.push(msg.text()));
    page.on('framenavigated', frame => {
      if (!redirects.includes(frame.url())) redirects.push(frame.url());
    });

    try {
      await page.goto(url, { waitUntil: 'domcontentloaded', timeout: 10000 });
      await page.screenshot({ path: screenshotPath });
      const cookieData = await page.cookies();
      cookies.push(...cookieData);
    } catch (err) {
      logs.push(`[ERROR] ${err.message}`);
    }

    await browser.close();

    return {
      logs,
      redirects,
      cookies,
      screenshotPath
    };
  }
};
// This service uses Puppeteer to navigate to the provided URL, capture logs, redirects, cookies, and take a screenshot.
// The screenshot is saved in a 'screenshots' directory with a unique filename based on the scan ID.
// The service returns an object containing the logs, redirects, cookies, and the path to