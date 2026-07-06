<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
import {
  createVortexParticles,
  stepVortexParticles,
  type VortexParticle
} from '../utils/vortex-particles'

const canvasRef = ref<HTMLCanvasElement | null>(null)
const splashRef = ref<HTMLElement | null>(null)
let context: CanvasRenderingContext2D | null = null
let frameId = 0
let resizeFrameId = 0
let lastFrameAt = 0
let startedAt = 0
let logicalWidth = 1
let logicalHeight = 1
let particles: VortexParticle[] = []
let resizeObserver: ResizeObserver | null = null
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

  const stage = splashRef.value ?? canvas.parentElement
  const rect = stage?.getBoundingClientRect()
  const nextWidth = Math.max(1, Math.floor(rect?.width || window.innerWidth))
  const nextHeight = Math.max(1, Math.floor(rect?.height || window.innerHeight))
  const dpr = Math.min(window.devicePixelRatio || 1, 2)
  logicalWidth = nextWidth
  logicalHeight = nextHeight
  canvas.width = Math.floor(nextWidth * dpr)
  canvas.height = Math.floor(nextHeight * dpr)
  canvas.style.width = '100%'
  canvas.style.height = '100%'

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

function scheduleCanvasResize(): void {
  if (resizeFrameId) {
    window.cancelAnimationFrame(resizeFrameId)
  }
  resizeFrameId = window.requestAnimationFrame(() => {
    resizeFrameId = 0
    resizeCanvas()
  })
}

function syncReducedMotion(): void {
  reducedMotion = Boolean(reducedMotionQuery?.matches)
  scheduleCanvasResize()
}

onMounted(() => {
  reducedMotionQuery = window.matchMedia('(prefers-reduced-motion: reduce)')
  reducedMotionQuery.addEventListener('change', syncReducedMotion)
  syncReducedMotion()
  if (typeof ResizeObserver !== 'undefined' && splashRef.value) {
    resizeObserver = new ResizeObserver(scheduleCanvasResize)
    resizeObserver.observe(splashRef.value)
  }
  window.addEventListener('resize', scheduleCanvasResize)
  scheduleCanvasResize()
  frameId = window.requestAnimationFrame(render)
})

onBeforeUnmount(() => {
  if (frameId) {
    window.cancelAnimationFrame(frameId)
  }
  if (resizeFrameId) {
    window.cancelAnimationFrame(resizeFrameId)
  }
  resizeObserver?.disconnect()
  window.removeEventListener('resize', scheduleCanvasResize)
  reducedMotionQuery?.removeEventListener('change', syncReducedMotion)
})
</script>

<template>
  <section ref="splashRef" class="splash-vortex" role="status" aria-label="DreamWorker 正在启动">
    <canvas ref="canvasRef" class="splash-vortex__canvas" aria-hidden="true"></canvas>
    <div class="splash-vortex__frame" aria-hidden="true">
      <span></span>
      <span></span>
      <span></span>
      <span></span>
    </div>
    <div class="splash-vortex__brand" aria-hidden="true">
      <span class="splash-vortex__mark">
        <img src="/aios/brand-mark.png" alt="" />
      </span>
      <strong>DreamWorker AI OS 2.0</strong>
      <p>正在装载项目工作流</p>
      <div class="splash-vortex__chips">
        <span>Agent</span>
        <span>Skill</span>
        <span>Workflow</span>
      </div>
      <i></i>
    </div>
  </section>
</template>

