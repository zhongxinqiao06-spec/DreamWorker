import { defineStore } from 'pinia'

type RailItemId = 'mission' | 'blueprint' | 'runs' | 'artifacts'

type RailItem = {
  readonly id: RailItemId
  readonly label: string
}

type Surface = {
  readonly id: string
  readonly index: string
  readonly title: string
  readonly summary: string
}

export const useAppShellStore = defineStore('app-shell', {
  state: () => ({
    activeRailItem: 'mission' as RailItemId,
    railItems: [
      { id: 'mission', label: '想法对话' },
      { id: 'blueprint', label: '孵化看板' },
      { id: 'runs', label: '运行记录' },
      { id: 'artifacts', label: '产物中心' }
    ] satisfies RailItem[],
    surfaces: [
      {
        id: 'idea-chat',
        index: '01',
        title: '想法对话',
        summary: '任务录入空壳'
      },
      {
        id: 'incubator-board',
        index: '02',
        title: '孵化看板',
        summary: '阶段概览空壳'
      },
      {
        id: 'run-timeline',
        index: '03',
        title: '运行时间线',
        summary: '事件流空壳'
      },
      {
        id: 'artifact-studio',
        index: '04',
        title: '产物工作室',
        summary: '交付物空壳'
      }
    ] satisfies Surface[]
  }),
  getters: {
    activeSurfaceTitle: (state): string => {
      const activeItem = state.railItems.find((item) => item.id === state.activeRailItem)
      return activeItem?.label ?? '想法对话'
    }
  },
  actions: {
    setActiveRailItem(id: RailItemId): void {
      this.activeRailItem = id
    }
  }
})
