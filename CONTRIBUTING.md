
# 🛡️ Contributing to PhishVault

_"Security tools don’t need to be complex — they need to be clear, effective, and built with purpose."_

Thank you for considering contributing to **PhishVault**, a modern phishing detection and threat analysis platform designed with cybersecurity professionals in mind.

We welcome contributions of all types — from code, documentation, testing, and Docker improvements to UI suggestions and bug reports. Whether you’re fixing a typo or implementing new scanning logic, your input is valued.

---

## 📦 Project Overview

**PhishVault** is a fast, secure, and containerized fullstack application that provides:
- Phishing scan orchestration with Puppeteer
- Real-time threat scoring and verdicts
- Full dashboard, API access, and MongoDB analytics
- Docker support and CI/CD workflows

**Stack**:
- Backend: Fastify (Node.js)
- Frontend: React.js
- DB: MongoDB
- Containerization: Docker
- Scanning: Puppeteer

---

## 🧠 How You Can Contribute

You can contribute by:
- 📄 Improving documentation
- 🐛 Reporting bugs
- 🚀 Suggesting or implementing new features
- 🐳 Enhancing Docker setup or CI/CD workflows
- 🎨 Refining frontend UI/UX
- ⚙️ Refactoring backend logic or APIs

---

## 🛠️ Local Development Setup

### 1. Clone the repository:
```bash
git clone https://github.com/your-username/phishvault.git
cd phishvault
```

### 2. Start MongoDB:
If not already running:
```bash
docker run -d --name phishvault-mongo -p 27017:27017 mongo
```

### 3. Configure Environment:
In `backend/.env`:
```
PORT=4000
MONGO_URI=mongodb://localhost:27017/phishvault
```

### 4. Start Backend:
```bash
cd phishvault-backend/
./devbackend.sh
```

### 5. Start Frontend (React):
```bash
cd ../phishvault-frontend/
./devfrontend.sh
```

---

## 🌱 Branch Naming Convention

| Type        | Prefix         | Example                        |
|-------------|----------------|--------------------------------|
| Feature     | `feat/`        | `feat/threat-scorer-v2`        |
| Bugfix      | `fix/`         | `fix/docker-port-4002`         |
| Docker      | `docker/`      | `docker/mongodb-volume-fix`    |
| Refactor    | `refactor/`    | `refactor/scan-controller`     |
| Docs        | `docs/`        | `docs/update-readme`           |


⚠️ **Avoid pushing directly to the `main` branch.** Always create a feature or fix branch and submit a pull request for review before merging.

---

## ✅ Commit Message Format

Use [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/):

```
<type>(scope): short summary
```

Examples:
- `feat(scan): add brand impersonation detection`
- `fix(frontend): render details field correctly`
- `docs(contributing): add setup instructions`

---

## 🐛 Reporting Bugs

1. Check [existing issues](https://github.com/Tejaswini4119/PhishVault/issues).
2. If it’s new, open an issue with:
   - Description
   - Steps to reproduce
   - Logs or screenshots
   - Expected vs. actual behavior

---

## 💡 Suggesting Enhancements

You can request or propose features by opening an issue or PR. Include:
- The problem it solves
- Possible approach
- A mockup or flow if it's UI-related

---

## 🧪 Testing & QA

Before submitting a PR:
- ✅ Run all scripts without crash
- ✅ Test frontend/backend integration
- ✅ Validate with MongoDB and scan controller
- ✅ Check formatting with `eslint` (if set up)

---

## 🤖 Docker Contributions

If you're working on Docker:
- Keep image size minimal (consider multi-stage builds)
- Add healthchecks where needed
- Update `README.md` if build/startup changes
- Test with MongoDB volume and custom networks

---

## 🔐 Security and Ethics

PhishVault is intended **strictly for educational, ethical, and research purposes**.

🚫 Do **not** use this tool for real-world phishing campaigns or malicious behavior.  
Violators will be blocked and reported.

If you find a security vulnerability:
📫 Email us at **security@phishvault.io** or contact the maintainers privately.

---

## 🧑‍💻 Maintainers

- **PardhuVarma** — Backend, Docker & Security Logic  
  _“Functionality, observability, and precision — the foundation of secure backend systems.”_

- **Tejaswini (Teju)** — Project Lead & Frontend  
  _“Every tool should feel intuitive — even when it’s built for complex problems.”_

---

## 🤝 Collaboration Principles

- Clear separation of backend/frontend ownership
- Secure-by-default coding
- Fast iteration cycles (CI/CD via GitLab)
- Documentation-first before pushing breaking changes

---

## 📜 License & Ethics

This project is licensed under MIT and bound by an **ethical usage policy**.

PhishVault stands for responsible security tooling.  
Let’s build defensively, ethically, and intelligently. 💡

---

## 📫 Questions or Support?

- Issues: [GitHub Issues](https://github.com/Tejaswini4119/PhishVault/issues)
- Discussions: GitHub or LinkedIn (PardhuVarma or Teju)
- Contact: Open an issue or message the team

---

_“PhishVault isn’t just a project. It’s proof that even a small, focused team can build something powerful, intuitive, and security-driven.”_

© 2025 PhishVault – Built with intent. Released with purpose.
