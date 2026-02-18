# Pulley — Inferred Technology Stack & Technical Profile

> This document combines publicly available information with inferences drawn from
> job postings, technology detection services, and domain analysis. Items marked
> **(confirmed)** come from direct public evidence; items marked **(inferred)** are
> educated deductions based on the product domain, hiring signals, and detected tooling.

---

## Architecture Overview (Inferred)

Pulley is a B2B SaaS platform operating in the fintech/legaltech space. Based on job
postings and detected technologies, it appears to run a **monolithic-to-services hybrid
architecture** with a Go backend, React frontend, PostgreSQL persistence, and an emerging
AI/agentic layer. The platform handles sensitive financial and legal data (equity ownership,
409A valuations, tax filings), which drives strong requirements around data integrity,
audit trails, compliance, and security.

---

## Backend

| Technology | Confidence | Source |
|-----------|------------|--------|
| **Go (Golang)** | Confirmed | Senior Backend Engineer job posting requires "strong proficiency in Go"; AI Fullstack role lists "backend services in Go" |
| **GraphQL** | Confirmed | Detected on himalayas.app tech stack; Hasura (GraphQL engine) also detected |
| **Hasura** | Confirmed | Detected via technology scanning (himalayas.app) — provides instant GraphQL APIs over PostgreSQL |
| **REST APIs** | Inferred | Job postings reference "APIs and service interfaces"; Nasdaq Private Market integration described as API-based |
| **Python** | Confirmed | Detected in tech stack scans (himalayas.app); likely used for data science, valuation models, or internal tooling |
| **Temporal** (or similar) | Likely | AI Fullstack job lists "Experience with Temporal or similar workflow orchestration frameworks" as nice-to-have |

### Backend Observations
- Go as the primary backend language is a strong signal for a performance-oriented,
  type-safe approach — common in fintech where correctness and concurrency matter.
- Hasura over PostgreSQL suggests they may use an auto-generated GraphQL layer for
  rapid API development, with custom Go services handling complex business logic
  (valuations, vesting calculations, compliance rules).
- The mention of "agentic workflows and tool orchestration" in the AI role points
  to an emerging event-driven or workflow-based architecture for AI features.

---

## Frontend

| Technology | Confidence | Source |
|-----------|------------|--------|
| **React** | Confirmed | Frontend Engineer job posting; tech stack scans |
| **TypeScript** | Confirmed | Frontend and Fullstack job postings require "JavaScript/TypeScript" and "TypeScript/React" |
| **Redux** | Confirmed | Detected via tech scanning (himalayas.app) |
| **Tailwind CSS** | Likely | Referenced in some external job aggregator descriptions |
| **Webpack** | Confirmed | Detected via tech scanning |
| **Tachyons** | Confirmed | Detected (CSS toolkit — possibly legacy or used alongside Tailwind) |
| **Astro** | Confirmed | Website source references "pulley astro console msg" — marketing site uses Astro framework |

### Frontend Observations
- React + TypeScript + Redux is a mature, well-understood frontend stack.
- The frontend team is product-focused ("influence product direction, own the end-to-end experience").
- The presence of both Tachyons and what appears to be Tailwind suggests a possible
  migration from one utility CSS approach to another over time.
- Astro for the marketing/public site (pulley.com) while React powers the authenticated
  application is a common pattern for performance optimization.

---

## Database & Data Layer

| Technology | Confidence | Source |
|-----------|------------|--------|
| **PostgreSQL** | Confirmed | AI Fullstack job posting explicitly lists "Experience with PostgreSQL"; detected in tech scans |
| **Redis** | Confirmed | AI Fullstack job posting lists "Experience with PostgreSQL, Redis, and system observability" |
| **Hasura** | Confirmed | GraphQL engine sitting atop PostgreSQL for auto-generated APIs |

### Data Layer Observations
- PostgreSQL is the natural choice for a fintech product requiring ACID compliance,
  complex relational queries (cap table ownership chains, vesting schedules, waterfall
  calculations), and audit trails.
- Redis likely serves as a caching layer and possibly for session management, job
  queues, or rate limiting.
- Job postings emphasize "database schema design, migrations, and data integrity" —
  strong signals that the data model is complex and carefully managed.

---

## Infrastructure & DevOps

| Technology | Confidence | Source |
|-----------|------------|--------|
| **Google Cloud Platform (GCP)** | Confirmed | Detected via tech scanning (himalayas.app) |
| **NGINX** | Confirmed | Detected via tech scanning |
| **OpenResty** | Confirmed | Detected (NGINX + Lua extension — likely for edge logic/rate limiting) |
| **Varnish** | Confirmed | Detected (HTTP caching/accelerator) |
| **Cloudflare** | Confirmed | Detected (CDN, DDoS protection, DNS) |
| **Amazon CloudFront** | Confirmed | Detected (CDN for static assets) |
| **Fastly** | Confirmed | Detected (CDN — possibly for specific asset delivery or AB testing edge logic) |
| **Docker** | Inferred | Standard for GCP deployments; Go services are commonly containerized |
| **Kubernetes / GKE** | Inferred | Likely given GCP hosting and team size; standard for scaling Go microservices |
| **Terraform / IaC** | Inferred | Common for GCP-based fintech companies at this scale |

