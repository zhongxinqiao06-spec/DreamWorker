import { describe, expect, it } from 'vitest'
import { isNearScrollBottom, scrollBottomDistance } from './chat-scroll'

describe('chat scroll helpers', () => {
  it('detects when the chat thread is pinned near the bottom', () => {
    expect(
      isNearScrollBottom({
        scrollTop: 904,
        scrollHeight: 1600,
        clientHeight: 620
      })
    ).toBe(true)
  })

  it('detects when the user has scrolled away from the bottom', () => {
    expect(
      isNearScrollBottom({
        scrollTop: 600,
        scrollHeight: 1600,
        clientHeight: 620
      })
    ).toBe(false)
  })

  it('never reports negative bottom distance', () => {
    expect(
      scrollBottomDistance({
        scrollTop: 1400,
        scrollHeight: 1200,
        clientHeight: 620
      })
    ).toBe(0)
  })
})
