import { beforeEach, describe, expect, it, vi } from 'vitest'
import { normalizeExternalHttpUrl, openExternalHttpUrl } from './external-url'

const electronMock = vi.hoisted(() => ({
  openExternal: vi.fn()
}))

vi.mock('electron', () => ({
  shell: {
    openExternal: electronMock.openExternal
  }
}))

describe('external urls', () => {
  beforeEach(() => {
    electronMock.openExternal.mockReset()
  })

  it('normalizes only http and https urls', () => {
    expect(normalizeExternalHttpUrl(' http://localhost:20128/dashboard ')).toBe(
      'http://localhost:20128/dashboard'
    )
    expect(normalizeExternalHttpUrl('https://9router.ai/login')).toBe('https://9router.ai/login')
    expect(normalizeExternalHttpUrl('file:///C:/secret.txt')).toBeNull()
    expect(normalizeExternalHttpUrl('')).toBeNull()
  })

  it('opens safe external urls through Electron shell', async () => {
    electronMock.openExternal.mockResolvedValue(undefined)

    await expect(openExternalHttpUrl('https://9router.ai/login')).resolves.toEqual({
      ok: true,
      url: 'https://9router.ai/login',
      message: null
    })

    expect(electronMock.openExternal).toHaveBeenCalledWith('https://9router.ai/login')
  })

  it('rejects unsupported protocols before calling Electron shell', async () => {
    await expect(openExternalHttpUrl('dreamworker://settings')).resolves.toEqual({
      ok: false,
      url: 'dreamworker://settings',
      message: 'Only http and https URLs can be opened externally.'
    })

    expect(electronMock.openExternal).not.toHaveBeenCalled()
  })
})
