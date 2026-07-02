# DreamWorker Agent Skills

`.agent/` 是 DreamWorker 项目内 Skill 源目录。Go Engine 启动时会自动扫描 `.agent/skills/<skill-name>/SKILL.md`，把 Skill frontmatter 和 markdown instructions 载入内存，供 Resource Center、Chat Runtime 和后续 Skill 安装流程使用。

## 目录结构

```text
.agent/
  |-- README.md
  `-- skills/
      |-- blueprint/SKILL.md
      |-- competitor-map/SKILL.md
      |-- evaluator/SKILL.md
      |-- launch-plan/SKILL.md
      |-- opportunity-scan/SKILL.md
      |-- prd-draft/SKILL.md
      `-- skillcreator/SKILL.md
```

## Skill 格式

每个 Skill 是一个目录，目录名就是稳定 command/name。目录内必须存在 `SKILL.md`：

```markdown
---
name: skill-name
description: 这个 Skill 什么时候被使用
---

# Skill Title

这里写给 Agent Runtime 的稳定指令、输入、输出、约束和质量标准。
```

兼容 Anthropic/Claude Code 风格：YAML frontmatter + Markdown instructions。Engine 只读取标准文件，不再依赖固定 seed 作为唯一 Skill 来源。

## 当前内置 Skill

- `opportunity-scan`：机会扫描。
- `competitor-map`：竞品地图。
- `prd-draft`：PRD 草案。
- `blueprint`：项目蓝图。
- `launch-plan`：发布计划。
- `evaluator`：评估与质量检查。
- `skillcreator`：生成和安装新 Skill。

## 运行时规则

- Skill 可以提供 instruction、allowed tools、artifact contract 和 runtime policy。
- Skill 可以被 Agent runtimeConfig、Project binding 或 Chat Context 引用。
- Skill 不能包含 raw secret、masked secret、MCP env value、provider raw response 或真实用户私密数据。
- Skill 生成和安装默认写入根目录 `.agent/skills/<skill-name>/SKILL.md`。
- Windows package 会把 `.agent` 作为 `extraResources` 一起打包。
- Skill 文案面向中文工作台时使用中文；协议名、字段名、模型名和命令名可以保留原文。

## 新增 Skill 清单

新增或修改 Skill 时至少检查：

- frontmatter 有稳定 `name` 和清晰 `description`。
- 指令说明什么时候使用、输入是什么、输出是什么、失败时如何降级。
- 不要求 Agent 绕过 Policy/Approval 执行高风险工具。
- 不写死 provider key、Engine token、本地绝对路径或用户私有路径。
- 如果会生成 artifact，要说明 artifact 类型、质量标准和评估方式。
- 修改后运行 Resource Center 或 Engine skill scan 相关测试，确认 Skill 能被读取。

## 与 `.codex/skills` 的关系

`.codex/skills` 是历史项目内 Agent 能力说明，主要用于规划和记忆；新的运行时 Skill 以根目录 `.agent/skills` 为准。需要迁移历史能力时，应把稳定指令转成 `SKILL.md`，并保留安全边界。
