# DreamWorker Agent Skills

`.agent/` 是 DreamWorker 项目内 Skill 源目录。Go Engine 启动时会自动扫描 `.agent/skills/<skill-name>/SKILL.md`，把 Skill frontmatter 和 markdown instructions 载入内存，供 Resource Center 和 Chat Runtime 使用。

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

每个 Skill 是一个目录，目录名就是 command/name。目录内必须存在 `SKILL.md`：

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
- Skill 不能包含 raw secret、masked secret、MCP env value 或 provider raw response。
- Skill 生成和安装默认写入根目录 `.agent/skills/<skill-name>/SKILL.md`。
- Windows package 会把 `.agent` 作为 `extraResources` 一起打包。
