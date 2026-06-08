import { Helmet } from 'react-helmet-async'
import { useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import Container from '../components/Container'
import MotionPage from '../components/MotionPage'
import { useAuthStore } from '../store/auth'
import { useProceduresStore } from '../store/procedures'
import { apiGet, apiPost } from '../utils/api'
import type { Booking, DaySlots, Procedure, Slot, User } from '../types'

const MONTHS = [
  'Январь', 'Февраль', 'Март', 'Апрель', 'Май', 'Июнь',
  'Июль', 'Август', 'Сентябрь', 'Октябрь', 'Ноябрь', 'Декабрь',
]
const WEEKDAYS = ['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс']

function ymd(date: Date) {
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}`
}

function formatDayLabel(date: string) {
  const d = new Date(date + 'T00:00:00')
  if (Number.isNaN(d.getTime())) return date
  return d.toLocaleDateString('ru-RU', { weekday: 'short', day: 'numeric', month: 'short' })
}

function formatSlotTime(iso: string) {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  return d.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })
}

// Дни месяца, выровненные по неделям (Пн первый), с хвостами соседних месяцев.
function buildMonthGrid(year: number, month: number) {
  const first = new Date(year, month, 1)
  const startOffset = (first.getDay() + 6) % 7 // Пн = 0
  const gridStart = new Date(year, month, 1 - startOffset)
  const cells: Date[] = []
  for (let i = 0; i < 42; i += 1) {
    cells.push(new Date(gridStart.getFullYear(), gridStart.getMonth(), gridStart.getDate() + i))
  }
  return cells
}

export default function BookingPage() {
  const token = useAuthStore((s) => s.token)
  const user = useAuthStore((s) => s.user)
  const setAuth = useAuthStore((s) => s.setAuth)

  const fetchAll = useProceduresStore((s) => s.fetchAll)
  const procedures = useProceduresStore((s) => s.items)

  const now = useMemo(() => new Date(), [])
  const todayStart = useMemo(() => {
    const d = new Date()
    d.setHours(0, 0, 0, 0)
    return d
  }, [])

  const [procedureId, setProcedureId] = useState<number>(0)
  const [viewYear, setViewYear] = useState(now.getFullYear())
  const [viewMonth, setViewMonth] = useState(now.getMonth())
  const [monthSlots, setMonthSlots] = useState<DaySlots[]>([])
  const [reloadKey, setReloadKey] = useState(0)
  const [loadingSlots, setLoadingSlots] = useState(false)
  const [slotsError, setSlotsError] = useState<string | null>(null)
  const [selectedDate, setSelectedDate] = useState<string>('')
  const [selectedSlot, setSelectedSlot] = useState<Slot | null>(null)
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

  // Загружаем свободные окна для отображаемого месяца.
  useEffect(() => {
    if (!procedureId || !token || !user?.phone_verified) {
      setMonthSlots([])
      return
    }
    const from = new Date(viewYear, viewMonth, 1, 0, 0, 0)
    const to = new Date(viewYear, viewMonth + 1, 1, 0, 0, 0)
    setLoadingSlots(true)
    setSlotsError(null)
    void apiGet<DaySlots[]>(
      `/bookings/availability?procedure_id=${procedureId}&from=${encodeURIComponent(from.toISOString())}&to=${encodeURIComponent(to.toISOString())}`,
    )
      .then((days) => setMonthSlots(Array.isArray(days) ? days : []))
      .catch((e) => setSlotsError((e as Error).message))
      .finally(() => setLoadingSlots(false))
  }, [procedureId, viewYear, viewMonth, token, user?.phone_verified, reloadKey])

  const slotsByDate = useMemo(() => {
    const map = new Map<string, Slot[]>()
    for (const day of monthSlots) {
      if (day.slots.length) map.set(day.date, day.slots)
    }
    return map
  }, [monthSlots])

  const cells = useMemo(() => buildMonthGrid(viewYear, viewMonth), [viewYear, viewMonth])
  const daySlots = useMemo(() => slotsByDate.get(selectedDate) ?? [], [slotsByDate, selectedDate])

  const canGoPrev = viewYear > now.getFullYear() || (viewYear === now.getFullYear() && viewMonth > now.getMonth())
  const goPrev = () => {
    if (!canGoPrev) return
    setSelectedDate('')
    setSelectedSlot(null)
    if (viewMonth === 0) {
      setViewYear((y) => y - 1)
      setViewMonth(11)
    } else {
      setViewMonth((m) => m - 1)
    }
  }
  const goNext = () => {
    setSelectedDate('')
    setSelectedSlot(null)
    if (viewMonth === 11) {
      setViewYear((y) => y + 1)
      setViewMonth(0)
    } else {
      setViewMonth((m) => m + 1)
    }
  }

  const onProcedureChange = (id: number) => {
    setProcedureId(id)
    setViewYear(now.getFullYear())
    setViewMonth(now.getMonth())
    setSelectedDate('')
    setSelectedSlot(null)
  }

  const submit = async () => {
    setError(null)
    setOkState(null)
    setLoading(true)
    try {
      if (!procedureId) throw new Error('Выберите процедуру')
      if (!selectedSlot) throw new Error('Выберите свободное окно')

      const resp = await apiPost<{ booking: Booking; bonus_balance: number }>('/bookings', {
        procedure_id: procedureId,
        datetime: selectedSlot.start_at,
        comment,
        bonus_used: 0,
        notify_sms: notifySMS,
        notify_telegram: notifyTelegram,
      })
      setOkState({ booking: resp.booking, balance: resp.bonus_balance })
      setComment('')
      setSelectedSlot(null)
      // Обновляем доступность — выбранное окно могло занять кто-то ещё.
      setReloadKey((k) => k + 1)
    } catch (e) {
      setError((e as Error).message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <MotionPage>
      <Helmet>
        <title>Запись — Холодная плазма (Северодвинск)</title>
        <meta name="description" content="Онлайн-запись на холодную плазму в Северодвинске: выберите процедуру и свободное время." />
      </Helmet>

      <section className="py-12">
        <Container>
          <div className="border-b border-white/5 pb-6">
            <div className="text-xs font-semibold uppercase tracking-widest text-zinc-500">Онлайн запись</div>
            <div className="mt-2 text-3xl font-semibold tracking-tight text-white sm:text-4xl">Запись</div>
            <div className="mt-3 max-w-2xl text-sm text-zinc-400">Выберите процедуру, дату и свободное окно — мы подтвердим запись.</div>
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
                          Мы получили вашу заявку и скоро подтвердим запись.
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
                    onChange={(e) => onProcedureChange(Number(e.target.value))}
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
                  <label className="text-xs font-semibold uppercase tracking-wider text-zinc-500">Дата</label>
                  {!procedureId ? (
                    <div className="mt-2 rounded-2xl border border-white/10 bg-white/[0.03] p-3 text-sm text-zinc-500">
                      Сначала выберите процедуру.
                    </div>
                  ) : (
                    <div className="mt-2 rounded-2xl border border-white/10 bg-white/[0.02] p-3">
                      <div className="flex items-center justify-between">
                        <button
                          type="button"
                          onClick={goPrev}
                          disabled={!canGoPrev}
                          aria-label="Предыдущий месяц"
                          className="flex h-9 w-9 items-center justify-center rounded-xl border border-white/10 text-lg text-white transition hover:border-white/30 disabled:cursor-not-allowed disabled:opacity-30"
                        >
                          ‹
                        </button>
                        <div className="text-sm font-semibold text-white">{MONTHS[viewMonth]} {viewYear}</div>
                        <button
                          type="button"
                          onClick={goNext}
                          aria-label="Следующий месяц"
                          className="flex h-9 w-9 items-center justify-center rounded-xl border border-white/10 text-lg text-white transition hover:border-white/30"
                        >
                          ›
                        </button>
                      </div>

                      <div className="mt-3 grid grid-cols-7 gap-1 text-center text-[11px] font-semibold uppercase text-zinc-600">
                        {WEEKDAYS.map((d) => <div key={d}>{d}</div>)}
                      </div>
                      <div className="mt-1 grid grid-cols-7 gap-1">
                        {cells.map((cell) => {
                          const key = ymd(cell)
                          const inMonth = cell.getMonth() === viewMonth && cell.getFullYear() === viewYear
                          const isPast = cell < todayStart
                          const hasSlots = slotsByDate.has(key)
                          const isToday = key === ymd(now)
                          const isSelected = key === selectedDate
                          const disabled = !inMonth || isPast || !hasSlots
                          return (
                            <button
                              key={key}
                              type="button"
                              disabled={disabled}
                              onClick={() => {
                                setSelectedDate(key)
                                setSelectedSlot(null)
                              }}
                              className={[
                                'relative flex aspect-square items-center justify-center rounded-xl text-sm transition',
                                isSelected
                                  ? 'bg-white font-semibold text-black'
                                  : disabled
                                    ? 'text-zinc-700'
                                    : 'border border-white/10 bg-white/[0.04] text-white hover:border-white/30',
                                !inMonth ? 'opacity-30' : '',
                              ].join(' ')}
                            >
                              {cell.getDate()}
                              {hasSlots && !isSelected && (
                                <span className="absolute bottom-1 h-1 w-1 rounded-full bg-emerald-400" />
                              )}
                              {isToday && !isSelected && (
                                <span className="absolute inset-x-2 top-1 h-px bg-emerald-400/60" />
                              )}
                            </button>
                          )
                        })}
                      </div>

                      <div className="mt-2 text-xs text-zinc-500">
                        {loadingSlots
                          ? 'Загружаем свободные окна…'
                          : slotsError
                            ? slotsError
                            : slotsByDate.size
                              ? 'Зелёная точка — есть свободные окна.'
                              : 'В этом месяце свободных окон нет — переключите месяц стрелкой ›.'}
                      </div>
                    </div>
                  )}
                </div>

                {procedureId > 0 && selectedDate && (
                  <div className="mt-4">
                    <label className="text-xs font-semibold uppercase tracking-wider text-zinc-500">
                      Свободное окно · {formatDayLabel(selectedDate)}
                    </label>
                    {!daySlots.length ? (
                      <div className="mt-2 rounded-2xl border border-white/10 bg-white/[0.03] p-3 text-sm text-zinc-400">
                        На эту дату окон не осталось.
                      </div>
                    ) : (
                      <div className="mt-2 grid grid-cols-3 gap-2 sm:grid-cols-4">
                        {daySlots.map((slot) => (
                          <button
                            key={slot.start_at}
                            type="button"
                            onClick={() => setSelectedSlot(slot)}
                            className={[
                              'rounded-2xl border px-2 py-2 text-sm font-semibold transition',
                              selectedSlot?.start_at === slot.start_at
                                ? 'border-white bg-white text-black'
                                : 'border-white/10 bg-white/[0.03] text-zinc-200 hover:border-white/30',
                            ].join(' ')}
                          >
                            {formatSlotTime(slot.start_at)}
                          </button>
                        ))}
                      </div>
                    )}
                  </div>
                )}

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
                  disabled={loading || !selectedSlot}
                  className="mt-6 h-11 w-full rounded-2xl bg-white text-xs font-semibold uppercase tracking-wider text-black transition hover:bg-zinc-200 disabled:opacity-40"
                  onClick={() => void submit()}
                >
                  {selectedSlot
                    ? `Записаться на ${formatDayLabel(selectedDate)}, ${formatSlotTime(selectedSlot.start_at)}`
                    : 'Создать запись'}
                </button>
              </div>

              <div className="glass-panel rounded-3xl p-6 shadow-xl">
                <div className="text-xs font-semibold uppercase tracking-widest text-zinc-500">Подсказка</div>
                <div className="mt-3 text-sm leading-relaxed text-zinc-400">
                  {user?.name ? (
                    <>
                      {user.name}, выберите процедуру — мы покажем свободные окна ежедневно с 12:00 до 19:00.
                      Если сомневаетесь, откройте чат внизу справа.
                    </>
                  ) : (
                    <>Выберите процедуру — мы покажем свободные окна ежедневно с 12:00 до 19:00.</>
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
                    <div className="mt-1 text-xs text-zinc-500">Длительность окна: {selected.duration_str || `${selected.duration_mins} мин`}</div>
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
