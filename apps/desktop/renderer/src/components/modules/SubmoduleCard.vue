<script setup lang="ts">
import { Boxes, ChevronRight } from 'lucide-vue-next'
import type { ProjectSubmodule } from '../../../../shared/dreamworker-api'
import { statusLabel } from '../../stores/workspace-navigation'

defineProps<{
  readonly submodule: ProjectSubmodule
  readonly active: boolean
}>()

defineEmits<{
  readonly select: [submoduleId: string]
}>()
</script>

<template>
  <article class="module-card submodule-card" :class="{ active }">
    <div class="submodule-card-head">
      <Boxes :size="18" aria-hidden="true" />
      <div>
        <strong>{{ submodule.displayName }}</strong>
        <span>{{ statusLabel(submodule.status) }}</span>
      </div>
    </div>
    <p>{{ submodule.summary }}</p>
    <small>{{ submodule.nextBestAction }}</small>
    <div class="submodule-meta">
      <span>{{ submodule.defaultAgents.length }} 个 Agent</span>
      <span>{{ submodule.enabledSkills.length }} 个 Skill</span>
      <span>{{ submodule.outputArtifacts.length }} 个产物</span>
    </div>
    <button type="button" @click="$emit('select', submodule.submoduleId)">
      查看子模块
      <ChevronRight :size="15" aria-hidden="true" />
    </button>
  </article>
</template>
