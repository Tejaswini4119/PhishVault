// utils/scrape.js
const axios = require('axios');
const cheerio = require('cheerio');

const scrapeWebsite = async (url) => {
  try {
    // Send a GET request to the URL
    const response = await axios.get(url);

    // Load the HTML into cheerio for parsing
    const $ = cheerio.load(response.data);

    // Extract the page title
    const title = $('title').text();

    // Extract meta description (if available)
    const metaDescription = $('meta[name="description"]').attr('content') || 'No description available';

    // Extract all links on the page
    const links = [];
    $('a').each((index, element) => {
      const href = $(element).attr('href');
      if (href) {
        links.push(href);
      }
    });

    // Return a structured response
    return {
      success: true,
      content: {
        title,
        metaDescription,
        links,
      },
    };
  } catch (error) {
    return {
      success: false,
      message: error.message,
    };
  }
};

module.exports = scrapeWebsite;
