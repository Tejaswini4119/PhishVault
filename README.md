# 🛡️ PhishVault

> **"Security tools don’t need to be complex — they need to be clear, effective, and built with purpose."**

---

## 🚀 Project Summary

**PhishVault** is a secure and intelligent platform to analyze, manage, and investigate potentially malicious URLs. Designed with cybersecurity professionals in mind, it simplifies phishing detection, logging, and reporting in a fast and scalable way.

> ⚙️ **Status**: `v1.0 Released`  
> 📆 **Project Duration**: 2 Months  
> 👥 **Team Size**: 2 Developers
> 📁 **Deployment**: Dockerized with CI/CD  
> 🛠 **Stack**: Node.js, React.js, MongoDB, Docker, Puppeteer

---

## 🔍 What Can PhishVault Do?

- Submit suspicious URLs for scanning  
- Retrieve real-time threat verdicts (Safe / Suspicious / Malicious)  
- Filter reports by verdicts and date  
- View full scan history and analytical summaries  
- Perform full CRUD operations on scan reports  
- Gain backend observability and RESTful API access

---

## 💡 Key Features

- 🔐 **Threat Intelligence Storage** (MongoDB Collections)  
- 🎯 **Custom Verdict Classification**  
- 📆 **Date-range Report Querying**  
- 📊 **Dashboard Analytics Module**  
- 🧾 **Report Generation via Aggregation Pipelines**  
- ⚙️ **Fastify REST API**  
- 📦 **Dockerized Container System**  
- 🤖 **Puppeteer-based Scan Orchestration**

---

## 🧰 Tech Stack

| Component | Technology        |
|----------|-------------------|
| Frontend | React.js *(In Progress)* |
| Backend  | Fastify (Node.js) |
| Database | MongoDB           |
| Scanning | Puppeteer         |
| Container| Docker            |

---

## ⚙️ Local Development and Run.
> Note: Docker image support coming soon! stay tuned.

### ⚙️ Requirements

- Node.js `v18+`
- MongoDB (local or Docker)
- Git
- Linux/macOS terminal or WSL (for `.sh` script support)

To run locally
### 1. 🏗️ Clone PhishVault :

```bash
git clone https://github.com/your-username/phishvault.git
cd phishvault
```

### 🧱 2. Start MongoDB
If MongoDB isn’t running locally, you can spin it up using Docker:
```
docker run -d --name phishvault-mongo -p 27017:27017 mongo
```

### 📦 3. Setup Environment Variables
Backend: backend/.env
```
PORT=4000
MONGO_URI=mongodb://localhost:27017/phishvault
```

### 🖥️ 4. Backend (Fastify + Puppeteer)
```bash
cd phishvault-backend/
./devbackend.sh
```

### 🌐 Frontend (React.js)
```bash
cd phishvault-frontend/
./devfrontend.sh
```
> any dev issues, reach us out at [email](mailto:varmacstp25@gmail.com) or [linkedin](https://www.linkedin.com/in/pardhu-sri-rushi-varma-konduru-696886279/) or here in [gitHub](https://github.com/PardhuSreeRushiVarma20060119)

---

## 🧑‍💻 Meet the Team

### 🎨 Tejaswini (Teju) — *Project Lead & Frontend Developer*

> “Every tool should feel intuitive — even when it’s built for complex problems.”

- Defined project scope and vision  
- Designed UI/UX and led frontend dev  
- Planned API integration and user experience  
- Orchestrated team collaboration and progress tracking

### 🛠️ PardhuVarma — *Backend Engineer & Docker Orchestrations Developer*

> “Functionality, observability, and precision — the foundation of secure backend systems.”

- Developed backend logic using Puppeteer  
- Designed verdict classification and threat scoring  
- Built complete REST API with Fastify  
- Engineered MongoDB schema and security layers

---

## 🧭 Roadmap: v1.0 Milestones

- ✅ Backend REST API completed  
- ✅ MongoDB schema integrated  
- ✅ Puppeteer scan orchestration implemented  
- ✅ Verdict classification logic added  
- ✅ Report generation via aggregation pipelines  
- ✅ CRUD support for scan reports  
- ✅ Dashboard analytics module  
- ✅ Docker containerization  
- ✅ CI/CD setup (GitLab)  
- ✅ Final QA, testing, and cleanup  
- ✅ **v1.0 Released**

---

## 🔐 Ethical Disclaimer

PhishVault is intended **strictly for ethical research, threat detection training, and cybersecurity education**.  
⚠️ **Misuse in real-world phishing campaigns is strictly prohibited and not supported by the developers.**

---

## 📫 Contact & Collaboration

For bug reports, contributions, or collaboration requests:

- **PardhuVarma** – [LinkedIn](https://www.linkedin.com/in/pardhu-sri-rushi-varma-konduru-696886279/) | [GitHub](https://github.com/PardhuSreeRushiVarma20060119) | [Email](mailto:varmacstp25@gmail.com)  
- **Tejaswini (Teju)** – [LinkedIn](https://www.linkedin.com/in/kandukoori-tejaswini-765774289/) | [GitHub](https://github.com/Tejaswini4119/)

---

## 🤝 Collaboration Principles

- 🔄 Clear division of frontend & backend roles  
- 🔐 Security-first development methodology  
- ⚡ Rapid prototyping + CI/CD for faster iteration  
- 📦 Scalable architecture with future extensibility in mind

## ℹ️ Information
> Refer [ChangeLog](CHANGELOG.md) for more info.

---

> **“PhishVault isn’t just a project. It’s proof that even a small, focused team can build something powerful, intuitive, and security-driven.”**

© 2025 **PhishVault** – Built with intent. Released with purpose.
