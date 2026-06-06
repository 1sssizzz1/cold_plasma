import { Helmet } from 'react-helmet-async'
import { useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import Container from '../components/Container'
import MotionPage from '../components/MotionPage'
import { useAuthStore } from '../store/auth'
import { useProceduresStore } from '../store/procedures'
import { apiGet, apiPost } from '../utils/api'
import type { Booking, Procedure, User } from '../types'

function toISO(local: string) {
  const d = new Date(local)
  if (Number.isNaN(d.getTime())) return null
  return d.toISOString()
}

export default function BookingPage() {
  const token = useAuthStore((s) => s.token)
  const user = useAuthStore((s) => s.user)
  const setAuth = useAuthStore((s) => s.setAuth)

  const fetchAll = useProceduresStore((s) => s.fetchAll)
  const procedures = useProceduresStore((s) => s.items)

  const [procedureId, setProcedureId] = useState<number>(0)
  const [datetime, setDatetime] = useState('')
  const [datetimes, setDatetimes] = useState<string[]>([])
  const [comment, setComment] = useState('')
  const [notifySMS, setNotifySMS] = useState(true)
  const [notifyTelegram, setNotifyTelegram] = useState(true)
  const [okState, setOkState] = useState<{ booking: Booking; balance: number } | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)
  const apiBase = ((import.meta.env.VITE_API_URL as string | undefined) ?? '/api/v1') as string

  useEffect(() => {
    void fetchAll()
  }, [fetchAll])

  useEffect(() => {
    if (!token) return
    void apiGet<User>('/me')
      .then((u) => setAuth(token, u))
      .catch(() => {})
  }, [token, setAuth])

  const selected: Procedure | undefined = useMemo(
    () => procedures.find((p) => p.id === procedureId),
    [procedures, procedureId],
  )

  return (
    <MotionPage>
      <Helmet>
        <title>Запись — Холодная плазма (Северодвинск)</title>
        <meta name="description" content="Онлайн-запись на холодную плазму в Северодвинске: выберите процедуру и время." />
      </Helmet>

      <section className="py-12">
        <Container>
          <div className="border-b border-white/5 pb-6">
            <div className="text-xs font-semibold uppercase tracking-widest text-zinc-500">Онлайн запись</div>
            <div className="mt-2 text-3xl font-semibold tracking-tight text-white sm:text-4xl">Запись</div>
            <div className="mt-3 max-w-2xl text-sm text-zinc-400">Заполните форму — мы подтвердим запись и подберём удобное время.</div>
          </div>

          {!token ? (
            <div className="glass-panel mt-8 rounded-3xl p-6 shadow-xl">
              <div className="text-lg font-semibold text-white">Нужен личный кабинет</div>
              <div className="mt-2 text-sm leading-relaxed text-zinc-400">
                Войдите или зарегистрируйтесь — так вы сможете видеть свои записи и бонусы.
              </div>
              <Link
                to="/account"
                className="mt-5 inline-flex h-11 items-center justify-center rounded-2xl bg-white px-5 text-xs font-semibold uppercase tracking-wider text-black transition hover:bg-zinc-200"
              >
                Перейти в кабинет
              </Link>
            </div>
          ) : !user?.phone_verified ? (
            <div className="mt-8 rounded-3xl border border-amber-300 bg-amber-50 p-6 shadow-xl">
              <div className="text-lg font-semibold text-amber-950">Нужно подтвердить телефон</div>
              <div className="mt-2 text-sm leading-relaxed text-amber-900">
                Онлайн-запись доступна только аккаунтам с подтверждённым российским номером телефона.
              </div>
              <Link
                to="/account"
                className="mt-5 inline-flex h-11 items-center justify-center rounded-2xl bg-black px-5 text-xs font-semibold uppercase tracking-wider text-white shadow-sm transition hover:bg-zinc-800"
              >
                Подтвердить в кабинете
              </Link>
            </div>
          ) : (
            <div className="mt-8 grid gap-6 md:grid-cols-2">
              <div className="glass-panel rounded-3xl p-6 shadow-xl">
                <div className="text-xs font-semibold uppercase tracking-widest text-zinc-500">Данные записи</div>

                {error && (
                  <div className="mt-4 rounded-2xl border border-red-500/20 bg-red-500/10 p-3 text-sm text-red-200">
                    {error}
                  </div>
                )}
                {okState && (
                  <div className="mt-4 rounded-2xl border border-emerald-500/30 bg-emerald-500/10 p-4">
                    <div className="flex items-start gap-3">
                      <div className="text-2xl">✅</div>
                      <div className="flex-1">
                        <div className="font-semibold text-emerald-100">Заявка успешно создана!</div>
                        <div className="mt-1 text-sm text-emerald-200">
                          Мы получили вашу заявку и скоро подтвердим подходящее время.
                        </div>
                        <div className="mt-3 flex items-center gap-2 text-xs text-emerald-300">
                          <span>💰 Бонусный баланс:</span>
                          <span className="font-semibold text-emerald-100">{okState.balance} баллов</span>
                        </div>
                      </div>
                    </div>
                  </div>
                )}

                <div className="mt-4">
                  <label className="text-xs font-semibold uppercase tracking-wider text-zinc-500">Процедура</label>
                  <select
                    className="mt-2 h-11 w-full rounded-2xl border border-white/10 bg-zinc-950 px-3 text-sm text-white outline-none focus:border-white/30"
                    value={procedureId}
                    onChange={(e) => setProcedureId(Number(e.target.value))}
                  >
                    <option value={0}>Выберите процедуру</option>
                    {procedures.map((p) => (
                      <option key={p.id} value={p.id}>
                        {p.title} — {p.price} ₽
                      </option>
                    ))}
                  </select>
                </div>

                <div className="mt-4">
                  <label className="text-xs font-semibold uppercase tracking-wider text-zinc-500">Варианты даты и времени</label>
                  <div className="mt-1 flex gap-2">
                    <input
                      type="datetime-local"
                      className="h-11 min-w-0 flex-1 rounded-2xl border border-white/10 bg-zinc-950 px-3 text-sm text-white outline-none focus:border-white/30"
                      value={datetime}
                      onChange={(e) => setDatetime(e.target.value)}
                    />
                    <button
                      type="button"
                      className="h-11 rounded-2xl border border-white/10 bg-white/5 px-4 text-xs font-semibold uppercase tracking-wider text-white transition hover:border-white/30 hover:bg-white/10"
                      onClick={() => {
                        if (!datetime) return
                        setDatetimes((items) => (items.includes(datetime) ? items : [...items, datetime]))
                        setDatetime('')
                      }}
                    >
                      Добавить
                    </button>
                  </div>
                  <div className="mt-2 text-xs text-zinc-500">Можно выбрать несколько удобных вариантов.</div>
                  <div className="mt-3 space-y-2">
                    {datetimes.map((item) => (
                      <div key={item} className="flex items-center justify-between rounded-2xl border border-white/10 bg-white/[0.03] px-3 py-2 text-sm text-zinc-300">
                        <span>{new Date(item).toLocaleString('ru-RU')}</span>
                        <button
                          type="button"
                          className="text-xs font-semibold uppercase tracking-wider text-zinc-500 hover:text-white"
                          onClick={() => setDatetimes((items) => items.filter((x) => x !== item))}
                        >
                          Убрать
                        </button>
                      </div>
                    ))}
                  </div>
                </div>

                <div className="mt-4">
                  <label className="text-xs font-semibold uppercase tracking-wider text-zinc-500">Сообщение (необязательно)</label>
                  <textarea
                    className="mt-2 min-h-20 w-full rounded-2xl border border-white/10 bg-zinc-950 px-3 py-2 text-sm text-white outline-none placeholder:text-zinc-600 focus:border-white/30"
                    value={comment}
                    onChange={(e) => setComment(e.target.value)}
                    placeholder="Например: чувствительная кожа, хочется мягкий уход…"
                  />
                </div>

                <div className="mt-4 space-y-2 rounded-2xl border border-white/10 bg-white/[0.03] p-4 text-sm text-zinc-300">
                  <label className="flex items-center gap-3">
                    <input type="checkbox" checked={notifySMS} onChange={(e) => setNotifySMS(e.target.checked)} />
                    <span>Сообщить о записи через SMS</span>
                  </label>
                  <label className="flex items-center gap-3">
                    <input
                      type="checkbox"
                      checked={notifyTelegram}
                      onChange={(e) => setNotifyTelegram(e.target.checked)}
                    />
                    <span>Сообщить о записи через Telegram</span>
                  </label>
                </div>

                <button
                  disabled={loading}
                  className="mt-6 h-11 w-full rounded-2xl bg-white text-xs font-semibold uppercase tracking-wider text-black transition hover:bg-zinc-200 disabled:opacity-40"
                  onClick={async () => {
                    setError(null)
                    setOkState(null)
                    setLoading(true)
                    try {
                      if (!procedureId) throw new Error('Выберите процедуру')
                      const dtIsoItems = datetimes.map(toISO).filter(Boolean)
                      if (!dtIsoItems.length) throw new Error('Добавьте хотя бы одну дату и время')

                      const resp = await apiPost<{ booking: Booking; bonus_balance: number }>('/bookings', {
                        procedure_id: procedureId,
                        datetimes: dtIsoItems,
                        comment,
                        bonus_used: 0,
                        notify_sms: notifySMS,
                        notify_telegram: notifyTelegram,
                      })
                      setOkState({ booking: resp.booking, balance: resp.bonus_balance })
                      setDatetimes([])
                      setComment('')
                    } catch (e) {
                      setError((e as Error).message)
                    } finally {
                      setLoading(false)
                    }
                  }}
                >
                  Создать запись
                </button>
              </div>

              <div className="glass-panel rounded-3xl p-6 shadow-xl">
                <div className="text-xs font-semibold uppercase tracking-widest text-zinc-500">Подсказка</div>
                <div className="mt-3 text-sm leading-relaxed text-zinc-400">
                  {user?.name ? (
                    <>
                      {user.name}, если сомневаетесь — откройте чат внизу справа. Ответим коротко и по делу и предложим
                      удобное время.
                    </>
                  ) : (
                    <>Если сомневаетесь — откройте чат внизу справа. Ответим коротко и по делу и предложим удобное время.</>
                  )}
                </div>

                <div className="mt-6 rounded-2xl border border-white/10 bg-white/[0.03] p-4 text-sm text-zinc-400">
                  <div className="font-semibold text-white">Памятка по уходу (PDF)</div>
                  <div className="mt-1">
                    Можно скачать заранее — она пригодится после процедуры.
                  </div>
                  <a
                    className="mt-3 inline-flex h-10 items-center justify-center rounded-2xl border border-white/10 bg-white/5 px-4 text-xs font-semibold uppercase tracking-wider text-white transition hover:border-white/30 hover:bg-white/10"
                    href={apiBase + '/pdf/care-memo'}
                    target="_blank"
                    rel="noreferrer"
                  >
                    Открыть PDF
                  </a>
                </div>

                {selected && (
                  <div className="mt-6 rounded-2xl border border-white/10 bg-white/[0.03] p-4">
                    <div className="text-xs font-semibold uppercase tracking-wider text-zinc-500">Вы выбрали</div>
                    <div className="mt-1 text-lg font-semibold text-white">{selected.title}</div>
                    <div className="mt-2 text-sm leading-relaxed text-zinc-400">{selected.description}</div>
                  </div>
                )}
              </div>
            </div>
          )}
        </Container>
      </section>
    </MotionPage>
  )
}