### Infrastructure Observations
- Multi-CDN setup (Cloudflare + CloudFront + Fastly) suggests either different CDNs
  for different purposes (marketing site vs. app assets vs. API edge caching) or
  a migration in progress.
- Varnish + OpenResty + NGINX is a sophisticated caching and reverse proxy stack,
  likely used for high-performance API serving and static asset delivery.
- GCP as the primary cloud is consistent with the detected Hasura usage (Hasura Cloud
  runs natively on GCP).

---

## AI & Machine Learning

| Technology | Confidence | Source |
|-----------|------------|--------|
| **LLM Integration** | Confirmed | AI Fullstack role requires "Experience integrating LLMs or AI capabilities into production applications" |
| **Agentic Frameworks** | Confirmed | Role requires "Familiarity with agentic frameworks and tool orchestration patterns" |
| **Temporal** (workflow orchestration) | Likely | Listed as nice-to-have for the AI role |
| **AI-native product direction** | Confirmed | Job postings state "redefine how early-stage leaders build and manage their companies in an AI-native world" |

### AI Observations
- Pulley is actively building AI-powered equity workflows — this is a current,
  high-priority investment area (all three engineering roles posted in early 2026
  are new).
- The "agentic workflow" language suggests they're building AI agents that can
  perform multi-step equity operations (e.g., modeling scenarios, generating
  compliance reports, answering founder questions about their cap table).
- Stipends for "AI tools" (Cursor, Copilot mentioned in job postings) indicate
  AI-assisted development is part of the engineering culture.

---

## Analytics & Observability

| Technology | Confidence | Source |
|-----------|------------|--------|
| **Segment** | Confirmed | Detected via tech scanning |
| **Amplitude** | Confirmed | Detected via tech scanning |
| **Google Analytics** | Confirmed | Detected via tech scanning |
| **Google Tag Manager** | Confirmed | Detected via tech scanning |
| **FullStory** | Confirmed | Detected (session replay and digital experience analytics) |
| **Google Optimize** | Confirmed | Detected (A/B testing — note: Google Optimize was sunset in 2023, may be residual) |
| **Facebook Pixel** | Confirmed | Detected (marketing attribution) |
| **System observability** | Confirmed | AI Fullstack job requires "system observability in production" |

### Observability Inferences
- **Datadog or similar APM** (inferred) — The explicit requirement for "system
  observability in production" alongside Go + PostgreSQL + Redis strongly suggests
  a dedicated observability platform. Datadog, Grafana, or New Relic are most likely
  given the GCP ecosystem.
- Segment as the customer data platform (CDP) feeding into Amplitude for product
  analytics and Google Analytics for marketing analytics is a modern, well-architected
  analytics pipeline.

---

## Marketing Site & CMS

| Technology | Confidence | Source |
|-----------|------------|--------|
| **Astro** | Confirmed | Console log in page source references Astro |
| **Webflow** | Confirmed | Detected via tech scanning (likely used previously or for specific landing pages) |
| **Google Fonts** | Confirmed | Detected |

---

## Business & Operations Tools

| Technology | Confidence | Source |
|-----------|------------|--------|
| **Greenhouse** | Confirmed | Job board hosted on job-boards.greenhouse.io/pulley |
| **Ashby** | Confirmed | Previously used Ashby for ATS (detected in tech scans; old career links point to jobs.ashbyhq.com/Pulley) |
| **Salesforce** | Confirmed | Detected via tech scanning (CRM) |
| **Intercom** | Confirmed | Detected (customer support/messaging) |
| **Google Workspace** | Confirmed | Detected (internal collaboration) |
| **Mailgun** | Confirmed | Detected (transactional email delivery) |
| **Notion** | Confirmed | GDPR/cookie policy and fraud protection pages hosted on Notion |

### Advertising Platforms (Confirmed)
- Google Ads
- Facebook Ads
- Twitter/X Ads
- LinkedIn Ads

---

## Security & Compliance

| Aspect | Details |
|--------|---------|
| **SOC 2 Type 2** | Certified; audited by Insight Assurance |
| **GDPR** | Compliant with published cookie policy |
| **Data Handling** | Sensitive financial data (cap tables, ownership, tax info, PII) |
| **E-Signing** | Built-in document signing (likely DocuSign API or custom implementation) |
| **Non-custodial Crypto** | Token distribution is non-custodial by design |
| **Transfer Restrictions** | API-driven enforcement with Nasdaq Private Market |

