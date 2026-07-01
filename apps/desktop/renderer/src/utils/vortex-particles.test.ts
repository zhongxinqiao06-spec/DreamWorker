import { describe, expect, it } from 'vitest'
import { createVortexParticles, stepVortexParticles, type VortexParticle } from './vortex-particles'

function seededRandom(seed = 7): () => number {
  let value = seed
  return () => {
    value = (value * 16807) % 2147483647
    return (value - 1) / 2147483646
  }
}

describe('vortex particles', () => {
  it('creates a stable particle count with initialized values', () => {
    const particles = createVortexParticles(24, 1280, 720, seededRandom())

    expect(particles).toHaveLength(24)
    expect(particles.every((particle) => particle.radius > 0)).toBe(true)
    expect(particles.every((particle) => particle.depth >= 0 && particle.depth <= 1)).toBe(true)
  })

  it('steps particles into drawable frames', () => {
    const random = seededRandom()
    const particles = createVortexParticles(4, 1024, 768, random)
    const firstAngle = particles[0]?.angle ?? 0
    const frames = stepVortexParticles(
      particles,
      {
        width: 1024,
        height: 768,
        deltaMs: 16,
        elapsedMs: 160
      },
      random
    )

    expect(frames).toHaveLength(4)
    expect(frames[0]?.x).toBeTypeOf('number')
    expect(frames[0]?.y).toBeTypeOf('number')
    expect(frames[0]?.alpha).toBeGreaterThan(0)
    expect(particles[0]?.angle).not.toBe(firstAngle)
  })

  it('resets particles that collapse into the vortex center', () => {
    const particle: VortexParticle = {
      angle: 0,
      radius: 1,
      depth: 0.9,
      angularVelocity: 1,
      radialVelocity: 20,
      size: 1,
      colorIndex: 0,
      shimmer: 0
    }

    stepVortexParticles(
      [particle],
      {
        width: 900,
        height: 600,
        deltaMs: 16,
        elapsedMs: 40
      },
      seededRandom()
    )

    expect(particle.radius).toBeGreaterThan(100)
    expect(particle.depth).toBeLessThan(0.6)
  })

  it('keeps reduced motion updates deliberately slow', () => {
    const normal = createVortexParticles(1, 1200, 800, seededRandom(3))
    const reduced = createVortexParticles(1, 1200, 800, seededRandom(3))
    const normalAngle = normal[0]?.angle ?? 0
    const reducedAngle = reduced[0]?.angle ?? 0

    stepVortexParticles(normal, { width: 1200, height: 800, deltaMs: 32, elapsedMs: 1000 })
    stepVortexParticles(reduced, {
      width: 1200,
      height: 800,
      deltaMs: 32,
      elapsedMs: 1000,
      reducedMotion: true
    })

    expect(Math.abs((reduced[0]?.angle ?? 0) - reducedAngle)).toBeLessThan(
      Math.abs((normal[0]?.angle ?? 0) - normalAngle)
    )
  })
})
