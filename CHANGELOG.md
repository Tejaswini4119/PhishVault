# 📜 Changelog

All notable changes to **PhishVault-2** will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [2.0.0-phase3] - 2026-01-18 (Intelligence Layer Complete)

### 🚀 Major Features
- **Atomic Intelligence Architecture**:
    - Introduced **Signal Abstraction Layer (SAL)** (`core/domain/sal.go`) to normalize outputs from all scanning engines using a strictly typed schema.
    - Implemented **Sovereign Engine** design pattern, isolating scanner, AI, and network modules.

- **Advanced Scanning Engine**:
    - **Stealth Browser (v2)**: Replaced Puppeteer with **Playwright**. Added `WebGL` and `Navigator` evasion modules to bypass industrial bot detection.
    - **Smart Fetcher**: Added robust, context-aware HTTP fetching with automated charset detection and redirect unwinding.

- **AI Analysis Pipeline**:
    - **Visual AI**: Implemented Perceptual Hashing (dHash) and **Golden Set** brand registry for detecting visual clones (Zero-Day Brand Impersonation).
    - **Deep NLP**: Added Bayesian Classifier for intent detection (Urgency, Scam, Credentials) and NER (Named Entity Recognition) for organization extraction.
    - **Verdict Engine**: Integrated **Open Policy Agent (OPA)** with Rego policies for deterministic, evidence-based decision making.

- **Campaign Intelligence & Graph**:
    - **Clustering**: Implemented **DBSCAN** spatial clustering to group artifacts by Visual Hash and network proximity.
    - **Graph Projector**: Added logic to project linear SAL results into a property graph structure (`Lure` -> `Relay` -> `Origin`).
    - **Origin Mapping**: Added "Confident Origin Scoring" based on redirect chain depth and infrastructure entropy.

### 🛠 Engineering & Quality
- **Robustness**:
    - Refactored `GraphProjector` to use Dependency Injection (`GraphDatabase` interface), enabling seamless swapping between Mock/Console and Live Neo4j adapters.
    - Standardized logging across `Orchestrator` and `Producer`.
- **Testing**:
    - Achieved 100% Logic Verification for Phase 2/3 modules via `go test ./services/...`.
    - Added comprehensive integration tests for the full Orchestrator pipeline.

---

## [1.0.1] - 2025-07-13 (Legacy PhishVault v1)
*(Legacy features superseded by v2 architecture)*
- Basic Puppeteer Scanning.
- JavaScript-based Threat Scoring (`threatScorer.js`).
- MongoDB-based simple storage.
- React Frontend Dashboard.

---
© 2026 PhishVault Development Team
