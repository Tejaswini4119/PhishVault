const scrapePhishingSite = require('../utils/scrape.js');

module.exports = async function (fastify, opts) {
  // This is where you define the route for scraping
  fastify.post('/scrape', async (request, reply) => {
    // Log the incoming request body for debugging
    console.log('Received body:', request.body);

    const { url } = request.body;

    // Validate the URL before proceeding with scraping
    if (!url || typeof url !== 'string' || !url.startsWith('http')) {
      console.error('Invalid URL passed:', url);
      return reply.status(400).send({ success: false, message: 'Invalid or missing URL' });
    }

    try {
      // Call scrape function only when a valid URL is passed
      const content = await scrapePhishingSite(url);
      return reply.send({ success: true, content });
    } catch (error) {
      console.error('Error scraping the site:', error);
      return reply.status(500).send({ success: false, message: 'Failed to scrape the site' });
    }
  });
};
