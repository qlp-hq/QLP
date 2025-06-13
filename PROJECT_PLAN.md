# QLP - Project Implementation Plan & Tracker

This document outlines the high-level plan and serves as a tracker for building the QLP platform. Our primary deployment target is **Microsoft Azure**.

---

## Phase 1: Foundational Setup & Service Decomposition (Completed)

The goal of this phase was to refactor the existing codebase into a containerized, microservices-based architecture and set up the core infrastructure.

-   [x] **Task 1.1: Set up Azure Resources** (Simulated via `docker-compose.yml`)
-   [x] **Task 1.2: Refactor Codebase into Microservices**
-   [x] **Task 1.3: Containerize Services**
-   [x] **Task 1.4: Implement Core Kafka Communication**
-   [x] **Task 1.5: Basic Kubernetes Deployment** (Simulated via `docker-compose.yml`)

---

## Phase 2: Agent Execution & Validation Loop (In Progress)

The goal of this phase is to implement the dynamic agent execution model and the core feedback loop.

-   [x] **Task 2.1: Implement Agent Worker Service**: Create a service that listens to the `task.ready` Kafka topic.
-   [x] **Task 2.2: Implement Agent Logic**: The worker contains logic for different specialist agents (e.g., `CodeGenAgent`) and publishes its results (`artifact.created` event).
-   [x] **Task 2.3: Implement Basic Validation Service**: A service consumes `artifact.created` events and will perform validation checks.
-   [x] **Task 2.4: Connect Orchestrator to Validation**: The `Orchestrator` consumes `artifact.validated` events to mark tasks as complete in the DAG.
-   [ ] **Task 2.5: Implement Dynamic Agents & Prompt Registry**: Implement more complex agents (`DynamicAgent`) and a PostgreSQL-backed registry for storing, retrieving, and improving agent prompts over time.

---

## Phase 3: Advanced Orchestration & Deployment (In Progress)

The goal of this phase is to build out the more sophisticated orchestration features and the deployment pipeline.

-   [x] **Task 3.1: Implement Ensemble & Judgement Logic**: The `Orchestrator` supports "ensemble" tasks that run a prompt against multiple models and a "judgement" task to pick the best result.
-   [ ] **Task 3.2: Implement Self-Correction/Refinement Loop**: The `Validation Service` should trigger a "refinement" loop if validation fails, sending the task back to the orchestrator with feedback.
-   [ ] **Task 3.3: Implement the Packaging Service**: Consumes `validation.passed` events and assembles the final `QLCapsule` artifact.
-   [ ] **Task 3.4: Implement the Deployment Service**: Contains logic to build the final application Docker image from the capsule and deploy it.

---

## Phase 4: User-Facing Services & Production Readiness (Not Started)

The goal of this phase is to build the UI and add the necessary operational components.

-   [ ] **Task 4.1: Implement Authentication**: Integrate Azure Entra ID with the API Gateway.
-   [ ] **Task 4.2: Implement the Dashboard Service**: Consumes events from all topics to provide real-time updates and a REST/WebSocket API for the frontend.
-   [ ] **Task 4.3: Implement Observability**: Add OpenTelemetry for distributed tracing, expose Prometheus metrics, and set up Grafana dashboards.
-   [ ] **Task 4.4: Build the Frontend UI**: Create the user dashboard to display the real-time progress and final results.