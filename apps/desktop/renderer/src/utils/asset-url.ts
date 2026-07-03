export function rendererAssetUrl(relativePath: string): string {
  const normalized = relativePath.replace(/^\/+/, '')
  if (typeof document !== 'undefined' && document.baseURI) {
    return new URL(normalized, document.baseURI).href
  }
  if (typeof window !== 'undefined' && window.location?.href) {
    return new URL(normalized, window.location.href).href
  }
  return normalized
}
