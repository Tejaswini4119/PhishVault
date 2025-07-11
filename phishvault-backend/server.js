// server.js - PhishVault Backend (Fixed Version)

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
import mongoose from 'mongoose';  // ✅ Use Mongoose instead of MongoClient

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

const fastify = Fastify({ logger: true });

// Environment
const PORT = process.env.PORT || 4000;
const MONGO_URI = process.env.MONGO_URI || 'mongodb://localhost:27017/phishvault'; // ✅ Use 27017 unless you're sure of 27019

// CORS Configuration
// ✅ Allow only your frontend to access the backend
// ✅ Use 'http://localhost:3000' for development
// ✅ Use 'https://your-frontend-domain.com' for production
// ✅ Ensure you have the correct origin set in your frontend's fetch requests    
await fastify.register(cors, {
  origin: 'http://localhost:3000', // Allow only your frontend
  methods: ['GET', 'POST', 'PUT', 'DELETE'], // Optional: restrict methods
  credentials: true               // If you're sending cookies or auth headers
});


// MongoDB Connection using Mongoose
async function connectDB() {
  try {
    await mongoose.connect(MONGO_URI, {
      useNewUrlParser: true,
      useUnifiedTopology: true
    });
    console.log('✅ Connected to MongoDB (via Mongoose)');
  } catch (err) {
    console.error('❌ Mongoose connection failed:', err);
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

// Log registered routes
fastify.addHook('onRoute', (route) => {
  console.log(`[ROUTE] ${route.method} ${route.url}`);
});

// Register Routes
fastify.register(scanRoutes, { prefix: '/api' });
fastify.register(reportRoutes, { prefix: '/api/report' });

// Global Error Handler
fastify.setErrorHandler((error, request, reply) => {
  console.error('Global Error:', error);
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
