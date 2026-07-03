import { afterEach, describe, expect, it, vi } from 'vitest'
import { rendererAssetUrl } from './asset-url'
import { providerLogoForProvider } from './provider-icons'

describe('renderer asset urls', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('resolves public assets relative to the packaged renderer document', () => {
    vi.stubGlobal('document', {
      baseURI:
        'file:///C:/Users/1/AppData/Local/Programs/DreamWorker/resources/app.asar/out/renderer/index.html'
    })

    expect(rendererAssetUrl('/provider-icons/deepseek.svg')).toBe(
      'file:///C:/Users/1/AppData/Local/Programs/DreamWorker/resources/app.asar/out/renderer/provider-icons/deepseek.svg'
    )
  })

  it('keeps relative asset paths when no browser globals are available', () => {
    expect(rendererAssetUrl('/provider-icons/openai.svg')).toBe('provider-icons/openai.svg')
  })

  it('uses packaged-safe provider icons for 9Router and normal providers', () => {
    vi.stubGlobal('document', {
      baseURI: 'file:///C:/DreamWorker/resources/app.asar/out/renderer/index.html'
    })

    expect(
      providerLogoForProvider({ providerId: 'provider_9router_local', providerType: 'openai' })
    ).toBe('file:///C:/DreamWorker/resources/app.asar/out/renderer/provider-icons/9router.svg')
    expect(
      providerLogoForProvider({ providerId: 'provider_deepseek', providerType: 'deepseek' })
    ).toBe('file:///C:/DreamWorker/resources/app.asar/out/renderer/provider-icons/deepseek.svg')
  })
})
