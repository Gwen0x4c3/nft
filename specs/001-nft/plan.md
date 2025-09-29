# Implementation Plan: 简单 NFT 铸造与拍卖平台

**Branch**: `001-nft` | **Date**: 2025-09-29 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-nft/spec.md`

## Execution Flow (/plan command scope)

```
1. Load feature spec from Input path
   → If not found: ERROR "No feature spec at {path}"
2. Fill Technical Context (scan for NEEDS CLARIFICATION)
   → Detect Project Type from file system structure or context (web=frontend+backend, mobile=app+api)
   → Set Structure Decision based on project type
3. Fill the Constitution Check section based on the content of the constitution document.
4. Evaluate Constitution Check section below
   → If violations exist: Document in Complexity Tracking
   → If no justification possible: ERROR "Simplify approach first"
   → Update Progress Tracking: Initial Constitution Check
5. Execute Phase 0 → research.md
   → If NEEDS CLARIFICATION remain: ERROR "Resolve unknowns"
6. Execute Phase 1 → contracts, data-model.md, quickstart.md, agent-specific template file (e.g., `CLAUDE.md` for Claude Code, `.github/copilot-instructions.md` for GitHub Copilot, `GEMINI.md` for Gemini CLI, `QWEN.md` for Qwen Code or `AGENTS.md` for opencode).
7. Re-evaluate Constitution Check section
   → If new violations: Refactor design, return to Phase 1
   → Update Progress Tracking: Post-Design Constitution Check
