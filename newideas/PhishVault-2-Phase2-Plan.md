# PhishVault-2 Phase 2 Implementation Plan

## Goal Description
Implement "Deep Intelligence & The Verdict" as defined in the development plan. This phase transforms the system from a simple collector into an intelligent analysis platform by adding Email support, AI analysis, and a Policy-based Decision Engine.

## Assignments

### Developer 1: Tejaswini (Email & Storage)
**Responsibility**: Enhance the system's "Memory" and "Ingestion" capabilities.

1.  **Blob Storage (PhishDB Tier 3)**
    *   **Task**: Setup MinIO (S3-compatible object storage) to handle large artifacts like DOM dumps and Screenshots.
    *   **Output**: `services/storage` service and Docker configuration.

2.  **Email Ingestion Engine**
    *   **Task**: Extend the Ingestion API to accept `.eml` and `.msg` file uploads.
    *   **Task**: Implement parsing logic to extract:
        *   Headers (Subject, From, To, Date)
        *   Body (HTML/Text)
        *   Attachments

3.  **Email Forensic Analysis**
    *   **Task**: Implement header verification logic.
    *   **Checklist**: SPF, DKIM, and DMARC result extraction and validation.

### Developer 2: Pardhu (AI & Decision Engine)
**Responsibility**: Build the "Brain" and enhanced "Eyes" of the system.

1.  **Stealth Browser Engine (Enhanced)**
    *   **Task**: Add evasion capabilities to Playwright to avoid bot detection.
    *   **Features**: WebGL spoofing, User-Agent rotation, Navigator property overwrites.

2.  **Visual AI Engine**
    *   **Task**: Implement a Brand Detection model (Siamese Neural Network or Perceptual Hashing) to compare screenshots against a "Golden Set" of target brands (e.g., login pages of major banks).

3.  **NLP & Text Analysis**
    *   **Task**: Implement basic analysis of the DOM text and Email Body.
    *   **Features**: Keyword extraction, urgency detection (e.g., "Action Required Immediately").

4.  **Verdict Engine V3 (The Brain)**
    *   **Task**: Integrate Open Policy Agent (OPA) with Rego policies.
    *   **Logic**: Combine signals (Visual Score + Domain Age + SPF Fail) to output a final `Malicious` or `Safe` verdict.

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
- **Decision**: Unit tests for Rego policies (input: "Visual match 90%", output: "Malicious").

### Manual Verification
- **Email Flow**: Upload a sample phishing email -> Verify body extract -> Verify SPF check -> Verify data in DB.
- **Verdict Flow**: Submit a URL -> Verify that the final JSON output contains a "Verdict" field with a calculated risk score.
