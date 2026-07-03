import { shell } from 'electron'
import type { OpenExternalResult } from '../shared/dreamworker-api'

const EXTERNAL_URL_PROTOCOLS = new Set(['http:', 'https:'])

export function normalizeExternalHttpUrl(rawUrl: string): string | null {
  const trimmed = rawUrl.trim()
  if (!trimmed) {
    return null
  }

  try {
    const url = new URL(trimmed)
    if (!EXTERNAL_URL_PROTOCOLS.has(url.protocol)) {
      return null
    }
    return url.toString()
  } catch {
    return null
  }
}

export async function openExternalHttpUrl(rawUrl: string): Promise<OpenExternalResult> {
  const normalizedUrl = normalizeExternalHttpUrl(rawUrl)
  if (!normalizedUrl) {
    return {
      ok: false,
      url: rawUrl,
      message: 'Only http and https URLs can be opened externally.'
    }
  }

  try {
    await shell.openExternal(normalizedUrl)
    return {
      ok: true,
      url: normalizedUrl,
      message: null
    }
  } catch (error) {
    return {
      ok: false,
      url: normalizedUrl,
      message: error instanceof Error ? error.message : String(error)
    }
  }
}
