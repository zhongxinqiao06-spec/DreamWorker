<script setup lang="ts">
import { Boxes, ChevronRight } from 'lucide-vue-next'
import { computed } from 'vue'
import type { ProjectSubmodule } from '../../../../shared/dreamworker-api'
import { statusLabel } from '../../stores/workspace-navigation'

const props = defineProps<{
  readonly submodule: ProjectSubmodule
  readonly active: boolean
}>()

const emit = defineEmits<{
  readonly select: [submoduleId: string]
  readonly enter: [submoduleId: string]
}>()

const isCodingAgent = computed(() => props.submodule.submoduleId === 'coding_agent')
const canEnterDetail = computed(
  () => props.submodule.submoduleId === 'requirement_analysis' || isCodingAgent.value
)

function openSubmodule(): void {
  emit('select', props.submodule.submoduleId)
  if (canEnterDetail.value) {
    emit('enter', props.submodule.submoduleId)
  }
}
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
      <template v-if="isCodingAgent">
        <span>3 Engine</span>
        <span>文件树</span>
        <span>直接写入</span>
      </template>
      <template v-else>
        <span>{{ submodule.defaultAgents.length }} 个 Agent</span>
        <span>{{ submodule.enabledSkills.length }} 个 Skill</span>
        <span>{{ submodule.outputArtifacts.length }} 个产物</span>
      </template>
    </div>
    <button type="button" @click="openSubmodule">
      查看子模块
      <ChevronRight :size="15" aria-hidden="true" />
    </button>
  </article>
</template>
