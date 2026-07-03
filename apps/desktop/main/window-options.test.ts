import { describe, expect, it } from 'vitest'
import { createMainWindowOptions } from './window-options'

describe('main window security options', () => {
  it('keeps the renderer sandboxed and isolated', () => {
    const options = createMainWindowOptions(
      'C:\\project\\DreamWorker\\apps\\desktop\\out\\preload\\index.cjs'
    )

    expect(options.webPreferences?.preload).toContain('preload')
    expect(options.webPreferences?.contextIsolation).toBe(true)
    expect(options.webPreferences?.nodeIntegration).toBe(false)
    expect(options.webPreferences?.sandbox).toBe(true)
  })

  it('opens wide enough for the three-column workspace by default', () => {
    const options = createMainWindowOptions(
      'C:\\project\\DreamWorker\\apps\\desktop\\out\\preload\\index.cjs'
    )

    expect(options.width).toBeGreaterThanOrEqual(1600)
    expect(options.minWidth).toBeGreaterThan(1240)
    expect(options.height).toBeGreaterThanOrEqual(900)
  })

  it('uses a themed title bar overlay for the AI OS shell', () => {
    const options = createMainWindowOptions(
      'C:\\project\\DreamWorker\\apps\\desktop\\out\\preload\\index.cjs'
    )

    expect(options.title).toBe('DreamWorker AI 工作台')
    expect(options.titleBarStyle).toBe('hidden')
    expect(options.titleBarOverlay).toEqual({
      color: '#f8fafc',
      symbolColor: '#0f172a',
      height: 40
    })
  })
})
