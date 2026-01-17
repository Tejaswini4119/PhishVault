# PhishVault-2 Phase 2 Implementation Plan

## Goal Description

Implement "Deep Intelligence & The Verdict" as defined in the development plan. This phase transforms the system from a simple collector into an intelligent analysis platform by adding Email support, AI analysis, and a Policy-based Decision Engine.

## Assignments

### Developer 1: Tejaswini (Email & Storage)

**Responsibility**: Enhance the system's "Memory" and "Ingestion" capabilities.

1. **Blob Storage (PhishDB Tier 3)**
    * **Task**: Setup MinIO (S3-compatible object storage) to handle large artifacts like DOM dumps and Screenshots.
    * **Output**: `services/storage` service and Docker configuration.

2. **Email Ingestion Engine**
    * **Task**: Extend the Ingestion API to accept `.eml` and `.msg` file uploads.
    * **Task**: Implement parsing logic to extract:
        * Headers (Subject, From, To, Date)
        * Body (HTML/Text)
        * Attachments

3. **Email Forensic Analysis**
    * **Task**: Implement header verification logic.
    * **Checklist**: SPF, DKIM, and DMARC result extraction and validation.

### Developer 2: Pardhu (AI & Decision Engine)

**Responsibility**: Build the "Brain" and enhanced "Eyes" of the system.

1. **Stealth Browser Engine (Enhanced)**
    * **Task**: Add evasion capabilities to Playwright to avoid bot detection.
    * **Features**: WebGL spoofing, User-Agent rotation, Navigator property overwrites.

## Phase 2+: Industrial Grade AI & Decision Engine (Upgrade)

### Goal

Refactor the MVP Analysis Service into a production-ready engine.
* **Concurrency**: Parallel analysis pipelines.
* **Deep Inspection**: HTML Structure analysis (Forms, Password fields).
* **Data-Driven**: Load rules/signatures from config files.

### Proposed Changes

#### [NEW] `services/analysis/engine/engine.go`

- Orchestrator that accepts SAL and runs all sub-analyzers in parallel (using goroutines).
* Aggregates results into a comprehensive `RiskReport`.

#### [MODIFY] `services/analysis/ai/nlp.go`

- **Rename to `text_html.go`**.
* Add `AnalyzeHTML(dom string)`:
  * Count `<input type="password">`.
  * Detect obfuscated scripts.
  * Calculate Form-Target domains (does the form post to a external domain?).

#### [MODIFY] `services/analysis/ai/visual.go`

- Add `ColorHistogram` comparison (to detect identical color schemes even if layout shifts).
* Add `LogoDetection` stub (interface for future TensorFlow/OpenCV integration).

#### [MODIFY] `services/analysis/decision/opa.go`

- Return detailed `Evidence` list with the verdict.

## Proposed Changes (File Structure)

### Infrastructure

#### [MODIFY] `deploy/docker-compose.yml`

- Add MinIO service.

### Storage Service (Tejaswini)

#### [NEW] `services/storage/minio.go`

- MinIO client wrapper.

#### [NEW] `services/storage/manager.go`

- High-level artifact saving logic.

### Ingestion Service (Tejaswini)

#### [MODIFY] `services/ingestion/main.go`

- Add `/submit-email` endpoint.

#### [NEW] `services/ingestion/parser/email.go`

- Email parsing logic.

### Analysis Service (Pardhu)

#### [NEW] `services/analysis/decision/opa.go`

- OPA/Rego wrapper.

#### [NEW] `services/analysis/ai/visual.go`

- Image comparison logic.

#### [NEW] `services/analysis/ai/nlp.go`

- Text analysis logic.

## Verification Plan

### Automated Tests

- **Email**: Unit tests for `.eml` parsing (ensure all headers are extracted).
* **Decision**: Unit tests for Rego policies (input: "Visual match 90%", output: "Malicious").

### Manual Verification

- **Email Flow**: Upload a sample phishing email -> Verify body extract -> Verify SPF check -> Verify data in DB.
* **Verdict Flow**: Submit a URL -> Verify that the final JSON output contains a "Verdict" field with a calculated risk score.
