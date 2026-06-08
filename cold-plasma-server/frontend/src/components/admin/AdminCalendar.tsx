import { useEffect, useMemo, useState } from 'react'
import { apiDelete, apiGet, apiPost } from '../../utils/api'
import type { AdminNote, CalendarBooking, CalendarData } from '../../types'

const MONTHS = [
  'Январь', 'Февраль', 'Март', 'Апрель', 'Май', 'Июнь',
  'Июль', 'Август', 'Сентябрь', 'Октябрь', 'Ноябрь', 'Декабрь',
]
const WEEKDAYS = ['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс']
const WORK_START_HOUR = 12
const WORK_END_HOUR = 19

function ymd(date: Date) {
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}`
}

function timeLabel(iso: string) {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return ''
  return d.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })
}

// Возвращает дни месяца, выровненные по неделям (Пн первый), с хвостами соседних месяцев.
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

type DayBucket = {
  bookings: CalendarBooking[]
  notes: AdminNote[]
}

const statusLabels: Record<string, string> = {
  new: 'Заявка',
  confirmed: 'Подтв.',
  completed: 'Проведена',
}

export default function AdminCalendar() {
  const today = new Date()
  const [year, setYear] = useState(today.getFullYear())
  const [month, setMonth] = useState(today.getMonth())
  const [data, setData] = useState<CalendarData | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [info, setInfo] = useState<string | null>(null)
  const [selectedDate, setSelectedDate] = useState<string>(ymd(today))

  // Форма заметки
  const [noteDate, setNoteDate] = useState<string>(ymd(today))
  const [noteStart, setNoteStart] = useState('12:00')
  const [noteEnd, setNoteEnd] = useState('13:00')
  const [noteTitle, setNoteTitle] = useState('')
  const [savingNote, setSavingNote] = useState(false)

  const load = async () => {
    setLoading(true)
    setError(null)
    try {
      const from = new Date(year, month, 1, 0, 0, 0)
      const to = new Date(year, month + 1, 1, 0, 0, 0)
      const resp = await apiGet<CalendarData>(
        `/admin/calendar?from=${encodeURIComponent(from.toISOString())}&to=${encodeURIComponent(to.toISOString())}`,
      )
      setData(resp)
    } catch (e) {
      setError((e as Error).message)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void load()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [year, month])

  const buckets = useMemo(() => {
    const map = new Map<string, DayBucket>()
    if (data) {
      for (const b of data.bookings) {
        const key = ymd(new Date(b.start_at))
        const bucket = map.get(key) ?? { bookings: [], notes: [] }
        bucket.bookings.push(b)
        map.set(key, bucket)
      }
      for (const n of data.notes) {
        const key = ymd(new Date(n.start_at))
        const bucket = map.get(key) ?? { bookings: [], notes: [] }
        bucket.notes.push(n)
        map.set(key, bucket)
      }
    }
    return map
  }, [data])

  const cells = useMemo(() => buildMonthGrid(year, month), [year, month])
  const selectedBucket = buckets.get(selectedDate) ?? { bookings: [], notes: [] }

  const goPrev = () => {
    if (month === 0) {
      setYear((y) => y - 1)
      setMonth(11)
    } else {
      setMonth((m) => m - 1)
    }
  }
  const goNext = () => {
    if (month === 11) {
      setYear((y) => y + 1)
      setMonth(0)
    } else {
      setMonth((m) => m + 1)
    }
  }

  const createNote = async () => {
    setError(null)
    setInfo(null)
    setSavingNote(true)
    try {
      const startAt = new Date(`${noteDate}T${noteStart}:00`)
      const endAt = new Date(`${noteDate}T${noteEnd}:00`)
      if (Number.isNaN(startAt.getTime()) || Number.isNaN(endAt.getTime())) {
        throw new Error('Укажите корректные дату и время')
      }
      if (endAt <= startAt) throw new Error('Конец должен быть позже начала')
      await apiPost<{ note: AdminNote }>('/admin/notes', {
        start_at: startAt.toISOString(),
        end_at: endAt.toISOString(),
        title: noteTitle,
      })
      setInfo('Заметка добавлена — это время закрыто для записи.')
      setNoteTitle('')
      await load()
    } catch (e) {
      setError((e as Error).message)
    } finally {
      setSavingNote(false)
    }
  }

  const deleteNote = async (note: AdminNote) => {
    if (!window.confirm('Удалить заметку? Это время снова станет доступным для записи.')) return
    setError(null)
    setInfo(null)
    try {
      await apiDelete<{ ok: boolean }>(`/admin/notes/${note.id}`)
      setInfo('Заметка удалена.')
      await load()
    } catch (e) {
      setError((e as Error).message)
    }
  }

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex items-center gap-2">
          <button type="button" onClick={goPrev} className="h-10 w-10 rounded-2xl border border-black/15 text-lg font-semibold hover:border-black/35">‹</button>
          <div className="min-w-44 text-center text-lg font-semibold">{MONTHS[month]} {year}</div>
          <button type="button" onClick={goNext} className="h-10 w-10 rounded-2xl border border-black/15 text-lg font-semibold hover:border-black/35">›</button>
        </div>
        <button
          type="button"
          onClick={() => { setYear(today.getFullYear()); setMonth(today.getMonth()); setSelectedDate(ymd(today)) }}
          className="h-10 rounded-2xl border border-black/15 px-4 text-sm font-semibold hover:border-black/35"
        >
          Сегодня
        </button>
      </div>

      {error && <div className="rounded-2xl border border-red-200 bg-red-50 p-4 text-sm text-red-700">{error}</div>}
      {info && <div className="rounded-2xl border border-emerald-200 bg-emerald-50 p-4 text-sm text-emerald-800">{info}</div>}

      <div className="grid gap-4 lg:grid-cols-[minmax(0,1fr)_360px]">
        <div className="rounded-3xl border border-black/10 bg-white p-4">
          <div className="mb-2 grid grid-cols-7 gap-1 text-center text-xs font-semibold text-black/45">
            {WEEKDAYS.map((d) => <div key={d}>{d}</div>)}
          </div>
          <div className="grid grid-cols-7 gap-1">
            {cells.map((cell) => {
              const key = ymd(cell)
              const inMonth = cell.getMonth() === month
              const bucket = buckets.get(key)
              const isToday = key === ymd(today)
              const isSelected = key === selectedDate
              const bookingCount = bucket?.bookings.length ?? 0
              const noteCount = bucket?.notes.length ?? 0
              return (
                <button
                  key={key}
                  type="button"
                  onClick={() => { setSelectedDate(key); setNoteDate(key) }}
                  className={[
                    'flex min-h-16 flex-col rounded-2xl border p-2 text-left transition',
                    inMonth ? 'bg-white' : 'bg-black/[0.02] text-black/35',
                    isSelected ? 'border-black ring-1 ring-black' : 'border-black/10 hover:border-black/25',
                  ].join(' ')}
                >
                  <span className={['text-xs font-semibold', isToday ? 'text-emerald-600' : ''].join(' ')}>
                    {cell.getDate()}
                  </span>
                  <span className="mt-auto flex flex-wrap gap-1">
                    {bookingCount > 0 && (
                      <span className="rounded-full bg-black px-1.5 py-0.5 text-[10px] font-semibold text-white">{bookingCount}</span>
                    )}
                    {noteCount > 0 && (
                      <span className="rounded-full bg-amber-500 px-1.5 py-0.5 text-[10px] font-semibold text-white">{noteCount}</span>
                    )}
                  </span>
                </button>
              )
            })}
          </div>
          <div className="mt-3 flex items-center gap-4 text-xs text-black/55">
            <span className="flex items-center gap-1"><span className="h-3 w-3 rounded-full bg-black" /> Записи</span>
            <span className="flex items-center gap-1"><span className="h-3 w-3 rounded-full bg-amber-500" /> Заметки</span>
            {loading && <span className="text-black/40">Обновляем…</span>}
          </div>
        </div>

        <div className="space-y-4">
          <div className="rounded-3xl border border-black/10 bg-white p-5">
            <div className="text-sm font-semibold">
              {new Date(selectedDate + 'T00:00:00').toLocaleDateString('ru-RU', { weekday: 'long', day: 'numeric', month: 'long' })}
            </div>

            <div className="mt-3 space-y-2">
              {!selectedBucket.bookings.length && !selectedBucket.notes.length && (
                <div className="rounded-2xl border border-black/10 p-3 text-sm text-black/50">На этот день ничего нет.</div>
              )}
              {selectedBucket.bookings
                .slice()
                .sort((a, b) => a.start_at.localeCompare(b.start_at))
                .map((b) => (
                  <div key={`b-${b.id}`} className="rounded-2xl border border-black/10 p-3 text-sm">
                    <div className="flex items-center justify-between gap-2">
                      <span className="font-semibold">{timeLabel(b.start_at)}–{timeLabel(b.end_at)}</span>
                      <span className="rounded-full border border-black/10 px-2 py-0.5 text-xs text-black/55">{statusLabels[b.status] ?? b.status}</span>
                    </div>
                    <div className="mt-1 text-black/70">{b.procedure_title}</div>
                    <div className="text-xs text-black/50">{b.user_name || 'Клиент'} · {b.user_phone || 'без телефона'}</div>
                  </div>
                ))}
              {selectedBucket.notes
                .slice()
                .sort((a, b) => a.start_at.localeCompare(b.start_at))
                .map((n) => (
                  <div key={`n-${n.id}`} className="rounded-2xl border border-amber-300 bg-amber-50 p-3 text-sm">
                    <div className="flex items-center justify-between gap-2">
                      <span className="font-semibold text-amber-900">{timeLabel(n.start_at)}–{timeLabel(n.end_at)}</span>
                      <button
                        type="button"
                        onClick={() => void deleteNote(n)}
                        className="text-xs font-semibold text-red-700 hover:underline"
                      >
                        Удалить
                      </button>
                    </div>
                    <div className="mt-1 text-amber-900">{n.title || 'Заметка (время закрыто)'}</div>
                  </div>
                ))}
            </div>
          </div>

          <div className="rounded-3xl border border-black/10 bg-white p-5">
            <div className="text-sm font-semibold">Закрыть время (заметка)</div>
            <div className="mt-3 space-y-3">
              <label className="block">
                <span className="text-xs font-semibold text-black/60">Дата</span>
                <input
                  type="date"
                  className="mt-1 h-10 w-full rounded-2xl border border-black/10 px-3 text-sm outline-none focus:border-black/30"
                  value={noteDate}
                  onChange={(e) => setNoteDate(e.target.value)}
                />
              </label>
              <div className="grid grid-cols-2 gap-3">
                <label className="block">
                  <span className="text-xs font-semibold text-black/60">С</span>
                  <input
                    type="time"
                    min={`${String(WORK_START_HOUR).padStart(2, '0')}:00`}
                    max={`${String(WORK_END_HOUR).padStart(2, '0')}:00`}
                    className="mt-1 h-10 w-full rounded-2xl border border-black/10 px-3 text-sm outline-none focus:border-black/30"
                    value={noteStart}
                    onChange={(e) => setNoteStart(e.target.value)}
                  />
                </label>
                <label className="block">
                  <span className="text-xs font-semibold text-black/60">До</span>
                  <input
                    type="time"
                    min={`${String(WORK_START_HOUR).padStart(2, '0')}:00`}
                    max={`${String(WORK_END_HOUR).padStart(2, '0')}:00`}
                    className="mt-1 h-10 w-full rounded-2xl border border-black/10 px-3 text-sm outline-none focus:border-black/30"
                    value={noteEnd}
                    onChange={(e) => setNoteEnd(e.target.value)}
                  />
                </label>
              </div>
              <label className="block">
                <span className="text-xs font-semibold text-black/60">Комментарий</span>
                <input
                  className="mt-1 h-10 w-full rounded-2xl border border-black/10 px-3 text-sm outline-none focus:border-black/30"
                  value={noteTitle}
                  onChange={(e) => setNoteTitle(e.target.value)}
                  placeholder="Например: перерыв, личное время"
                />
              </label>
              <button
                type="button"
                disabled={savingNote}
                onClick={() => void createNote()}
                className="h-11 w-full rounded-2xl bg-black text-sm font-semibold text-white disabled:opacity-40"
              >
                {savingNote ? 'Сохраняем…' : 'Закрыть это время'}
              </button>
              <div className="text-xs text-black/45">Рабочие окна — ежедневно с 12:00 до 19:00. Заметка блокирует выбранный интервал для онлайн-записи.</div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
