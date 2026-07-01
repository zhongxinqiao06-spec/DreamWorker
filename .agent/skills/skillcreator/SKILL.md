---
name: Skill Creator
description: Creates or improves DreamWorker/Agent Skills in standard SKILL.md format with clear triggers, instructions, allowed tools, tests, and installation layout.
when_to_use: Use when the user asks to create a skill, install a skill, improve a skill, convert repeated instructions into a reusable skill, or design skill packaging.
allowed-tools: filesystem_project_read, artifact_write, human_question
category: general
version: 0.1.0
output-artifacts: SKILL.md
dreamworker-built-in: true
---

## Instructions

Create production-quality skills under `.agent/skills/<skill-name>/SKILL.md`.

For each skill:

- Use YAML frontmatter with `name`, `description`, `when_to_use`, `allowed-tools`, `category`, `version`, `output-artifacts`, and `dreamworker-built-in`.
- Keep the description specific enough for automatic invocation.
- Put durable operating instructions in markdown.
- Include expected inputs, outputs, quality checks, and failure modes when useful.
- Prefer concise instructions over long background prose.
- Keep secrets, raw provider responses, and environment values out of skill content.

When improving an existing skill, preserve its command name unless the user explicitly asks to rename it.
