<script setup lang="ts">
import { useAppShellStore } from './stores/app-shell'

const appShell = useAppShellStore()
</script>

<template>
  <main class="workspace-shell">
    <aside class="left-rail" aria-label="工作区导航">
      <div class="brand-mark">梦工</div>
      <nav class="rail-items">
        <button
          v-for="item in appShell.railItems"
          :key="item.id"
          class="rail-button"
          :class="{ active: item.id === appShell.activeRailItem }"
          type="button"
          @click="appShell.setActiveRailItem(item.id)"
        >
          {{ item.label }}
        </button>
      </nav>
    </aside>

    <section class="center-workspace">
      <header class="top-bar">
        <div>
          <p class="eyebrow">仓库初始化 / 桌面空壳</p>
          <h1>DreamWorker</h1>
        </div>
        <span class="status-pill">阶段 01-02</span>
      </header>

      <section class="mission-surface" aria-label="孵化器工作区">
        <div class="surface-header">
          <p class="eyebrow">孵化器工作台</p>
          <h2>{{ appShell.activeSurfaceTitle }}</h2>
        </div>
        <div class="surface-grid">
          <article
            v-for="surface in appShell.surfaces"
            :key="surface.id"
            class="surface-tile"
          >
            <span class="tile-index">{{ surface.index }}</span>
            <h3>{{ surface.title }}</h3>
            <p>{{ surface.summary }}</p>
          </article>
        </div>
      </section>
    </section>

    <aside class="right-inspector" aria-label="工作区检查器">
      <p class="eyebrow">空壳状态</p>
      <dl>
        <div>
          <dt>渲染层</dt>
          <dd>仅负责界面</dd>
        </div>
        <div>
          <dt>预加载层</dt>
          <dd>未开放接口</dd>
        </div>
        <div>
          <dt>引擎</dt>
          <dd>未连接</dd>
        </div>
      </dl>
    </aside>
  </main>
</template>
