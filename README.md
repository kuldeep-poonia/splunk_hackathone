# Autonomous SRE: Chaos Simulation, Observability & AI Incident Analysis

## Overview

Autonomous SRE is a personal project built for the Splunk Hackathon.

The goal was simple: simulate real-world distributed system failures, visualize system behavior through Splunk dashboards, and explore whether AI can help engineers understand incidents faster.

This project combines chaos engineering simulations, telemetry analysis, Splunk Enterprise dashboards, and AI-generated incident summaries into a single workflow.

Rather than focusing on production deployment, the focus of this project is experimentation, learning, and demonstrating how observability and AI can work together during failure scenarios.

---

## Why This Project?

Modern systems fail in unexpected ways.

Traffic spikes, node failures, retry storms, network partitions, and cascading outages can quickly make debugging difficult.

The idea behind this project was to answer three questions:

* What happens when multiple failures occur at the same time?
* Can telemetry be visualized clearly through Splunk?
* Can AI help summarize incidents and provide recovery suggestions?

---

## What It Does

### Chaos Simulation

The simulation engine generates realistic failure scenarios such as:

* Flash Crowd Events
* Node Failures
* Network Partitions
* Retry Storms
* Cascading Failures

These simulations generate telemetry data that represents how a distributed system behaves under stress.

### Splunk Observability

Generated telemetry is ingested into Splunk Enterprise.

Dashboards visualize:

* Risk Score
* Queue Depth
* Latency
* Scaling Behavior
* Failure Events

This provides a clear timeline of how incidents evolve.

### AI Incident Analysis

A lightweight analyzer fetches telemetry from Splunk using the REST API and forwards relevant incident information to an AI model through GitHub Models.

The AI produces:

* Severity Assessment
* Root Cause Summary
* Business Impact
* Suggested Recovery Actions

---

## System Architecture

```text
Chaos Simulation Engine
          │
          ▼
simulation_telemetry.jsonl
          │
          ▼
    Splunk Enterprise
          │
          ├── Dashboards
          ├── Search
          └── Analytics
          │
          ▼
   AI Incident Analyzer
          │
          ▼
 Incident Summary Report
```

---

## Project Structure

```text
.
├── main.go
├── splunk_ai_analyzer.go
├── simulation_telemetry.jsonl
├── monte_carlo_results.json
├── README.md
│
└── control/
    ├── chaos_simulation_test.go
    ├── adversarial_physics_test.go
    ├── ekf_robustness_test.go
    ├── policy_controller.go
    ├── coordinated_optimizer.go
    ├── kalman.go
    ├── actuator_dynamics.go
    ├── state_transition.go
    └── ...
```


```

End-to-End Data Flow
Chaos Simulation
        │
        ▼
simulation_telemetry.jsonl
        │
        ▼
Splunk Enterprise
(Data Ingestion + Dashboards)
        │
        ▼
Splunk REST API
        │
        ▼
AI Incident Analyzer
        │
        ▼
GitHub Models API
(DeepSeek-V3-0324)
        │
        ▼
Incident Summary


```



---

## Running The Project

### Generate Telemetry

```bash
go test -v -run TestChaos_CascadingDeathSpiral ./control
```

### Visualize in Splunk

Configure Splunk to monitor:

```text
simulation_telemetry.jsonl
```

Create dashboards using the ingested telemetry.

### Generate AI Incident Summary

```bash
go run .
```

---

## Technology Stack

* Go
* Splunk Enterprise
* GitHub Models
* DeepSeek-V3-0324
* JSONL Telemetry
* Chaos Engineering Concepts
* Control Theory Concepts





## Personal Note

This project was built independently with limited time and resources during the hackathon.

I tried to implement as much functionality as possible within those system constraints and focused on building a complete end-to-end workflow rather than a perfect production system.

There is still a lot that can be improved, and I plan to continue developing this project after the hackathon by improving simulations, telemetry pipelines, dashboards, and AI analysis capabilities.

For now, this repository represents my best attempt at exploring how chaos engineering, observability, and AI can work together in a practical workflow.

Thank you for taking the time to review it.



Open Source License

This project is released under the MIT License.

The repository is intended for educational, research, and experimentation purposes.