<style scoped>
.splash-vortex {
  position: fixed;
  inset: 0;
  z-index: 9999;
  display: grid;
  width: 100vw;
  height: 100vh;
  min-height: 100dvh;
  overflow: hidden;
  place-items: center;
  background:
    linear-gradient(rgba(124, 58, 237, 0.055) 1px, transparent 1px),
    linear-gradient(90deg, rgba(37, 99, 235, 0.05) 1px, transparent 1px),
    radial-gradient(ellipse at 50% 47%, rgba(124, 58, 237, 0.16), transparent 30%),
    radial-gradient(ellipse at 58% 52%, rgba(37, 99, 235, 0.12), transparent 34%),
    radial-gradient(ellipse at 42% 58%, rgba(20, 184, 166, 0.09), transparent 38%),
    linear-gradient(135deg, #f8fbff 0%, #ffffff 42%, #eef5ff 100%);
  background-size:
    42px 42px,
    42px 42px,
    auto,
    auto,
    auto,
    auto;
  isolation: isolate;
  pointer-events: none;
}

.splash-vortex::before,
.splash-vortex::after {
  position: absolute;
  z-index: 1;
  border-radius: 999px;
  content: '';
}

.splash-vortex::before {
  width: min(76vw, 980px);
  height: min(76vw, 980px);
  border: 1px solid rgba(124, 58, 237, 0.18);
  box-shadow:
    inset 0 0 70px rgba(37, 99, 235, 0.06),
    0 0 120px rgba(124, 58, 237, 0.12);
  transform: rotate(-18deg) scaleX(1.24);
}

.splash-vortex::after {
  width: min(54vw, 680px);
  height: min(54vw, 680px);
  border: 1px solid rgba(20, 184, 166, 0.16);
  box-shadow: inset 0 0 50px rgba(20, 184, 166, 0.08);
  transform: rotate(16deg) scaleX(1.38);
}

.splash-vortex__canvas {
  position: absolute;
  inset: 0;
  z-index: 0;
  width: 100%;
  height: 100%;
}

.splash-vortex__frame {
  position: absolute;
  inset: clamp(18px, 4vw, 64px);
  z-index: 2;
}

.splash-vortex__frame span {
  position: absolute;
  width: clamp(62px, 8vw, 124px);
  height: clamp(62px, 8vw, 124px);
  border-color: rgba(124, 58, 237, 0.24);
  border-style: solid;
}

.splash-vortex__frame span:nth-child(1) {
  top: 0;
  left: 0;
  border-width: 1px 0 0 1px;
}

.splash-vortex__frame span:nth-child(2) {
  top: 0;
  right: 0;
  border-width: 1px 1px 0 0;
}

.splash-vortex__frame span:nth-child(3) {
  right: 0;
  bottom: 0;
  border-width: 0 1px 1px 0;
}

.splash-vortex__frame span:nth-child(4) {
  bottom: 0;
  left: 0;
  border-width: 0 0 1px 1px;
}

.splash-vortex__brand {
  position: relative;
  z-index: 3;
  display: grid;
  width: min(520px, calc(100vw - 48px));
  padding: clamp(28px, 5vh, 48px) clamp(24px, 5vw, 48px);
  border: 1px solid rgba(124, 58, 237, 0.18);
  border-radius: 28px;
  background:
    linear-gradient(135deg, rgba(255, 255, 255, 0.78), rgba(248, 250, 252, 0.5)),
    linear-gradient(90deg, rgba(124, 58, 237, 0.08), rgba(20, 184, 166, 0.05));
  box-shadow:
    0 28px 90px rgba(15, 23, 42, 0.12),
    inset 0 1px 0 rgba(255, 255, 255, 0.82);
  gap: 12px;
  justify-items: center;
  backdrop-filter: blur(22px);
  color: #0f172a;
  text-align: center;
}

.splash-vortex__mark {
  display: grid;
  width: 76px;
  height: 76px;
  place-items: center;
  border: 1px solid rgba(124, 58, 237, 0.2);
  border-radius: 24px;
  background:
    linear-gradient(145deg, rgba(255, 255, 255, 0.9), rgba(237, 231, 255, 0.62)),
    rgba(255, 255, 255, 0.72);
  box-shadow:
    0 18px 50px rgba(124, 58, 237, 0.22),
    inset 0 1px 0 rgba(255, 255, 255, 0.9);
}

.splash-vortex__mark img {
  width: 44px;
  height: 44px;
}

.splash-vortex__brand strong {
  margin-top: 4px;
  color: #0f172a;
  font-size: clamp(20px, 3vw, 30px);
  font-weight: 950;
  letter-spacing: 0;
  line-height: 1.15;
}

.splash-vortex__brand p {
  margin: 0;
  color: #536179;
  font-size: 14px;
  font-weight: 700;
  line-height: 1.6;
}

.splash-vortex__chips {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 8px;
  margin: 6px 0 2px;
}

.splash-vortex__chips span {
  display: inline-grid;
  min-height: 28px;
  padding: 0 12px;
  place-items: center;
  border: 1px solid rgba(37, 99, 235, 0.14);
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.72);
  box-shadow: 0 10px 26px rgba(37, 99, 235, 0.08);
  color: #4f46e5;
  font-size: 12px;
  font-weight: 900;
}

.splash-vortex__brand i {
  display: block;
  width: min(360px, 78vw);
  height: 3px;
  margin-top: 6px;
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

@media (max-width: 720px) {
  .splash-vortex__brand {
    border-radius: 22px;
  }

  .splash-vortex__frame {
    inset: 18px;
  }

  .splash-vortex__mark {
    width: 64px;
    height: 64px;
    border-radius: 20px;
  }

  .splash-vortex__mark img {
    width: 38px;
    height: 38px;
  }
}
</style>
