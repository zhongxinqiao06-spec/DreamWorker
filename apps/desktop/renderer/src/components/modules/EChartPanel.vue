<script setup lang="ts">
import * as echarts from 'echarts'
import type { ECharts, EChartsOption } from 'echarts'
import { nextTick, onBeforeUnmount, onMounted, ref, shallowRef, watch } from 'vue'

const props = withDefaults(
  defineProps<{
    readonly title: string
    readonly caption?: string
    readonly option: EChartsOption
    readonly height?: number
  }>(),
  {
    caption: '',
    height: 292
  }
)

const chartElement = ref<HTMLDivElement | null>(null)
const chart = shallowRef<ECharts | null>(null)
let resizeObserver: ResizeObserver | null = null
let resizeFrame = 0

function resizeChart(): void {
  chart.value?.resize()
}

function scheduleResize(): void {
  if (resizeFrame) {
    return
  }
  resizeFrame = window.requestAnimationFrame(() => {
    resizeFrame = 0
    resizeChart()
  })
}

function setChartOption(): void {
  chart.value?.setOption(props.option, { notMerge: true, lazyUpdate: true })
}

onMounted(() => {
  if (!chartElement.value) {
    return
  }
  chart.value = echarts.init(chartElement.value, undefined, { renderer: 'canvas' })
  setChartOption()
  resizeObserver = new ResizeObserver(scheduleResize)
  resizeObserver.observe(chartElement.value)
  void nextTick(scheduleResize)
})

watch(
  () => props.option,
  async () => {
    await nextTick()
    setChartOption()
    scheduleResize()
  }
)

onBeforeUnmount(() => {
  resizeObserver?.disconnect()
  resizeObserver = null
  if (resizeFrame) {
    window.cancelAnimationFrame(resizeFrame)
    resizeFrame = 0
  }
  chart.value?.dispose()
  chart.value = null
})
</script>

<template>
  <article class="echart-panel">
    <div class="section-heading-row">
      <div>
        <p class="eyebrow">ECharts</p>
        <h3>{{ title }}</h3>
      </div>
      <span v-if="caption">{{ caption }}</span>
    </div>
    <div
      ref="chartElement"
      class="echart-surface"
      :style="{ height: `${height}px` }"
      role="img"
      :aria-label="title"
    />
  </article>
</template>
