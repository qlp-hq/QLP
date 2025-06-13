# QLP - Overall System Architecture

This document outlines the master architectural blueprint for the QuantumLayer Universal Agent Orchestration System (QLP). The system is designed as a cloud-native, multi-tenant, event-driven platform capable of transforming a user's natural language intent into a fully deployed, validated software application.

## Core Principles

- **Event-Driven:** All services are decoupled and communicate asynchronously via Kafka topics.
- **Microservices:** The system is composed of small, specialized, independently deployable services.
- **Ephemeral Execution:** Agent tasks are executed in dedicated, single-use, sandboxed Kubernetes pods for security and isolation.
- **AI-in-the-Loop (AITL):** The system uses a self-critique mechanism to iteratively improve and validate its own generated code.
- **Multi-Tenancy:** The platform is designed from the ground up to support multiple tenants with secure data and process isolation.
- **Observability:** Real-time progress, logs, and "thought bubbles" are streamed to the user to provide transparency and engagement.

## Master Blueprint Diagram

The following diagram illustrates the complete, end-to-end flow of the system, from user authentication to final application deployment.

```mermaid
graph TD;
    A[User] --> UI[Dashboard UI];
    UI -- "Submits Intent and Token" --> GW[API Gateway];

    subgraph "Cloud Platform (AKS)"
        GW --> IS[Intent Service];
        IS -- "1. Invokes RequirementsAgent" --> RA{Reqs Agent};
        RA -- "2. Creates requirements.md" --> IS;
        IS -- "3. Publishes Intent w/ Reqs" --> K1(Kafka);

        K1 --> O[Orchestrator];
        O --> K1;

        K1 -- Events --> DS[Dashboard Service];
        DS -- "Pushes Scores, Docs, Logs" --> UI;

        O -- "Publishes Tasks" --> K2(Kafka);
        K2 --> AW[Agent Workers];
        
        AW -- "Logs 'Thought Bubbles'" --> K_LOG(Kafka);
        K_LOG --> DS;

        AW -- "Publishes Artifacts (QuantumDrops)" --> K3(Kafka);
        K3 --> V["Validation Service (AITL)"];
        V -- "Publishes Reports" --> DS;
        
        V --> P[Packaging Service];
        P --> DEP[Deployment Service];
        DEP -- "Requires HITL?" --> H((Human Approval));
        H --> DEP;
        DEP --> LA((Live App));
    end
    
    style UI fill:#bbf,stroke:#333,stroke-width:2px;
    style DS fill:#bbf,stroke:#333,stroke-width:2px;
end
```

## Component Overview

- **API Gateway:** The single, secure entry point. Validates Azure Entra ID JWT tokens and routes requests.
- **Intent Service:** Receives the user's request and invokes a `RequirementsAgent` to produce a formal requirements document.
- **Orchestrator Service:** Consumes new intents, builds a Directed Acyclic Graph (DAG) of tasks, and publishes ready tasks for execution.
- **Agent Workers (via K8s Jobs):** Ephemeral pods that execute a single task using specialized agents (e.g., CodeGen, TestGen, DiagramGen).
- **Validation Service:** Assembles all generated artifacts and performs holistic validation, including building the code and running tests in a secure sandbox. Employs an AITL loop to fix issues.
- **Dashboard Service:** Consumes events from all over the system to provide a real-time view of the process, including logs, scores, and intermediate "QuantumDrop" deliverables.
- **Packaging Service:** Assembles the validated artifacts into the final `QLCapsule`.
- **Deployment Service:** Takes the final capsule and deploys it to the target environment, potentially waiting for Human-in-the-Loop (HITL) approval.

This architecture provides a robust, scalable, and secure foundation for the QLP platform. 