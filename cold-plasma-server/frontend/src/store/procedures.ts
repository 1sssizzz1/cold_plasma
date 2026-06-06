import { create } from 'zustand'
import type { Procedure } from '../types'
import { apiGet } from '../utils/api'

type ProceduresState = {
  items: Procedure[]
  loading: boolean
  error: string | null
  fetchAll: () => Promise<void>
}

export const useProceduresStore = create<ProceduresState>((set, get) => ({
  items: [],
  loading: false,
  error: null,

  fetchAll: async () => {
    if (get().loading) return
    set({ loading: true, error: null })
    try {
      const items = await apiGet<Procedure[]>('/procedures')
      set({ items })
    } catch (e) {
      set({ error: (e as Error).message })
    } finally {
      set({ loading: false })
    }
  },
}))