8. Plan Phase 2 → Describe task generation approach (DO NOT create tasks.md)
9. STOP - Ready for /tasks command
```

**IMPORTANT**: The /plan command STOPS at step 7. Phases 2-4 are executed by other commands:

- Phase 2: /tasks command creates tasks.md
- Phase 3-4: Implementation execution (manual or via tools)

## Summary

NFT minting and auction platform enabling digital artists to create, sell, and trade NFTs through competitive bidding. Core features include user authentication, blockchain-based NFT creation, real-time auction system with bidding, and comprehensive notification system. Built with microservices architecture using Golang backend, React frontend, PostgreSQL database, Redis caching, and Kafka message queues to support 1000+ concurrent users with <200ms API response times.

## Technical Context

**Language/Version**: Golang 1.21+, React.js 18+, Solidity 0.8+  
**Primary Dependencies**: Gin (web framework), go-ethereum (blockchain), GORM (ORM), Redis, Kafka, PostgreSQL  
**Storage**: PostgreSQL (primary database), Redis (caching), Ethereum Sepolia (blockchain), Local file system (images)  
**Testing**: Go testing package, JMeter (load testing), E2E testing (frontend-backend integration)  
**Target Platform**: Linux server (backend microservices), Web browsers (React frontend)
**Project Type**: web - determines source structure (frontend + backend)  
**Performance Goals**: API responses <200ms p95, >500 req/s throughput, 1000+ concurrent users  
**Constraints**: <200ms p95 API response, Ethereum gas costs, blockchain confirmation times  
**Scale/Scope**: 1000+ concurrent users, 5 microservices, Real-time WebSocket notifications

## Constitution Check

_GATE: Must pass before Phase 0 research. Re-check after Phase 1 design._

### Code Quality Excellence

- [x] TDD approach planned with tests written before implementation (Go testing, contract tests first)
- [x] Code review process defined with clear approval criteria (GitHub PR reviews required)
- [x] Static analysis tools integrated into development workflow (golangci-lint, ESLint)
- [x] Documentation requirements identified for new APIs/features (OpenAPI specs, README updates)

### User Experience First

- [x] User experience impact assessed and optimized (React UI with Ant Design, real-time updates)
- [x] Performance requirements defined (UI <100ms, API <200ms p95)
- [x] Error handling and user feedback mechanisms planned (Toast notifications, WebSocket status)
- [x] Accessibility considerations documented (ARIA labels, keyboard navigation)

### Performance & Scalability

- [x] Performance targets defined and measurable (API <200ms p95, >500 req/s, 1000+ users)
- [x] Caching strategy planned for frequently accessed data (Redis for NFT metadata, auction data)
- [x] Horizontal scaling approach documented (microservices with load balancers)
- [x] Database query optimization considered (PostgreSQL indexes, connection pooling)

### Microservices Integration

- [x] Service communication patterns defined (gRPC/HTTP REST, Kafka for async)
- [x] API versioning strategy documented (/api/v1 prefix, backward compatibility)
- [x] Circuit breaker and timeout handling planned (service resilience patterns)
- [x] Message queue (Kafka) usage identified where needed (NFT minting, notifications)
- [x] Caching layer (Redis) integration planned (session management, hot data)

### Observability & Monitoring

- [x] Structured logging approach defined (JSON format with zap logger)
- [x] Metrics collection strategy planned (Prometheus metrics, business KPIs)
- [x] Health check endpoints identified (/health for each microservice)
- [x] Alert conditions and thresholds defined (response time >200ms, error rate >1%)

## Project Structure

### Documentation (this feature)

```
specs/[###-feature]/
├── plan.md              # This file (/plan command output)
├── research.md          # Phase 0 output (/plan command)
├── data-model.md        # Phase 1 output (/plan command)
├── quickstart.md        # Phase 1 output (/plan command)
├── contracts/           # Phase 1 output (/plan command)
└── tasks.md             # Phase 2 output (/tasks command - NOT created by /plan)
```

### Source Code (repository root)

<!--
  ACTION REQUIRED: Replace the placeholder tree below with the concrete layout
  for this feature. Delete unused options and expand the chosen structure with
  real paths (e.g., apps/admin, packages/something). The delivered plan must
  not include Option labels.
-->

```
backend/
├── cmd/                     # Main applications
│   ├── user-service/
│   ├── nft-service/
│   ├── auction-service/
│   ├── notification-service/
│   └── query-service/
├── internal/                # Private application and library code
│   ├── models/             # Data models and entities
│   ├── services/           # Business logic layer
│   ├── repositories/       # Data access layer
│   ├── handlers/           # HTTP handlers
│   ├── middleware/         # HTTP middleware
│   └── config/             # Configuration management
├── pkg/                    # Public library code
│   ├── auth/              # JWT authentication
│   ├── blockchain/        # Ethereum integration
│   ├── cache/             # Redis cache utilities
│   └── queue/             # Kafka message queue
├── contracts/             # Solidity smart contracts
│   ├── NFT.sol
│   └── Auction.sol
├── deployments/           # Docker and deployment configs
└── tests/
    ├── contract/          # Contract tests
    ├── integration/       # Service integration tests
    └── unit/              # Unit tests

frontend/
├── src/
│   ├── components/        # Reusable UI components
│   ├── pages/            # Route-based page components
│   ├── services/         # API client services
│   ├── hooks/            # Custom React hooks
│   ├── utils/            # Utility functions
│   └── types/            # TypeScript type definitions
├── public/               # Static assets
└── tests/
    ├── components/       # Component tests
    └── integration/      # E2E tests

docker-compose.yml        # Local development services
uploads/                  # Local file storage (gitignored)
```

**Structure Decision**: Web application structure selected based on microservices architecture with separate backend services and React frontend. Backend follows Go project layout standards with cmd/, internal/, and pkg/ directories. Frontend uses standard React project structure with component-based organization.

## Phase 0: Outline & Research

1. **Extract unknowns from Technical Context** above:

   - For each NEEDS CLARIFICATION → research task
   - For each dependency → best practices task
   - For each integration → patterns task

2. **Generate and dispatch research agents**:

   ```
   For each unknown in Technical Context:
     Task: "Research {unknown} for {feature context}"
   For each technology choice:
     Task: "Find best practices for {tech} in {domain}"
   ```

3. **Consolidate findings** in `research.md` using format:
   - Decision: [what was chosen]
   - Rationale: [why chosen]
   - Alternatives considered: [what else evaluated]

**Output**: research.md with all NEEDS CLARIFICATION resolved

## Phase 1: Design & Contracts

_Prerequisites: research.md complete_

1. **Extract entities from feature spec** → `data-model.md`:

   - Entity name, fields, relationships
   - Validation rules from requirements
   - State transitions if applicable

2. **Generate API contracts** from functional requirements:

   - For each user action → endpoint
   - Use standard REST/GraphQL patterns
   - Output OpenAPI/GraphQL schema to `/contracts/`

3. **Generate contract tests** from contracts:

   - One test file per endpoint
   - Assert request/response schemas
   - Tests must fail (no implementation yet)

4. **Extract test scenarios** from user stories:

   - Each story → integration test scenario
   - Quickstart test = story validation steps

5. **Update agent file incrementally** (O(1) operation):
   - Run `.specify/scripts/bash/update-agent-context.sh cursor`
     **IMPORTANT**: Execute it exactly as specified above. Do not add or remove any arguments.
   - If exists: Add only NEW tech from current plan
   - Preserve manual additions between markers
   - Update recent changes (keep last 3)
   - Keep under 150 lines for token efficiency
   - Output to repository root

**Output**: data-model.md, /contracts/\*, failing tests, quickstart.md, agent-specific file

## Phase 2: Task Planning Approach

_This section describes what the /tasks command will do - DO NOT execute during /plan_

**Task Generation Strategy**:

- Load `.specify/templates/tasks-template.md` as base
- Generate tasks from Phase 1 design docs (contracts, data model, quickstart)
- Each contract → contract test task [P]
- Each entity → model creation task [P]
- Each user story → integration test task
- Implementation tasks to make tests pass

**Ordering Strategy**:

- TDD order: Tests before implementation
- Dependency order: Models before services before UI
- Mark [P] for parallel execution (independent files)

**Estimated Output**: 25-30 numbered, ordered tasks in tasks.md

**IMPORTANT**: This phase is executed by the /tasks command, NOT by /plan

## Phase 3+: Future Implementation

_These phases are beyond the scope of the /plan command_

**Phase 3**: Task execution (/tasks command creates tasks.md)  
**Phase 4**: Implementation (execute tasks.md following constitutional principles)  
**Phase 5**: Validation (run tests, execute quickstart.md, performance validation)

## Complexity Tracking

_Fill ONLY if Constitution Check has violations that must be justified_

| Violation                  | Why Needed         | Simpler Alternative Rejected Because |
| -------------------------- | ------------------ | ------------------------------------ |
| [e.g., 4th project]        | [current need]     | [why 3 projects insufficient]        |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient]  |

## Progress Tracking

_This checklist is updated during execution flow_

**Phase Status**:

- [ ] Phase 0: Research complete (/plan command)
- [ ] Phase 1: Design complete (/plan command)
- [ ] Phase 2: Task planning complete (/plan command - describe approach only)
- [ ] Phase 3: Tasks generated (/tasks command)
- [ ] Phase 4: Implementation complete
- [ ] Phase 5: Validation passed

**Gate Status**:

- [ ] Initial Constitution Check: PASS
- [ ] Post-Design Constitution Check: PASS
- [ ] All NEEDS CLARIFICATION resolved
- [ ] Complexity deviations documented

---

_Based on Constitution v1.0.0 - See `.specify/memory/constitution.md`_
