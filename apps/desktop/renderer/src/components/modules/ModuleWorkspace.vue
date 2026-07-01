<script setup lang="ts">
import { computed } from 'vue'
import ModuleInspector from './ModuleInspector.vue'
import ProjectContextPanel from './ProjectContextPanel.vue'
import SubmoduleCard from './SubmoduleCard.vue'
import { useAppShellStore } from '../../stores/app-shell'
import { moduleShortTitle, moduleTitle, statusLabel } from '../../stores/workspace-navigation'

const appShell = useAppShellStore()

const moduleId = computed(() => appShell.activeModuleWorkspace)
const module = computed(() => appShell.activeModule)
</script>

<template>
  <section class="workspace-layout module-workspace-layout">
    <ProjectContextPanel />

    <section class="module-center panel-surface" aria-label="项目闭环模块">
      <div class="project-header">
        <div>
          <p class="eyebrow">{{ moduleShortTitle(moduleId ?? 'explore') }}工作台</p>
          <h2>{{ moduleId ? moduleTitle(moduleId) : '模块工作台' }}</h2>
          <p>
            {{
              module?.summary ??
              '请选择项目后查看当前闭环模块。项目配置只在项目页维护，运行模块从这里进入。'
            }}
          </p>
        </div>
        <div class="context-pills">
          <span>{{ statusLabel(module?.status) }}</span>
          <span>{{ module?.defaultAgents.length ?? 0 }} 个 Agent</span>
          <span>{{ module?.enabledSkills.length ?? 0 }} 个 Skill</span>
        </div>
      </div>

      <section v-if="module?.submodules.length" class="module-grid submodule-grid">
        <SubmoduleCard
          v-for="submodule in module.submodules"
          :key="submodule.submoduleId"
          :submodule="submodule"
          :active="appShell.activeSubmodule?.submoduleId === submodule.submoduleId"
          @select="moduleId && appShell.selectSubmodule(moduleId, $event)"
        />
      </section>

      <section v-else class="placeholder-panel">
        <h3>暂无子模块</h3>
        <p>请先在项目页新增项目，或等待 Engine 返回当前项目的模块配置。</p>
      </section>
    </section>

    <ModuleInspector />
  </section>
</template>
