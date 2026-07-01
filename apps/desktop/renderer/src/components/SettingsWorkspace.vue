<script setup lang="ts">
import { LockKeyhole, MonitorCog, Shield } from 'lucide-vue-next'
</script>

<template>
  <section class="settings-grid">
    <article class="panel-surface settings-card">
      <MonitorCog :size="22" aria-hidden="true" />
      <h2>桌面工作台偏好</h2>
      <p>
        设置区先保留本地优先、中文界面和资源隔离原则。后续接入持久化后，所有配置仍通过 Main 转发到
        Go Engine。
      </p>
      <label class="check-row">
        <input type="checkbox" checked disabled />
        界面文字使用中文
      </label>
      <label class="check-row">
        <input type="checkbox" checked disabled />
        Renderer 不保存密钥
      </label>
    </article>

    <article class="panel-surface settings-card">
      <Shield :size="22" aria-hidden="true" />
      <h2>安全边界</h2>
      <p>
        Renderer 只通过 typed preload API 访问能力；Main 管理桌面生命周期和本地代理；Go Engine
        负责资源、策略和运行时。
      </p>
      <ul>
        <li>不直接访问 Node、文件系统、SQLite 或 secret。</li>
        <li>外部链接由 Main 安全打开。</li>
        <li>高风险能力必须进入 PolicyEngine 和 Approval。</li>
      </ul>
    </article>

    <article class="panel-surface settings-card">
      <LockKeyhole :size="22" aria-hidden="true" />
      <h2>密钥策略</h2>
      <p>
        API Key、MCP Secret 和远程连接令牌只允许保存在 Engine 侧。界面最多展示 hasApiKey、maskedKey
        和 maskedSecrets。
      </p>
    </article>
  </section>
</template>
