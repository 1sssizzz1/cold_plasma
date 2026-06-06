import { create } from 'zustand'
import type { User } from '../types'

type AuthState = {
  token: string | null
  user: User | null
  hydrated: boolean
  setAuth: (token: string, user: User) => void
  logout: () => void
  hydrate: () => void
}

const LS_TOKEN = 'cp_token'
const LS_USER = 'cp_user'

export const useAuthStore = create<AuthState>((set, get) => ({
  token: null,
  user: null,
  hydrated: false,

  setAuth: (token, user) => {
    localStorage.setItem(LS_TOKEN, token)
    localStorage.setItem(LS_USER, JSON.stringify(user))
    set({ token, user })
  },

  logout: () => {
    localStorage.removeItem(LS_TOKEN)
    localStorage.removeItem(LS_USER)
    set({ token: null, user: null })
  },

  hydrate: () => {
    if (get().hydrated) return
    let token = localStorage.getItem(LS_TOKEN)
    const rawUser = localStorage.getItem(LS_USER)
    let user: User | null = null
    if (rawUser) {
      try {
        user = JSON.parse(rawUser) as User
      } catch {
        localStorage.removeItem(LS_TOKEN)
        localStorage.removeItem(LS_USER)
        token = null
      }
    }
    set({ token: token || null, user, hydrated: true })
  },
}))
