export interface ScrollMetrics {
  scrollTop: number
  scrollHeight: number
  clientHeight: number
}

export const CHAT_BOTTOM_THRESHOLD_PX = 96

export function scrollBottomDistance(metrics: ScrollMetrics): number {
  return Math.max(0, metrics.scrollHeight - metrics.clientHeight - metrics.scrollTop)
}

export function isNearScrollBottom(
  metrics: ScrollMetrics,
  thresholdPx = CHAT_BOTTOM_THRESHOLD_PX
): boolean {
  return scrollBottomDistance(metrics) <= thresholdPx
}
