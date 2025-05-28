// server.js - PhishVault Backend
require('dotenv').config();
const fastify = require('fastify')({ logger: true });
const path = require('path');
const cors = require('@fastify/cors');
const { MongoClient } = require('mongodb');

// Environment
const PORT = process.env.PORT || 4000;
const MONGO_URI = process.env.MONGO_URI || 'mongodb://localhost:27018';
const DB_NAME = 'phishvault';

let db;

// Register CORS
fastify.register(cors, {
  origin: true,
  methods: ['GET', 'POST', 'PUT', 'DELETE']
});

// Connect to MongoDB
async function connectDB() {
  try {
    const client = await MongoClient.connect(MONGO_URI);
    db = client.db(DB_NAME);
    console.log('✅ Connected to MongoDB');
  } catch (err) {
    console.error('❌ MongoDB connection failed:', err);
    process.exit(1);
  }
}

// Health check
fastify.get('/', async () => {
  return { message: `PhishVault backend running on port ${PORT}` };
});

// Register route plugins only after DB connection
async function start() {
  await connectDB();

  // Register routes for scraping after DB is connected
  fastify.register(require('./routes/scrapingRoutes'), { db });

  try {
    await fastify.listen({ port: PORT, host: '0.0.0.0' });
    console.log(`🚀 Server is listening at http://localhost:${PORT}`);
  } catch (err) {
    fastify.log.error(err);
    process.exit(1);
  }
}

start();
