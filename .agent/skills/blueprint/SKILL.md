---
name: Architecture Blueprint
description: Turns product scope into system architecture, module boundaries, data contracts, risks, and implementable engineering slices.
when_to_use: Use when the user asks for architecture, development planning, module boundaries, API design, technical risk, or PR breakdown.
allowed-tools: code_reference_read, filesystem_project_read, model_reasoning, artifact_write, human_question
category: development
version: 0.1.0
output-artifacts: blueprint.yaml, dev_plan.md
dreamworker-built-in: true
---

## Instructions

Produce an engineering blueprint with:

- System boundary and runtime flow.
- Core modules and ownership boundaries.
- Data model and API/event contracts.
- Security, privacy, and failure modes.
- Development slices ordered by dependency and risk.
- Tests and acceptance gates.

Match the MVP scope. Do not platformize early.
