export interface VortexParticle {
  angle: number
  radius: number
  depth: number
  angularVelocity: number
  radialVelocity: number
  size: number
  colorIndex: number
  shimmer: number
}

export interface VortexParticleFrame {
  x: number
  y: number
  size: number
  alpha: number
  color: string
}

export interface VortexStepOptions {
  width: number
  height: number
  deltaMs: number
  elapsedMs: number
  reducedMotion?: boolean
  palette?: readonly string[]
}

export const VORTEX_PALETTE = ['#9b73ff', '#6fe8ff', '#7df2b0', '#ffd166'] as const

function randomBetween(random: () => number, min: number, max: number): number {
  return min + (max - min) * random()
}

function maxRadiusFor(width: number, height: number): number {
  return Math.max(180, Math.hypot(width, height) * 0.58)
}

export function resetVortexParticle(
  particle: VortexParticle,
  width: number,
  height: number,
  random: () => number = Math.random
): VortexParticle {
  const maxRadius = maxRadiusFor(width, height)
  particle.angle = randomBetween(random, 0, Math.PI * 2)
  particle.radius = randomBetween(random, maxRadius * 0.72, maxRadius)
  particle.depth = randomBetween(random, 0.08, 0.48)
  particle.angularVelocity = randomBetween(random, 0.32, 1.28) * (random() > 0.18 ? 1 : -1)
  particle.radialVelocity = randomBetween(random, 18, 46)
  particle.size = randomBetween(random, 0.85, 2.8)
  particle.colorIndex = Math.floor(random() * VORTEX_PALETTE.length)
  particle.shimmer = randomBetween(random, 0, Math.PI * 2)
  return particle
}

export function createVortexParticles(
  count: number,
  width: number,
  height: number,
  random: () => number = Math.random
): VortexParticle[] {
  return Array.from({ length: Math.max(0, count) }, () =>
    resetVortexParticle({} as VortexParticle, width, height, random)
  )
}

export function stepVortexParticles(
  particles: VortexParticle[],
  options: VortexStepOptions,
  random: () => number = Math.random
): VortexParticleFrame[] {
  const width = Math.max(1, options.width)
  const height = Math.max(1, options.height)
  const palette = options.palette ?? VORTEX_PALETTE
  const deltaMs = Math.min(Math.max(options.deltaMs, 0), 48)
  const motionScale = options.reducedMotion ? 0.12 : 1
  const elapsed = options.elapsedMs * 0.001
  const centerX = width * 0.5 + Math.sin(elapsed * 0.35) * width * 0.025 * motionScale
  const centerY = height * 0.5 + Math.cos(elapsed * 0.28) * height * 0.025 * motionScale
  const maxRadius = maxRadiusFor(width, height)

  return particles.map((particle) => {
    const drift = deltaMs * 0.001 * motionScale
    particle.angle += particle.angularVelocity * drift
    particle.radius -= particle.radialVelocity * drift
    particle.depth = Math.min(1, particle.depth + 0.08 * drift)
    particle.shimmer += deltaMs * 0.002 * motionScale

    if (particle.radius < 12 || particle.radius > maxRadius * 1.08) {
      resetVortexParticle(particle, width, height, random)
    }

    const perspective = 0.34 + particle.depth * 0.92
    const spiralPull = Math.sin(particle.depth * Math.PI) * 18
    const x = centerX + Math.cos(particle.angle) * (particle.radius * perspective - spiralPull)
    const y = centerY + Math.sin(particle.angle) * (particle.radius * perspective - spiralPull)
    const pulse = 0.72 + Math.sin(particle.shimmer) * 0.18
    const alpha = Math.min(0.94, Math.max(0.16, (0.2 + particle.depth * 0.62) * pulse))
    const size = particle.size * (0.72 + particle.depth * 1.35)

    return {
      x,
      y,
      size,
      alpha,
      color: palette[particle.colorIndex % palette.length] ?? VORTEX_PALETTE[0]
    }
  })
}
