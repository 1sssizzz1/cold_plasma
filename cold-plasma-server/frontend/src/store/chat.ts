import { create } from 'zustand'
import { apiPost } from '../utils/api'
import { useAuthStore } from './auth'

export type ChatMessage = {
  id: string
  role: 'user' | 'assistant'
  text: string
  ts: number
}

type ChatState = {
  open: boolean
  sending: boolean
  sessionId: string | null
  messages: ChatMessage[]
  toggle: () => void
  openChat: () => void
  close: () => void
  send: (text: string, onIntent?: (intent: string) => void) => Promise<void>
  reset: () => void
}

const welcomeText =
  'Здравствуйте! Подскажу по холодной плазме, уходу за кожей и волосами. Что вас беспокоит — кожа, веки, акне или волосы?'

function uid() {
  return Math.random().toString(16).slice(2) + Date.now().toString(16)
}

function newSessionId() {
  return `ai_${uid()}`
}

function welcomeMessage(): ChatMessage {
  return {
    id: uid(),
    role: 'assistant',
    text: welcomeText,
    ts: Date.now(),
  }
}

export const useChatStore = create<ChatState>((set, get) => ({
  open: false,
  sending: false,
  sessionId: null,
  messages: [welcomeMessage()],

  toggle: () => {
    if (get().open) {
      get().close()
      return
    }
    get().openChat()
  },
  openChat: () => {
    if (get().open) return
    const sessionId = newSessionId()
    const currentUser = useAuthStore.getState().user
    set({ open: true, sessionId, messages: [welcomeMessage()] })
    void apiPost('/chat/session', {
      session_id: sessionId,
      user_name: currentUser?.name || '',
      user_email: currentUser?.email || '',
      user_phone: currentUser?.phone || '',
    }).catch(() => undefined)
  },
  close: () => {
    const sessionId = get().sessionId
    if (sessionId) {
      void apiPost('/chat/session/close', { session_id: sessionId }).catch(() => undefined)
    }
    set({ open: false, sending: false, sessionId: null, messages: [welcomeMessage()] })
  },
  reset: () => set({ messages: [welcomeMessage()] }),

  send: async (text, onIntent) => {
    text = text.trim()
    if (!text) return
    if (get().sending) return

    const currentUser = useAuthStore.getState().user
    let sessionId = get().sessionId
    if (!sessionId) {
      sessionId = newSessionId()
      set({ open: true, sessionId, messages: [welcomeMessage()] })
    }
    const history = get().messages.slice(-12).map((m) => ({ role: m.role, text: m.text }))
    const userMsg: ChatMessage = { id: uid(), role: 'user', text, ts: Date.now() }
    set({ sending: true, messages: [...get().messages, userMsg] })

    try {
      const resp = await apiPost<{ text: string; model: string; intent?: string }>('/chat', {
        session_id: sessionId,
        message: text,
        user_name: currentUser?.name || '',
        user_email: currentUser?.email || '',
        user_phone: currentUser?.phone || '',
        history,
      })
      const botMsg: ChatMessage = { id: uid(), role: 'assistant', text: resp.text, ts: Date.now() }
      set({ messages: [...get().messages, botMsg] })
      if (resp.intent) onIntent?.(resp.intent)
    } catch (e) {
      const botMsg: ChatMessage = {
        id: uid(),
        role: 'assistant',
        text: 'Сейчас не получается ответить. Попробуйте ещё раз чуть позже — и мы обязательно поможем.',
        ts: Date.now(),
      }
      set({ messages: [...get().messages, botMsg] })
    } finally {
      set({ sending: false })
    }
  },
}))
