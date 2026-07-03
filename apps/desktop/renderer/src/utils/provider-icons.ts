import type { ProviderType, SafeModelProvider } from '../../../shared/dreamworker-api'
import { rendererAssetUrl } from './asset-url'

export function providerIconUrl(fileName: string): string {
  return rendererAssetUrl(`provider-icons/${fileName}`)
}

const providerLogoFiles: Record<ProviderType, string> = {
  deepseek: 'deepseek.svg',
  siliconflow: 'siliconflow.svg',
  glm: 'glm.png',
  openai: 'openai.svg',
  anthropic: 'anthropic.ico',
  openai_compatible: 'openai.svg',
  volcano: 'volcano.png',
  gemini: 'gemini.png',
  ollama: 'ollama.png',
  custom: 'openai.svg'
}

export function providerLogoForProvider(
  provider?: Pick<SafeModelProvider, 'providerId' | 'providerType'>
): string {
  if (provider?.providerId === 'provider_9router_local') {
    return providerIconUrl('9router.svg')
  }
  return provider
    ? providerIconUrl(providerLogoFiles[provider.providerType])
    : providerIconUrl('openai.svg')
}
