# QLP - Key Architectural Views

This document provides detailed diagrams for specific, critical sub-systems and workflows within the QLP platform.

## 1. AITL Validation & Self-Critique Loop

This diagram illustrates the process by which the system validates and iteratively improves its own generated code.

```mermaid
graph TD;
    subgraph "Validation Service"
        K1(Kafka: Artifacts) --> C{Consumer};
        C -- "Receives all artifacts for an intent" --> A[1. Assemble Full Project];
        A --> J(2. Create K8s Job for Sandbox);
        
        subgraph "Ephemeral Sandbox Pod"
            J -- "Starts Pod with Project Code" --> V[Validator];
            V -- "a. Build & Unit Tests" --> R1{Result 1};
            V -- "b. Static & Security Analysis" --> R2{Result 2};
        end

        R1 & R2 --> E("3. Evaluate Scores");
        
        subgraph "Decision"
            E -- "Score >= 95%" --> P(4a. Publish VALIDATION_PASSED);
            E -- "Score < 95%" --> F(4b. Generate Refinement Prompt);
        end

        F -- "Instructs LLM to fix code" --> A;
        P --> K2(Kafka: Validated Artifacts);
    end
```

### Flow Description

1.  **Assembly:** The Validation Service consumes all `CODE_GENERATED`, `TESTS_GENERATED`, etc. events for a specific job from the `Artifacts` topic. It assembles these files into a complete, runnable project structure.
2.  **Sandboxing:** It creates a dedicated Kubernetes `Job` to run the validation in a secure, ephemeral pod. The pod contains the full assembled project.
3.  **Holistic Testing:** Inside the sandbox, the project is built, unit tests are run, and a battery of static analysis and security scans are performed.
4.  **Decision & Refinement:**
    - If all scores meet the high threshold (e.g., 95%), a `VALIDATION_PASSED` event is published.
    - If the scores are too low, the service generates a "refinement prompt" detailing all the failures. It then re-enters the loop at Step 1, using the refinement prompt to have an LLM fix the code. This loop continues until the score passes or a maximum number of attempts is reached. 