// models/Scan.js

import mongoose from 'mongoose';

const ScanSchema = new mongoose.Schema({
  url: { type: String, required: true },
  screenshot: { type: String },
  redirects: [String],
  logs: [String],
  cookies: [Object],
  verdict: { type: String, enum: ['Safe', 'Suspicious', 'Malicious'], required: true },
  score: { type: Number },
  notes: [String],
  timestamp: { type: Date, default: Date.now }
});

export default mongoose.model('Scan', ScanSchema);
// This code defines a Mongoose schema for the Scan model.
// It includes fields for the URL, screenshot path, redirects, logs, cookies, verdict,
// score, notes, and timestamp.
// The schema is then exported as a Mongoose model named 'Scan'.
// This model will be used to store scan results in the MongoDB database.
// The verdict field is an enum with three possible values: 'Safe', 'Suspicious',
// and 'Malicious', ensuring that only valid verdicts can be stored.
// The timestamp field defaults to the current date and time when a scan is created.
// The model can be used to create, read, update, and delete scan records in the database.
// It will be used in the scanController to save scan results after processing a URL.
// The model will also be used to retrieve scan results when requested by the frontend.