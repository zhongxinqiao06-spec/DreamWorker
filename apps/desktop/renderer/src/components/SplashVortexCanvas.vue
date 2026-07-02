<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
import {
  createVortexParticles,
  stepVortexParticles,
  type VortexParticle
} from '../utils/vortex-particles'

const canvasRef = ref<HTMLCanvasElement | null>(null)
let context: CanvasRenderingContext2D | null = null
let frameId = 0
let lastFrameAt = 0
let startedAt = 0
let logicalWidth = 1
let logicalHeight = 1
let particles: VortexParticle[] = []
let reducedMotionQuery: MediaQueryList | null = null
let reducedMotion = false
const splashPalette = ['#7c3aed', '#2563eb', '#14b8a6', '#ede7ff'] as const

function targetParticleCount(width: number, height: number): number {
  if (reducedMotion) {
    return 72
  }
  return Math.min(280, Math.max(150, Math.round((width * height) / 5200)))
}

function resizeCanvas(): void {
  const canvas = canvasRef.value
  if (!canvas) {
    return
  }

  const rect = canvas.getBoundingClientRect()
  const nextWidth = Math.max(1, Math.floor(rect.width))
  const nextHeight = Math.max(1, Math.floor(rect.height))
  const dpr = Math.min(window.devicePixelRatio || 1, 2)
  logicalWidth = nextWidth
  logicalHeight = nextHeight
  canvas.width = Math.floor(nextWidth * dpr)
  canvas.height = Math.floor(nextHeight * dpr)
  canvas.style.width = `${nextWidth}px`
  canvas.style.height = `${nextHeight}px`

  context = canvas.getContext('2d')
  context?.setTransform(dpr, 0, 0, dpr, 0, 0)

  const nextCount = targetParticleCount(nextWidth, nextHeight)
  if (particles.length !== nextCount) {
    particles = createVortexParticles(nextCount, nextWidth, nextHeight)
  }
}

function drawCore(ctx: CanvasRenderingContext2D, elapsedMs: number): void {
  const centerX = logicalWidth * 0.5
  const centerY = logicalHeight * 0.5
  const pulse = reducedMotion ? 0.5 : 0.5 + Math.sin(elapsedMs * 0.002) * 0.5
  const coreRadius = Math.min(logicalWidth, logicalHeight) * (0.055 + pulse * 0.012)

  ctx.save()
  ctx.globalCompositeOperation = 'lighter'
  ctx.lineWidth = 1.4
  ctx.strokeStyle = `rgba(124, 58, 237, ${0.16 + pulse * 0.18})`
  ctx.beginPath()
  ctx.arc(centerX, centerY, coreRadius, 0, Math.PI * 2)
  ctx.stroke()
  ctx.strokeStyle = `rgba(20, 184, 166, ${0.12 + pulse * 0.14})`
  ctx.beginPath()
  ctx.arc(centerX, centerY, coreRadius * 1.62, Math.PI * 0.18, Math.PI * 1.42)
  ctx.stroke()
  ctx.restore()
}

