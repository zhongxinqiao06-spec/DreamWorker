<script setup lang="ts">
import { LockKeyhole, MonitorCog, PlugZap, Shield } from 'lucide-vue-next'
import { useAppShellStore } from '../stores/app-shell'

const appShell = useAppShellStore()

function updateRunMode(event: Event): void {
  const value = (event.target as HTMLSelectElement).value
  void appShell.updateNineRouterSettings({
    nineRouterRunMode: value === 'managed' ? 'managed' : 'external'
  })
}

function updateText(
  field: 'nineRouterBaseURL' | 'nineRouterDashboardURL' | 'nineRouterDefaultModel',
  event: Event
): void {
  void appShell.updateNineRouterSettings({
    [field]: (event.target as HTMLInputElement).value
  })
}

function updateBoolean(
  field: 'enableNineRouterIntegration' | 'allowAgentsUseNineRouter',
  event: Event
): void {
  void appShell.updateNineRouterSettings({
    [field]: (event.target as HTMLInputElement).checked
  })
}
</script>

<template>
  <section class="settings-grid">
    <article class="panel-surface settings-card">
      <MonitorCog :size="22" aria-hidden="true" />
      <h2>桌面工作台偏好</h2>
      <p>
        设置区保留本地优先、中文界面和资源隔离原则。持久化后，所有配置仍通过 Main 转发到 Go Engine。
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

    <article class="panel-surface settings-card settings-extension-card">
      <PlugZap :size="22" aria-hidden="true" />
      <h2>9Router 拓展能力</h2>
      <p>
        9Router 可以作为 DreamWorker 的 OpenAI
        兼容上游模型路由。外部服务模式只连接本机服务，受管模式由 Go Engine
        管理安装、启动、停止、健康检查和日志。
      </p>
      <label>
        运行模式
        <select :value="appShell.settings.nineRouterRunMode" @change="updateRunMode">
          <option value="external">外部服务</option>
          <option value="managed">DreamWorker 受管</option>
        </select>
      </label>
      <label>
        Base URL
        <input
          :value="appShell.settings.nineRouterBaseURL"
          @change="updateText('nineRouterBaseURL', $event)"
        />
      </label>
      <label>
        Dashboard URL
        <input
          :value="appShell.settings.nineRouterDashboardURL"
          @change="updateText('nineRouterDashboardURL', $event)"
        />
      </label>
      <label>
        默认模型
        <input
          :value="appShell.settings.nineRouterDefaultModel"
          @change="updateText('nineRouterDefaultModel', $event)"
        />
      </label>
      <label class="check-row">
        <input
          :checked="appShell.settings.enableNineRouterIntegration"
          type="checkbox"
          @change="updateBoolean('enableNineRouterIntegration', $event)"
        />
        启用 9Router 集成
      </label>
      <label class="check-row">
        <input
          :checked="appShell.settings.allowAgentsUseNineRouter"
          type="checkbox"
          @change="updateBoolean('allowAgentsUseNineRouter', $event)"
        />
        允许 Agent 和聊天使用
      </label>
      <button type="button" @click="appShell.resetNineRouterSettings()">恢复默认配置</button>
    </article>
  </section>
</template>
