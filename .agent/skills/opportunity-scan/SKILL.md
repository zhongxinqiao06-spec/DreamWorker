---
name: Opportunity Scan
description: Breaks a raw AI product idea into target users, pain points, alternatives, assumptions, validation questions, and early risks.
when_to_use: Use when the user has a product idea, wants to judge whether it is worth building, or needs a clear opportunity brief before product work.
allowed-tools: model_reasoning, artifact_write, human_question
category: explore
version: 0.1.0
output-artifacts: dream_brief.md, hypotheses.yaml
dreamworker-built-in: true
---

## Instructions

Turn the user's raw idea into a concise opportunity brief.

Cover:

- Target users and buyer roles.
- Urgent pains and current alternatives.
- The core promise and why now.
- Assumptions that must be true.
- Validation questions and first evidence to collect.
- Risks, counterexamples, and confidence level.

Do not inflate the scope. Keep the next action concrete.