function render(time: number): void {
  const canvas = canvasRef.value
  const ctx = context
  if (!canvas || !ctx) {
    return
  }

  if (!startedAt) {
    startedAt = time
  }
  const deltaMs = lastFrameAt ? time - lastFrameAt : 16
  lastFrameAt = time
  const elapsedMs = time - startedAt

  ctx.clearRect(0, 0, logicalWidth, logicalHeight)
  ctx.fillStyle = 'rgba(248, 250, 252, 0.38)'
  ctx.fillRect(0, 0, logicalWidth, logicalHeight)

  const haze = ctx.createRadialGradient(
    logicalWidth * 0.5,
    logicalHeight * 0.5,
    0,
    logicalWidth * 0.5,
    logicalHeight * 0.5,
    Math.max(logicalWidth, logicalHeight) * 0.68
  )
  haze.addColorStop(0, 'rgba(124, 58, 237, 0.08)')
  haze.addColorStop(0.32, 'rgba(37, 99, 235, 0.06)')
  haze.addColorStop(0.72, 'rgba(20, 184, 166, 0.04)')
  haze.addColorStop(1, 'rgba(248, 250, 252, 0)')
  ctx.fillStyle = haze
  ctx.fillRect(0, 0, logicalWidth, logicalHeight)

  const frames = stepVortexParticles(particles, {
    width: logicalWidth,
    height: logicalHeight,
    deltaMs,
    elapsedMs,
    reducedMotion,
    palette: splashPalette
  })

  ctx.save()
  ctx.globalCompositeOperation = 'lighter'
  for (const frame of frames) {
    ctx.globalAlpha = frame.alpha
    ctx.fillStyle = frame.color
    ctx.beginPath()
    ctx.arc(frame.x, frame.y, frame.size, 0, Math.PI * 2)
    ctx.fill()
  }
  ctx.restore()

  drawCore(ctx, elapsedMs)
  frameId = window.requestAnimationFrame(render)
}

function syncReducedMotion(): void {
  reducedMotion = Boolean(reducedMotionQuery?.matches)
  resizeCanvas()
}

onMounted(() => {
  reducedMotionQuery = window.matchMedia('(prefers-reduced-motion: reduce)')
  reducedMotionQuery.addEventListener('change', syncReducedMotion)
  syncReducedMotion()
  window.addEventListener('resize', resizeCanvas)
  frameId = window.requestAnimationFrame(render)
})

onBeforeUnmount(() => {
  if (frameId) {
    window.cancelAnimationFrame(frameId)
  }
  window.removeEventListener('resize', resizeCanvas)
  reducedMotionQuery?.removeEventListener('change', syncReducedMotion)
})
</script>

<template>
  <section class="splash-vortex" role="status" aria-label="DreamWorker 正在启动">
    <canvas ref="canvasRef" class="splash-vortex__canvas" aria-hidden="true"></canvas>
    <div class="splash-vortex__brand" aria-hidden="true">
      <span>DreamWorker AI OS 2.0</span>
      <i></i>
    </div>
  </section>
</template>

<style scoped>
.splash-vortex {
  position: fixed;
  inset: 0;
  z-index: 1000;
  overflow: hidden;
  background:
    radial-gradient(circle at 50% 46%, rgba(124, 58, 237, 0.14), transparent 28%),
    radial-gradient(circle at 58% 53%, rgba(37, 99, 235, 0.1), transparent 34%),
    radial-gradient(circle at 42% 58%, rgba(20, 184, 166, 0.08), transparent 36%),
    linear-gradient(135deg, #f8fafc 0%, #ffffff 46%, #eef4ff 100%);
  pointer-events: none;
}

.splash-vortex__canvas {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
}

.splash-vortex__brand {
  position: absolute;
  left: 50%;
  bottom: 11vh;
  display: grid;
  width: min(280px, 44vw);
  transform: translateX(-50%);
  gap: 14px;
  justify-items: center;
  color: #0f172a;
  font-size: 13px;
  font-weight: 900;
  letter-spacing: 0;
  text-transform: uppercase;
}

.splash-vortex__brand i {
  display: block;
  width: 100%;
  height: 2px;
  overflow: hidden;
  border-radius: 999px;
  background: rgba(148, 163, 184, 0.2);
}

.splash-vortex__brand i::before {
  display: block;
  width: 46%;
  height: 100%;
  animation: splash-scan 1.6s ease-in-out infinite;
  border-radius: inherit;
  background: linear-gradient(90deg, #7c3aed, #2563eb, #14b8a6);
  content: '';
}

@keyframes splash-scan {
  0% {
    transform: translateX(-120%);
  }

  100% {
    transform: translateX(240%);
  }
}

@media (prefers-reduced-motion: reduce) {
  .splash-vortex__brand i::before {
    animation-duration: 4.8s;
  }
}
</style>
