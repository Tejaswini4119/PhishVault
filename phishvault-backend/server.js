// server.js - PhishVault Backend

import fastifyStatic from '@fastify/static';
import path from 'path';
import { fileURLToPath } from 'url';
import { dirname } from 'path';

import scanRoutes from './routes/scanRoutes.js';
import reportRoutes from './routes/reportRoutes.js';

import dotenv from 'dotenv';
dotenv.config();

import Fastify from 'fastify';
import cors from '@fastify/cors';
import { MongoClient } from 'mongodb';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

const fastify = Fastify({ logger: true });

// Environment
const PORT = process.env.PORT || 4000;
const MONGO_URI = process.env.MONGO_URI || 'mongodb://localhost:27019';
const DB_NAME = 'phishvault';

let db;

// Register CORS
await fastify.register(cors, {
  origin: true,
  methods: ['GET', 'POST', 'PUT', 'DELETE'],
});

// MongoDB Connection
async function connectDB() {
  try {
    const client = await MongoClient.connect(MONGO_URI);
    db = client.db(DB_NAME);
    fastify.decorate('db', db);  // Make db accessible in routes
    console.log('✅ Connected to MongoDB');
  } catch (err) {
    console.error('❌ MongoDB connection failed:', err);
    process.exit(1);
  }
}

// Root health check
fastify.get('/', async () => {
  return { message: `PhishVault backend running on port ${PORT}` };
});

// Static for screenshots
await fastify.register(fastifyStatic, {
  root: path.join(__dirname, 'screenshots'),
  prefix: '/screenshots/',
});

// Route hooks to confirm registration
fastify.addHook('onRoute', (route) => {
  console.log(`[ROUTE] ${route.method} ${route.url}`);
});

// Register Routes
fastify.register(scanRoutes, { prefix: '/api' });        // ✅ Enables POST /api/scan
fastify.register(reportRoutes, { prefix: '/api/report' }); // ✅ Enables GET /api/report/:id

// Global Error Handler
fastify.setErrorHandler((error, request, reply) => {
  console.error('Error:', error);
  reply.status(500).send({ error: 'Internal Server Error' });
});

// Start Server
async function start() {
  await connectDB();
  try {
    await fastify.listen({ port: PORT, host: '0.0.0.0' });
    console.log(`🚀 Server is listening at http://localhost:${PORT}`);
  } catch (err) {
    fastify.log.error(err);
    process.exit(1);
  }
}

start();