### Security Inferences
- SOC 2 Type 2 certification at a 68-person company indicates mature security
  practices: encrypted data at rest and in transit, access controls, audit logging,
  incident response procedures, and regular penetration testing.
- Operating in fintech with equity data likely means they implement role-based
  access control (RBAC), row-level security in PostgreSQL, and comprehensive audit
  trails for every data mutation.
- The mention of "high-trust domains" in job postings reinforces that security and
  data integrity are first-class concerns.

---

## Integration Ecosystem

| Integration | Type |
|------------|------|
| **HRIS Systems** | Payroll and employee data sync |
| **Accounting Systems** | ASC 718 and financial reporting |
| **Nasdaq Private Market** | Liquidity and tender offers via Cap Table Connect API |
| **Crypto Exchanges** | Automated token distribution |
| **Custodians & Multi-sigs** | Token custody integration |
| **Payroll Providers** | Token compensation tax withholding sync |

---

## Engineering Culture & Practices

**From job postings and public statements:**

- **End-to-end ownership** — Engineers own features from concept through production
- **Product-minded engineering** — Engineers influence product direction, not just implement specs
- **AI-native development** — Stipends for AI tools; Cursor and Copilot explicitly mentioned
- **Remote-first** — All engineering roles listed as Remote (US/Canada)
- **Code quality emphasis** — "Clean, testable code" and code reviews highlighted
- **Small, high-leverage teams** — 68 total employees; engineering is a significant portion
- **Ship & iterate** — Fast shipping with user feedback loops

**Compensation Ranges (2026 postings):**

| Role | Salary Range |
|------|-------------|
| Frontend Engineer II | $150,000–$180,000 |
| Senior Backend Engineer | $180,000–$200,000 |
| Senior Fullstack Engineer, AI | $180,000–$200,000 |
| Senior Staff Software Engineer | $237,000–$284,000 |

---

## Technology Stack Summary

```
┌─────────────────────────────────────────────────────────┐
│                    CDN / Edge Layer                      │
│         Cloudflare · CloudFront · Fastly                │
├─────────────────────────────────────────────────────────┤
│                  Reverse Proxy / Cache                   │
│            NGINX · OpenResty · Varnish                  │
├────────────────────┬────────────────────────────────────┤
│   Marketing Site   │         Application Frontend       │
│       Astro        │     React · TypeScript · Redux     │
│     (Webflow?)     │       Webpack · Tailwind CSS       │
├────────────────────┴────────────────────────────────────┤
│                      API Layer                           │
│          Hasura (GraphQL) · Go (REST/gRPC)              │
├─────────────────────────────────────────────────────────┤
│                   Backend Services                       │
│   Go (core business logic, APIs, agentic workflows)     │
│   Python (valuations, data science, tooling)            │
├─────────────────────────────────────────────────────────┤
│                 AI / Workflow Layer                       │
│     LLM Integration · Agentic Frameworks · Temporal     │
├─────────────────────────────────────────────────────────┤
│                    Data Layer                             │
│           PostgreSQL (primary) · Redis (cache)          │
├─────────────────────────────────────────────────────────┤
│                   Infrastructure                         │
│    Google Cloud Platform · Docker · Kubernetes (GKE)    │
├─────────────────────────────────────────────────────────┤
│               Observability & Analytics                  │
│  Segment · Amplitude · FullStory · Google Analytics     │
│          System APM (Datadog/Grafana inferred)          │
├─────────────────────────────────────────────────────────┤
│                 Business Operations                      │
│  Greenhouse · Salesforce · Intercom · Mailgun · Notion  │
└─────────────────────────────────────────────────────────┘
```

---

## Key Technical Differentiators (Inferred)

1. **Hasura + Go hybrid** — Auto-generated GraphQL for standard CRUD with Go services
   for complex domain logic (waterfall calculations, vesting engines, 409A models).
   This gives rapid development speed without sacrificing control over critical paths.

2. **AI-native ambition** — Pulley is investing heavily in AI capabilities in early 2026,
   building agentic workflows that can automate multi-step equity operations. This is
   a clear competitive bet against Carta.

3. **API-first liquidity** — The Nasdaq Private Market integration via Cap Table Connect
   API is a technical moat — Pulley was the first cap table platform to offer this
   programmatic liquidity integration.

4. **Multi-chain token support** — Token cap tables work with "any chain," suggesting
   a chain-agnostic integration layer rather than being locked to specific blockchains.

5. **Lean engineering leverage** — With ~68 total employees and 4,000+ customers,
   Pulley operates at high engineering leverage, suggesting strong automation,
   infrastructure-as-code, and product-led growth mechanics.

---

*Sources: Pulley job postings (Greenhouse, February 2026), himalayas.app tech stack detection,
YC company profile, pulley.com source code inspection, Nasdaq Private Market press releases,
TechCrunch, Business Insider. Last updated: February 2026.*
