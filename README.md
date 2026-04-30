# BetKZ — Online Sportsbook Website

## Project Overview
BetKZ is a web-based sportsbook platform developed as an academic full-stack project.

The goal of the project is to design a fast, simple, and well-structured betting website that demonstrates how complex platforms can be improved through clean architecture and user-focused interface design.

This project focuses on technical implementation and usability rather than commercial deployment.

---

## Problem Statement
Many existing sportsbook websites have common issues that negatively affect user experience:

- Overloaded interfaces with excessive menus and advertisements
- Slow page loading and poor mobile optimization
- Complicated betting workflows
- Difficult navigation for new users

These problems often make platforms hard to use, especially for casual users.

---

## Global Competitor Comparison
The sportsbook market includes a variety of well-known global platforms, such as Bet365, William Hill, 1xBet, and others.  
While these platforms are established and feature-rich, they share common issues that affect usability and accessibility.

Below is a high-level comparison of typical strengths and weaknesses seen in leading sportsbook platforms:

| Competitor | Strengths | Weaknesses |
|------------|-----------|------------|
| Bet365 | Extensive live markets, advanced analytics | Complex navigation, high feature overload |
| William Hill | Long market presence, diverse markets | Outdated interface, less optimized on mobile |
| 1xBet | Huge selection of events and markets | Aggressive ads, confusing UI |
| Regional platforms | Local focus, language support | Slow performance, limited feature polish |

### How BetKZ Differs
BetKZ is not intended to compete feature-by-feature with large global bookies.  
Instead, BetKZ focuses on:

- Simplified experience — fewer unnecessary components, more clarity  
- Faster core workflows — browse events, place bets, track results  
- Usability-first design — minimal learning curve for new users  
- Responsive layout — works well on mobile and desktop without heavy UI elements

This approach makes the platform more accessible for casual users and provides a useful case study on how UX and architecture affect real-world systems.

### Key Insight
Most competitors prioritize feature quantity, while usability and simplicity are often overlooked.

## Deployment and Monitoring

This repository now includes:

- ✅ Containerized deployment with Docker Compose in `deployments/docker-compose.yml`
- ✅ Prometheus monitoring and Grafana dashboard provisioning
- ✅ Backend metrics exposed through `/metrics`
- ✅ Terraform infrastructure provisioning scaffold in `terraform/`
- ✅ Operational reports in `docs/incident_report.md`, `docs/postmortem.md`, and `docs/terraform_report.md`
- ✅ Incident response simulation with documented failure and recovery

**Current Status**: All services are running successfully with monitoring enabled.

See `docs/deployment_guide.md` for step-by-step local deployment instructions.

---

## Project Advantages
BetKZ addresses these weaknesses with a different approach:

- Minimal and clean interface focused on essential actions
- Clear user flow: Browse → Bet → Track → Withdraw
- Responsive design optimized for both desktop and mobile
- Simple onboarding with intuitive navigation
- MVP-focused scope without unnecessary complexity

The project demonstrates that improving usability does not require adding more features.

---

## Project Objectives
The main objectives of this project are:

- To analyze existing sportsbook platforms and identify usability problems
- To design a simplified and structured betting workflow
- To implement a client–server architecture
- To practice frontend and backend integration
- To build a scalable MVP using modern web technologies

---

## Target Users
The platform is designed for:

- Sports fans aged 18–35
- Casual bettors interested in football, boxing, and e-sports

The system intentionally avoids advanced professional tools to keep the interface simple and accessible.

---

## Core Functionality (MVP)

### User Flow
Browse → Bet → Track → Withdraw

### User Features
- User registration and authentication
- Sports event browsing
- Single and combined bets
- Balance management
- Bet history and notifications

### Administrative Features
- Event and match management
- Odds configuration
- User activity monitoring

---

## System Architecture
The platform follows a client–server architecture:

- The frontend handles user interaction and presentation
- The backend processes business logic and data management
- Communication is implemented using REST APIs

This architecture allows easy extension in future versions.

---

## Technology Stack
- Frontend: React
- Backend: Go

These technologies were chosen for performance, scalability, and real-world applicability.

---

## Project Scope
This project represents a Minimum Viable Product (MVP).

Only core features required to demonstrate the idea and technical approach are implemented.  
Commercial deployment, real financial integrations, and regulatory aspects are outside the project scope.

---

## Development Timeline
- Weeks 1–4: UI design, authentication, betting pages
- Weeks 5–8: backend logic, admin panel
- Week 9+: testing and MVP completion

---

## Conclusion
BetKZ demonstrates how a sportsbook platform can be redesigned with a focus on usability, simplicity, and clean system architecture.

The project serves as a practical example of full-stack web development and highlights the importance of thoughtful design over feature overload.