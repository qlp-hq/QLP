# QLP - Critical Architectural Decisions

This document records the key architectural decisions made for the QLP platform and the rationale behind them.

## 1. Microservices over Monolith

- **Decision:** The platform is designed as a collection of small, independent microservices that communicate over a network, rather than a single, large monolithic application.
- **Rationale:**
    - **Scalability:** Allows for independent scaling of services. The `Agent Worker` service can be scaled to hundreds of replicas without affecting the `Intent Service`.
    - **Resilience:** A crash in one service (e.g., the `Documentation Service`) does not bring down the entire platform.
    - **Maintainability:** Smaller codebases are easier to understand, maintain, and update. Different teams can work on different services in parallel.
    - **Technological Flexibility:** Allows for using the best technology for a specific job, though we are standardizing on Go for now.

## 2. Asynchronous Communication with Kafka over Direct API Calls

- **Decision:** Services communicate asynchronously using a Kafka message broker instead of making direct synchronous requests (e.g., REST, gRPC) to each other.
- **Rationale:**
    - **Decoupling:** Services are completely decoupled. A producer service does not need to know which service will consume its event, or even if the consumer is currently online.
    - **Durability & Resilience:** Kafka acts as a durable log. If a consumer service crashes, the messages are safely stored in the topic and can be reprocessed when the service restarts. This prevents data loss.
    - **Load Balancing:** Kafka automatically handles the distribution of messages across multiple instances of a consumer service, providing load balancing out of the box.

## 3. Ephemeral Kubernetes Jobs for Agent Execution over Long-Running Workers

- **Decision:** Agent tasks are executed in single-use, dedicated Kubernetes pods created by a `Job`, which are destroyed immediately after task completion.
- **Rationale:**
    - **Security:** This is the most critical driver. It provides a perfect sandbox. Any malicious code an agent might generate is contained within a temporary environment with no host access and is destroyed moments later.
    - **Resource Efficiency:** We only consume compute resources when a task is actively being executed. We don't have idle worker pods sitting around waiting for work.
    - **Clean State:** Every task runs in a fresh, clean environment, eliminating the risk of state from a previous task interfering with the current one.

## 4. Centralized Authentication at the API Gateway

- **Decision:** All authentication, specifically the validation of Azure Entra ID JWTs, is handled exclusively at the API Gateway. Internal services do not handle authentication.
- **Rationale:**
    - **Single Responsibility:** Centralizes security concerns at the edge of the network.
    - **Simplified Services:** Internal services can be much simpler as they can operate on the assumption that any request they receive has already been authenticated.
    - **Flexibility:** If we ever need to change our authentication method, we only need to update the API Gateway, not every single microservice.

## 5. Self-Optimizing Agent Registry over Static Prompts

- **Decision:** Agent prompts are dynamically generated and the successful ones are stored in a PostgreSQL database (Agent Registry) for future reuse.
- **Rationale:**
    - **Learning System:** Allows the platform to learn and improve over time. The system gets faster and more accurate as it processes more requests.
    - **Agility:** Prompts can be updated and new agent capabilities can be added without needing to recompile or redeploy services.
    - **Performance:** Reusing a known, high-quality prompt from the database is significantly faster and cheaper than generating a new one with an LLM call every time. 