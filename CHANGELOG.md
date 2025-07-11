# 📜 Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),  
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [1.0.0] – 2025-07-11

### 🎉 Initial Stable Release

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

## Pre-1.0 Versions

- 🔧 Internal prototypes, backend logic testing, and UI scaffolding  
- ❌ These versions are no longer supported or maintained

---

© 2025 PhishVault Development Team
