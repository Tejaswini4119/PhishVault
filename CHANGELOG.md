# 📜 Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),  
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [1.0.0] – 2025-07-11

### 🎉 First Stable Release

#### 🚀 Features
- Complete **Frontend UI** built with React.js
  - URL submission interface
  - Dashboard analytics with filters (verdict/date)
  - Report list and detailed view page
  - Integrated API calls for scanning and CRUD
- Fully functional **Backend API** (Fastify + MongoDB)
  - Puppeteer-based scanning orchestration
  - Custom logic for threat scoring and verdict classification
  - MongoDB schema with aggregation pipelines for reporting
  - RESTful routes for scan, report, and analytics endpoints
- **Integration Layer**
  - Complete frontend-backend integration with error handling
  - Reusable service layer for API in frontend
  - Consistent report rendering and state management
- **Deployment & DevOps**
  - Dockerized containers (multi-stage build for frontend/backend)
  - GitLab CI/CD Pipeline for continuous deployment
  - Environment-specific configuration support

#### 🛡️ Security & Stability
- Puppeteer scan isolation and sandboxing
- Input validation and sanitization on backend
- Secure API design with clear request/response structure
- MongoDB query hardening for injection protection

#### 📊 Analytics & Reports
- Dashboard view for scan verdict statistics
- Date-range filters for investigative queries
- Detailed report pages with threat metadata

---

## Pre-v1.0 Versions

- 🔧 Internal prototypes, backend logic testing, and UI scaffolding  
- ❌ These versions are no longer supported or maintained


## 🧪 Pre-v1.0 Version Development History

### 🗓️ 17/06/2025 – Backend Core Functionality (by PardhuVarma)
- Implemented POST /api/scan endpoint with URL validation and Puppeteer scan.
- Developed threatScorer.js with scoring and verdict assignment logic.
- Designed MongoDB schema for storing URL scan metadata and results.
- Built GET /api/scan/:scanId endpoint for retrieving scan reports.
- Set up Fastify server with CORS, static file hosting, and modular routes.
- Verified endpoints using Postman – all passed.

### 🗓️ 18/06/2025 – Puppeteer & Scan Intelligence (by PardhuVarma)
- Integrated Puppeteer headless browser logic for dynamic scanning.
- Captured screenshots, console logs, cookies, and redirect chains.
- Connected threat scoring system with the scan engine.
- Extended MongoDB schema with timestamp and structured metadata.
- Added endpoints: GET /api/scans and /api/scans/:verdict.
- Configured static screenshot serving using Fastify plugin.
- All features tested successfully via Postman.

### 🗓️ 18/06/2025 – Frontend Setup & Submission Page (by Tejaswini)
- Bootstrapped the frontend with React + Tailwind CSS.
- Built the submission form with Axios POST request integration.
- Implemented React Router navigation to redirect post-submit.
- Added loading states, toast messages, and error handling.
- UX design focused on minimalism, responsiveness, and clarity.
- Submission page finalized as frontend foundation.

### 🗓️ 19/06/2025 – Componentization & Report Viewer (by Tejaswini)
- Introduced reusable components: `VerdictBadge`, `Loader`.
- Modularized ReportPage with components and improved styling.
- Enhanced error handling and Axios logic for invalid scan IDs.
- Created clean and scalable file structure using `/components`.
- Report page now fully modular and scalable for future features.

### 🗓️ 27/06/2025 – Threat Detection Validation (by PardhuVarma)
- Tested PhishVault against a locally hosted phishing site.
- Features enabled: password detection, external form checker, keyword analysis.
- Successfully flagged phishing page with a score of 10 (Malicious verdict).
- Static heuristic evaluation shown to be effective for phishing detection.
- Captured screenshots and validated API outputs.
- Validated the strength of backend detection logic.

> These pre-release version milestones laid the groundwork for a robust, modular, and intelligent phishing detection platform.


---

© 2025 PhishVault Development Team
