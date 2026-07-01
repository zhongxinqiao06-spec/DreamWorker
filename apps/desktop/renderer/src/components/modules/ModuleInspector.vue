<script setup lang="ts">
import { Play } from 'lucide-vue-next'
import { useAppShellStore } from '../../stores/app-shell'
import { statusLabel } from '../../stores/workspace-navigation'

const appShell = useAppShellStore()
</script>

<template>
  <aside class="right-panel" aria-label="模块检查器">
    <section class="inspector-card">
      <p class="eyebrow">当前子模块</p>
      <h3>{{ appShell.activeSubmodule?.displayName ?? '暂无子模块' }}</h3>
      <p>{{ appShell.activeSubmodule?.summary ?? '请选择一个子模块查看配置。' }}</p>
      <dl>
        <div>
          <dt>状态</dt>
          <dd>{{ statusLabel(appShell.activeSubmodule?.status) }}</dd>
        </div>
        <div>
          <dt>阶段</dt>
          <dd>{{ appShell.activeSubmodule?.config.stage ?? '暂无' }}</dd>
        </div>
        <div>
          <dt>项目</dt>
          <dd>{{ appShell.activeProject?.projectId ?? '暂无' }}</dd>
        </div>
      </dl>
    </section>

    <section class="inspector-card">
      <p class="eyebrow">能力组合</p>
      <h3>Agent / Skill / 工具</h3>
      <div class="tag-list">
        <span v-for="agent in appShell.activeSubmodule?.defaultAgents" :key="agent">{{
          agent
        }}</span>
        <span v-for="skill in appShell.activeSubmodule?.enabledSkills" :key="skill">{{
          skill
        }}</span>
        <span v-for="tool in appShell.activeSubmodule?.enabledTools" :key="tool">{{ tool }}</span>
      </div>
    </section>

    <section class="inspector-card">
      <p class="eyebrow">下一步</p>
      <h3>Next Best Action</h3>
      <p>{{ appShell.activeSubmodule?.nextBestAction ?? '等待子模块配置加载。' }}</p>
      <div class="tag-list">
        <span v-for="artifact in appShell.activeSubmodule?.outputArtifacts" :key="artifact">
          {{ artifact }}
        </span>
      </div>
      <button type="button">
        <Play :size="15" aria-hidden="true" />
        进入子模块
      </button>
    </section>
  </aside>
</template>
