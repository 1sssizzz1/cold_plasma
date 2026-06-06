import { motion, AnimatePresence } from 'framer-motion'
import { useEffect, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useChatStore } from '../../store/chat'

const quickActions = [
  { label: 'Цена', text: 'Сколько стоит процедура холодной плазмы?' },
  { label: 'Противопоказания', text: 'Какие есть противопоказания для холодной плазмы?' },
  { label: 'Как проходит', text: 'Как проходит процедура холодной плазмы?' },
  { label: 'Адрес', text: 'Где вы находитесь и как к вам добраться?' },
  { label: 'Записаться', text: 'Хочу записаться на процедуру холодной плазмы' },
]

export default function ChatWidget() {
  const navigate = useNavigate()
  const open = useChatStore((s) => s.open)
  const toggle = useChatStore((s) => s.toggle)
  const close = useChatStore((s) => s.close)
  const messages = useChatStore((s) => s.messages)
  const sending = useChatStore((s) => s.sending)
  const send = useChatStore((s) => s.send)

  const [text, setText] = useState('')
  const listRef = useRef<HTMLDivElement | null>(null)

  useEffect(() => {
    listRef.current?.scrollTo({ top: listRef.current.scrollHeight, behavior: 'smooth' })
  }, [messages.length, open])

  const openBooking = (intent: string) => {
    if (intent !== 'booking') return
    close()
    navigate('/booking')
  }

  const submit = (value: string) => {
    const prepared = value.trim()
    if (!prepared) return
    void send(prepared, openBooking)
    setText('')
  }

  const finishDialog = () => {
    close()
  }

  return (
    <div className="fixed bottom-4 right-4 z-50">
      <AnimatePresence>
        {open && (
          <motion.div
            initial={{ opacity: 0, y: 12, scale: 0.98 }}
            animate={{ opacity: 1, y: 0, scale: 1 }}
            exit={{ opacity: 0, y: 12, scale: 0.98 }}
            transition={{ duration: 0.2 }}
            className="glass-panel mb-3 w-[92vw] max-w-sm overflow-hidden rounded-2xl shadow-2xl"
          >
            <div className="flex items-center justify-between gap-3 border-b border-white/5 px-4 py-3">
              <div>
                <div className="text-sm font-semibold text-white">AI-консультант</div>
                <div className="text-xs text-zinc-500">Помнит только текущий диалог</div>
              </div>
              <div className="flex gap-2">
                <button
                  type="button"
                  onClick={finishDialog}
                  className="rounded-full border border-white/10 bg-white/5 px-3 py-1 text-xs text-zinc-400 hover:text-white"
                >
                  Завершить
                </button>
                <button
                  type="button"
                  onClick={toggle}
                  className="rounded-full border border-white/10 bg-white/5 px-3 py-1 text-xs text-zinc-400 hover:text-white"
                >
                  Закрыть
                </button>
              </div>
            </div>

            <div ref={listRef} className="max-h-[50vh] space-y-3 overflow-auto px-4 py-4">
              {messages.map((m) => (
                <div key={m.id} className={m.role === 'user' ? 'flex justify-end' : 'flex justify-start'}>
                  <div
                    className={[
                      'max-w-[85%] whitespace-pre-line px-3.5 py-2.5 text-sm leading-snug shadow-sm',
                      m.role === 'user'
                        ? 'rounded-[18px] rounded-br-md bg-black text-white'
                        : 'rounded-[18px] rounded-bl-md border border-black/10 bg-white text-black',
                    ].join(' ')}
                  >
                    {m.text}
                  </div>
                </div>
              ))}
              {sending && (
                <div className="flex justify-start">
                  <div className="rounded-[18px] rounded-bl-md border border-black/10 bg-white px-3.5 py-2.5 text-sm text-zinc-500 shadow-sm">Печатает…</div>
                </div>
              )}
            </div>

            <div className="flex flex-wrap gap-2 border-t border-white/5 px-3 py-3">
              {quickActions.map((action) => (
                <button
                  key={action.label}
                  type="button"
                  disabled={sending}
                  onClick={() => submit(action.text)}
                  className="rounded-full border border-white/10 bg-white/5 px-3 py-1.5 text-xs font-semibold text-zinc-400 transition hover:border-white/30 hover:text-white disabled:opacity-40"
                >
                  {action.label}
                </button>
              ))}
            </div>

            <form
              className="flex gap-2 border-t border-white/5 p-3"
              onSubmit={(e) => {
                e.preventDefault()
                submit(text)
              }}
            >
              <input
                value={text}
                onChange={(e) => setText(e.target.value)}
                placeholder="Спросите про холодную плазму…"
                className="h-10 flex-1 rounded-xl border border-white/10 bg-zinc-950 px-3 text-sm text-white outline-none placeholder:text-zinc-600 focus:border-white/30"
              />
              <button
                disabled={sending || !text.trim()}
                className="h-10 rounded-xl bg-white px-4 text-xs font-semibold uppercase tracking-wider text-black transition hover:bg-zinc-200 disabled:opacity-40"
              >
                Отправить
              </button>
            </form>
          </motion.div>
        )}
      </AnimatePresence>

      <button
        onClick={toggle}
        className="flex h-12 items-center justify-center gap-2 rounded-full border border-white/10 bg-white px-4 text-xs font-semibold uppercase tracking-wider text-black shadow-[0_0_30px_rgba(255,255,255,0.15)] transition hover:bg-zinc-200"
      >
        {open ? 'Чат' : 'Чат с AI'}
        <span className="inline-flex h-2 w-2 rounded-full bg-black/70" />
      </button>
    </div>
  )
}
