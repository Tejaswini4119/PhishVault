<div align="center">
  
# Security Policy
</div>

## 1. Reporting a Vulnerability

**DO NOT** file public GitHub issues for security vulnerabilities.
If you believe you have found a security vulnerability in PhishVault-2 (e.g., Remote Code Execution via artifact processing, SQL Injection, Logic Bypasses), please report it immediately to the **Internal Security Team**.

* **Email**: <security-team@phishvault.internal>
* **Slack**: `#security-incidents` (Urgent)

We will investigate all legitimate reports and do our best to quickly fix the problem.

## 2. Developer Safety Protocol

**ALL Developers and Analysts MUST adhere to the following:**

* **Sandboxing**: Never execute the `services/scanner` module on a personal host machine without proper containerization. It executes untrusted JavaScript.
* **Artifact Handling**: Treat all files in `storage/artifacts/` as toxic. Do not double-click `.html`, `.pdf`, or `.js` files exported from the system on a production Windows/Mac workstation.
* **Egress Controls**: The scanner makes outbound connections to malicious infrastructure. Ensure your IP masking (VPN/Proxy) is active before running live scans.

## 3. Supported Versions

| Version | Supported | Notes |
| :--- | :--- | :--- |
| **v2.x (Phase 3)** | ✅ Yes | Active Development Branch. |
| v1.x (Legacy) | ❌ No | Deprecated. Unsafe for modern threats. |

---
<div align="center">
  
**Confidence in Code, Caution in Execution.**
</div>
