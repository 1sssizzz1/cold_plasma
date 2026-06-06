import { useEffect, useMemo, useState } from 'react'
import { Helmet } from 'react-helmet-async'
import Container from '../components/Container'
import MotionPage from '../components/MotionPage'
import { apiDelete, apiGet, apiPost, apiPut, apiUpload } from '../utils/api'
import { useAuthStore } from '../store/auth'
import type { BeforeAfterResult, Procedure, ProcedureReview, User } from '../types'

type AdminSectionId = 'bookings' | 'procedures' | 'beforeAfter' | 'master' | 'ai' | 'bonus' | 'notifications'
type BookingTabId = 'requests' | 'active' | 'completed'

type NotificationSettings = {
  notify_telegram: boolean
  notify_sms: boolean
  admin_sms_phone: string
}

type AdminBooking = {
  id: number
  user_id: number
  user_name: string
  user_email: string
  user_phone: string
  procedure_id: number
  procedure_title: string
  datetime: string
  requested_datetimes: string[]
  comment: string
  status: string
  notify_sms: boolean
  notify_telegram: boolean
  created_at: string
}

type ChatSession = {
  id: string
  title: string
  user_id: number | null
  name: string
  email: string
  phone: string
  last_at: string
  messages: Array<{
    role: 'user' | 'assistant'
    text: string
    model?: string
    intent?: string
    created_at: string
  }>
}

type UploadResult = {
  url: string
  mime: string
  size: number
}

type MasterProfile = {
  name: string
  title: string
  bio: string
  photo_url: string
  certificates: string[]
  gallery: string[]
}

const adminSections: Array<{ id: AdminSectionId; label: string; description: string }> = [
  { id: 'bookings', label: 'Записи', description: 'Заявки, активные и завершённые записи' },
  { id: 'procedures', label: 'Процедуры', description: 'Каталог, медиа и отзывы' },
  { id: 'beforeAfter', label: 'До/После', description: 'Результаты процедур, превью и публикация' },
  { id: 'master', label: 'Мастер', description: 'Биография, фото и сертификаты' },
  { id: 'ai', label: 'AI-чаты', description: 'Диалоги по сессиям' },
  { id: 'bonus', label: 'Бонусы', description: 'Поиск и списание по телефону' },
  { id: 'notifications', label: 'Уведомления', description: 'Telegram и SMS админу' },
]

const bookingTabs: Array<{ id: BookingTabId; label: string }> = [
  { id: 'requests', label: 'Заявки на запись' },
  { id: 'active', label: 'Активные записи' },
  { id: 'completed', label: 'Завершённые записи' },
]

function formatDate(value: string) {
  if (!value) return '—'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '—'
  return date.toLocaleString('ru-RU')
}

function toDateTimeLocal(value: string) {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return ''
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`
}

function fromDateTimeLocal(value: string) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return ''
  return date.toISOString()
}

const emptyProcedure: Procedure = {
  id: 0,
  title: '',
  description: '',
  duration_mins: 0,
  price: 0,
  bonus_earned: 0,
  category: '',
  image_url: '',
  video_url: '',
  services: '',
  duration_str: '',
  popular: false,
  is_active: true,
}

const emptyBeforeAfter: BeforeAfterResult = {
  id: 0,
  procedure_id: null,
  procedure: '',
  title: '',
  description: '',
  before_url: '',
  after_url: '',
  is_featured: false,
  sort_order: 0,
  is_active: true,
  created_at: '',
  updated_at: '',
}

export default function AdminPage() {
  const hydrated = useAuthStore((s) => s.hydrated)
  const [section, setSection] = useState<AdminSectionId>('bookings')
  const [bookingTab, setBookingTab] = useState<BookingTabId>('requests')
  const [requests, setRequests] = useState<AdminBooking[]>([])
  const [activeBookings, setActiveBookings] = useState<AdminBooking[]>([])
  const [completedBookings, setCompletedBookings] = useState<AdminBooking[]>([])
  const [procedures, setProcedures] = useState<Procedure[]>([])
  const [procedureReviews, setProcedureReviews] = useState<ProcedureReview[]>([])
  const [procedureForm, setProcedureForm] = useState<Procedure>(emptyProcedure)
  const [beforeAfterResults, setBeforeAfterResults] = useState<BeforeAfterResult[]>([])
  const [beforeAfterForm, setBeforeAfterForm] = useState<BeforeAfterResult>(emptyBeforeAfter)
  const [masterProfile, setMasterProfile] = useState<MasterProfile>({
    name: '',
    title: '',
    bio: '',
    photo_url: '',
    certificates: [],
    gallery: [],
  })
  const [chatSessions, setChatSessions] = useState<ChatSession[]>([])
  const [openedChatId, setOpenedChatId] = useState<string | null>(null)
  const [rescheduleValues, setRescheduleValues] = useState<Record<number, string>>({})
  const [loading, setLoading] = useState(false)
  const [savingSettings, setSavingSettings] = useState(false)
  const [searchingBonusUser, setSearchingBonusUser] = useState(false)
  const [awardingBonus, setAwardingBonus] = useState(false)
  const [spendingBonus, setSpendingBonus] = useState(false)
  const [uploadingImage, setUploadingImage] = useState(false)
  const [uploadingBeforeAfter, setUploadingBeforeAfter] = useState<'before' | 'after' | null>(null)
  const [deletingChatId, setDeletingChatId] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [info, setInfo] = useState<string | null>(null)
  const [bonusPhone, setBonusPhone] = useState('')
  const [bonusUser, setBonusUser] = useState<User | null>(null)
  const [bonusAmount, setBonusAmount] = useState(0)
  const [bonusComment, setBonusComment] = useState('')
  const [settings, setSettings] = useState<NotificationSettings>({
    notify_telegram: true,
    notify_sms: false,
    admin_sms_phone: '',
  })

  const activeMeta = useMemo(() => adminSections.find((item) => item.id === section), [section])

  const load = async () => {
    setLoading(true)
    setError(null)
    setInfo(null)
    try {
      const [nextRequests, nextActive, nextCompleted, nextProcedures, nextReviews, nextBeforeAfter, nextMaster, nextSessions, notif] = await Promise.all([
        apiGet<AdminBooking[]>('/admin/booking-requests'),
        apiGet<AdminBooking[]>('/admin/active-bookings'),
        apiGet<AdminBooking[]>('/admin/completed-bookings'),
        apiGet<Procedure[]>('/admin/procedures'),
        apiGet<ProcedureReview[]>('/admin/reviews'),
        apiGet<BeforeAfterResult[]>('/admin/before-after-results?limit=300'),
        apiGet<MasterProfile>('/admin/master-profile'),
        apiGet<ChatSession[]>('/admin/chat-sessions?limit=300'),
        apiGet<NotificationSettings>('/admin/notification-settings'),
      ])
      setRequests(Array.isArray(nextRequests) ? nextRequests : [])
      setActiveBookings(Array.isArray(nextActive) ? nextActive : [])
      setCompletedBookings(Array.isArray(nextCompleted) ? nextCompleted : [])
      setProcedures(Array.isArray(nextProcedures) ? nextProcedures : [])
      setProcedureReviews(Array.isArray(nextReviews) ? nextReviews : [])
      setBeforeAfterResults(Array.isArray(nextBeforeAfter) ? nextBeforeAfter : [])
      setMasterProfile({
        name: nextMaster.name || '',
        title: nextMaster.title || '',
        bio: nextMaster.bio || '',
        photo_url: nextMaster.photo_url || '',
        certificates: Array.isArray(nextMaster.certificates) ? nextMaster.certificates : [],
        gallery: Array.isArray(nextMaster.gallery) ? nextMaster.gallery : [],
      })
      setChatSessions(Array.isArray(nextSessions) ? nextSessions : [])
      setSettings(notif)
    } catch (e) {
      setError((e as Error).message)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    if (hydrated) void load()
  }, [hydrated])

  const confirmBooking = async (id: number, datetime: string) => {
    setError(null)
    setInfo(null)
    try {
      await apiPost<{ ok: boolean }>(`/admin/bookings/${id}/confirm`, { datetime })
      setInfo('Заявка принята.')
      setBookingTab('active')
      await load()
    } catch (e) {
      setError((e as Error).message)
    }
  }

  const completeBooking = async (id: number) => {
    setError(null)
    setInfo(null)
    try {
      await apiPost<{ ok: boolean }>(`/admin/bookings/${id}/complete`)
      setInfo('Запись отмечена как проведённая.')
      await load()
    } catch (e) {
      setError((e as Error).message)
    }
  }

  const rescheduleBooking = async (booking: AdminBooking) => {
    setError(null)
    setInfo(null)
    try {
      const value = rescheduleValues[booking.id] ?? toDateTimeLocal(booking.datetime)
      const datetime = fromDateTimeLocal(value)
      if (!datetime) throw new Error('Укажите новую дату и время')
      await apiPost<{ ok: boolean }>(`/admin/bookings/${booking.id}/reschedule`, { datetime })
      setInfo('Запись перенесена.')
      await load()
    } catch (e) {
      setError((e as Error).message)
    }
  }

  const deleteBooking = async (id: number) => {
    setError(null)
    setInfo(null)
    try {
      await apiDelete<{ ok: boolean }>(`/admin/bookings/${id}`)
      setInfo('Заявка удалена.')
      await load()
    } catch (e) {
      setError((e as Error).message)
    }
  }

  const deleteChatSession = async (session: ChatSession) => {
    const confirmed = window.confirm(`Удалить AI-чат "${session.title}"? История диалога будет удалена без восстановления.`)
    if (!confirmed) return
    setError(null)
    setInfo(null)
    setDeletingChatId(session.id)
    try {
      await apiDelete<{ ok: boolean }>(`/admin/chat-sessions/${encodeURIComponent(session.id)}`)
      setInfo('AI-чат удалён.')
      setOpenedChatId((current) => (current === session.id ? null : current))
      await load()
    } catch (e) {
      setError((e as Error).message)
    } finally {
      setDeletingChatId(null)
    }
  }

  const renderBookingList = (items: AdminBooking[], mode: BookingTabId) => (
    <div className="space-y-3">
      {items.map((booking) => (
        <article key={booking.id} className="rounded-3xl border border-black/10 bg-white p-5">
          <div className="flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
            <div>
              <div className="text-sm font-semibold">{booking.user_name || 'Клиент без имени'}</div>
              <div className="mt-1 text-sm text-black/60">{booking.user_phone || 'телефон не указан'}</div>
              <div className="mt-1 text-sm text-black/60">{booking.user_email || 'email не указан'}</div>
            </div>
            <div className="rounded-full border border-black/10 px-3 py-1 text-xs text-black/55">{booking.status}</div>
          </div>

          <div className="mt-4 grid gap-3 md:grid-cols-2">
            <div className="rounded-2xl border border-black/10 p-4">
              <div className="text-xs text-black/45">Процедура</div>
              <div className="mt-1 text-sm font-semibold">{booking.procedure_title}</div>
            </div>
            <div className="rounded-2xl border border-black/10 p-4">
              <div className="text-xs text-black/45">{mode === 'completed' ? 'Дата проведения' : 'Дата записи'}</div>
              <div className="mt-1 text-sm font-semibold">{formatDate(booking.datetime)}</div>
            </div>
          </div>

          <div className="mt-4 rounded-2xl border border-black/10 p-4">
            <div className="text-xs text-black/45">Варианты клиента</div>
            <div className="mt-2 flex flex-wrap gap-2">
              {(booking.requested_datetimes.length ? booking.requested_datetimes : [booking.datetime]).map((date) => (
                <button
                  key={date}
                  type="button"
                  disabled={mode !== 'requests'}
                  onClick={() => {
                    if (mode !== 'requests') return
                    void confirmBooking(booking.id, date)
                  }}
                  className="rounded-full border border-black/10 px-3 py-1 text-xs font-semibold text-black/65 disabled:cursor-default"
                >
                  {formatDate(date)}
                </button>
              ))}
            </div>
          </div>

          <div className="mt-4 rounded-2xl border border-black/10 bg-black/[0.02] p-4 text-sm text-black/70">
            {booking.comment || 'Сообщения нет'}
          </div>

          {mode === 'requests' && (
            <div className="mt-4 flex flex-col gap-3 sm:flex-row">
              <button
                type="button"
                className="h-10 rounded-2xl bg-black px-4 text-sm font-semibold text-white"
                onClick={() => void confirmBooking(booking.id, booking.requested_datetimes[0] || booking.datetime)}
              >
                Принять
              </button>
              <button
                type="button"
                className="h-10 rounded-2xl border border-red-200 px-4 text-sm font-semibold text-red-700 hover:border-red-300"
                onClick={() => void deleteBooking(booking.id)}
              >
                Удалить
              </button>
            </div>
          )}

          {mode === 'active' && (
            <div className="mt-4 grid gap-3 md:grid-cols-[1fr_auto_auto] md:items-end">
              <div>
                <label className="text-xs font-semibold text-black/60">Новая дата и время</label>
                <input
                  type="datetime-local"
                  className="mt-1 h-11 w-full rounded-2xl border border-black/10 px-3 text-sm outline-none focus:border-black/30"
                  value={rescheduleValues[booking.id] ?? toDateTimeLocal(booking.datetime)}
                  onChange={(e) => setRescheduleValues((values) => ({ ...values, [booking.id]: e.target.value }))}
                />
              </div>
              <button
                type="button"
                className="h-11 rounded-2xl border border-black/15 px-4 text-sm font-semibold text-black hover:border-black/35"
                onClick={() => void rescheduleBooking(booking)}
              >
                Перенести
              </button>
              <button
                type="button"
                className="h-11 rounded-2xl bg-black px-4 text-sm font-semibold text-white"
                onClick={() => void completeBooking(booking.id)}
              >
                Проведена
              </button>
            </div>
          )}
        </article>
      ))}
      {!items.length && (
        <div className="rounded-3xl border border-black/10 bg-white p-10 text-center text-sm text-black/50">
          Ничего нет.
        </div>
      )}
    </div>
  )

  const renderBookings = () => {
    const items =
      bookingTab === 'requests' ? requests : bookingTab === 'active' ? activeBookings : completedBookings
    return (
      <div>
        <div className="mb-4 flex flex-wrap gap-2">
          {bookingTabs.map((tab) => (
            <button
              key={tab.id}
              type="button"
              onClick={() => setBookingTab(tab.id)}
              className={[
                'h-10 rounded-2xl border px-4 text-sm font-semibold transition',
                bookingTab === tab.id
                  ? 'border-black bg-black text-white'
                  : 'border-black/15 bg-white text-black/80 hover:border-black/30 hover:text-black',
              ].join(' ')}
            >
              {tab.label}
            </button>
          ))}
        </div>
        {renderBookingList(items, bookingTab)}
      </div>
    )
  }

  const renderChats = () => (
    <div className="space-y-3">
      {chatSessions.map((session) => {
        const opened = openedChatId === session.id
        return (
          <article key={session.id} className="overflow-hidden rounded-3xl border border-black/10 bg-white">
            <div className="flex flex-col gap-3 p-5 transition hover:bg-black/[0.02] md:flex-row md:items-center md:justify-between">
              <button
                type="button"
                aria-expanded={opened}
                onClick={() => setOpenedChatId(opened ? null : session.id)}
                className="flex-1 text-left"
              >
                <div>
                  <div className="text-sm font-semibold">{session.title}</div>
                  <div className="mt-1 text-sm text-black/60">
                    {session.email || 'email не указан'} · {session.phone || 'телефон не указан'}
                  </div>
                </div>
              </button>
              <div className="flex items-center gap-3">
                <div className="text-xs text-black/50">{formatDate(session.last_at)}</div>
                <button
                  type="button"
                  disabled={deletingChatId === session.id}
                  onClick={() => void deleteChatSession(session)}
                  className="h-9 rounded-2xl border border-red-200 px-3 text-xs font-semibold text-red-700 transition hover:border-red-300 disabled:opacity-45"
                >
                  {deletingChatId === session.id ? 'Удаляем...' : 'Удалить'}
                </button>
              </div>
            </div>
            {opened && (
              <div className="space-y-3 border-t border-black/10 p-5">
                {session.messages.map((message, index) => (
                  <div
                    key={`${message.created_at}-${index}`}
                    className={[
                      'max-w-3xl rounded-2xl border p-4 text-sm leading-relaxed',
                      message.role === 'assistant'
                        ? 'border-black/10 bg-black/[0.02] text-black/70'
                        : 'ml-auto border-black/15 bg-white text-black',
                    ].join(' ')}
                  >
                    <div className="mb-2 text-xs font-semibold uppercase text-black/40">
                      {message.role === 'assistant' ? 'AI' : 'Клиент'} · {formatDate(message.created_at)}
                    </div>
                    <div className="whitespace-pre-line">{message.text}</div>
                  </div>
                ))}
              </div>
            )}
          </article>
        )
      })}
      {!chatSessions.length && (
        <div className="rounded-3xl border border-black/10 bg-white p-10 text-center text-sm text-black/50">
          Истории AI-диалогов пока нет.
        </div>
      )}
    </div>
  )

  const saveProcedure = async () => {
    setError(null)
    setInfo(null)
    try {
      const body = {
        title: procedureForm.title,
        description: procedureForm.description,
        duration_mins: Number(procedureForm.duration_mins),
        price: Number(procedureForm.price),
        bonus_earned: Number(procedureForm.bonus_earned),
        category: procedureForm.category,
        image_url: procedureForm.image_url,
        video_url: procedureForm.video_url,
        services: procedureForm.services,
        duration_str: procedureForm.duration_str,
        popular: procedureForm.popular,
        is_active: procedureForm.is_active,
      }
      if (procedureForm.id) {
        await apiPut<Procedure>(`/admin/procedures/${procedureForm.id}`, body)
        setInfo('Процедура обновлена.')
      } else {
        await apiPost<Procedure>('/admin/procedures', body)
        setInfo('Процедура добавлена.')
      }
      setProcedureForm(emptyProcedure)
      await load()
    } catch (e) {
      setError((e as Error).message)
    }
  }

  const uploadProcedureImage = async (file: File | null) => {
    if (!file) return
    setUploadingImage(true)
    setError(null)
    setInfo(null)
    try {
      const form = new FormData()
      form.append('kind', 'image')
      form.append('file', file)
      const out = await apiUpload<UploadResult>('/admin/uploads/procedure-media', form)
      setProcedureForm((p) => ({ ...p, image_url: out.url }))
      setInfo('Картинка загружена и оптимизирована.')
    } catch (e) {
      setError((e as Error).message)
    } finally {
      setUploadingImage(false)
    }
  }

  const uploadImage = async (file: File | null) => {
    if (!file) return ''
    const form = new FormData()
    form.append('kind', 'image')
    form.append('file', file)
    const out = await apiUpload<UploadResult>('/admin/uploads/procedure-media', form)
    return out.url
  }



  const uploadBeforeAfterImage = async (field: 'before' | 'after', file: File | null) => {
    if (!file) return
    setUploadingBeforeAfter(field)
    setError(null)
    setInfo(null)
    try {
      const url = await uploadImage(file)
      if (url) {
        setBeforeAfterForm((item) => ({ ...item, [field === 'before' ? 'before_url' : 'after_url']: url }))
        setInfo('Фото загружено и оптимизировано.')
      }
    } catch (e) {
      setError((e as Error).message)
    } finally {
      setUploadingBeforeAfter(null)
    }
  }

  const saveBeforeAfter = async () => {
    setError(null)
    setInfo(null)
    try {
      const body = {
        procedure_id: beforeAfterForm.procedure_id || null,
        procedure: beforeAfterForm.procedure,
        title: beforeAfterForm.title,
        description: beforeAfterForm.description,
        before_url: beforeAfterForm.before_url,
        after_url: beforeAfterForm.after_url,
        is_featured: beforeAfterForm.is_featured,
        sort_order: Number(beforeAfterForm.sort_order),
        is_active: beforeAfterForm.is_active,
      }
      if (beforeAfterForm.id) {
        await apiPut<BeforeAfterResult>(`/admin/before-after-results/${beforeAfterForm.id}`, body)
        setInfo('Результат обновлён.')
      } else {
        await apiPost<BeforeAfterResult>('/admin/before-after-results', body)
        setInfo('Результат добавлен.')
      }
      setBeforeAfterForm(emptyBeforeAfter)
      await load()
    } catch (e) {
      setError((e as Error).message)
    }
  }

  const deleteBeforeAfter = async (item: BeforeAfterResult) => {
    const title = item.title || item.procedure || '#' + item.id
    const confirmed = window.confirm(`Удалить результат "${title}"? Это действие нельзя отменить.`)
    if (!confirmed) return
    setError(null)
    setInfo(null)
    try {
      await apiDelete<{ ok: boolean }>(`/admin/before-after-results/${item.id}`)
      setInfo('Результат удалён.')
      if (beforeAfterForm.id === item.id) setBeforeAfterForm(emptyBeforeAfter)
      await load()
    } catch (e) {
      setError((e as Error).message)
    }
  }

  const renderMaster = () => (
    <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_360px]">
      <div className="rounded-3xl border border-black/10 bg-white p-5">
        <div className="text-sm font-semibold">Профиль мастера</div>
        <div className="mt-4 grid gap-3 md:grid-cols-2">
          <label className="block">
            <span className="text-xs font-semibold text-black/60">Имя</span>
            <input className="mt-1 h-10 w-full rounded-2xl border border-black/10 px-3 text-sm" value={masterProfile.name} onChange={(e) => setMasterProfile((p) => ({ ...p, name: e.target.value }))} />
          </label>
          <label className="block">
            <span className="text-xs font-semibold text-black/60">Специализация</span>
            <input className="mt-1 h-10 w-full rounded-2xl border border-black/10 px-3 text-sm" value={masterProfile.title} onChange={(e) => setMasterProfile((p) => ({ ...p, title: e.target.value }))} />
          </label>
        </div>
        <label className="mt-3 block">
          <span className="text-xs font-semibold text-black/60">Биография</span>
          <textarea className="mt-1 min-h-32 w-full rounded-2xl border border-black/10 p-3 text-sm" value={masterProfile.bio} onChange={(e) => setMasterProfile((p) => ({ ...p, bio: e.target.value }))} />
        </label>
        <label className="mt-3 block">
          <span className="text-xs font-semibold text-black/60">Главное фото</span>
          <input className="mt-1 h-10 w-full rounded-2xl border border-black/10 px-3 text-sm" value={masterProfile.photo_url} onChange={(e) => setMasterProfile((p) => ({ ...p, photo_url: e.target.value }))} />
          <input
            type="file"
            accept="image/jpeg,image/png,image/webp,image/heic,image/heif,.heic,.heif"
            className="mt-2 block w-full text-xs text-black/60"
            onChange={async (e) => {
              const url = await uploadImage(e.target.files?.[0] ?? null)
              if (url) setMasterProfile((p) => ({ ...p, photo_url: url }))
            }}
          />
        </label>
        <button
          type="button"
          className="mt-4 h-11 rounded-2xl bg-black px-5 text-sm font-semibold text-white"
          onClick={async () => {
            setError(null)
            setInfo(null)
            try {
              const saved = await apiPost<MasterProfile>('/admin/master-profile', masterProfile)
              setMasterProfile(saved)
              setInfo('Информация о мастере сохранена.')
            } catch (e) {
              setError((e as Error).message)
            }
          }}
        >
          Сохранить
        </button>
      </div>

      <div className="space-y-4">
        {(['certificates', 'gallery'] as const).map((field) => (
          <div key={field} className="rounded-3xl border border-black/10 bg-white p-5">
            <div className="text-sm font-semibold">{field === 'certificates' ? 'Сертификаты' : 'Фотографии'}</div>
            <input
              type="file"
              accept="image/jpeg,image/png,image/webp,image/heic,image/heif,.heic,.heif"
              className="mt-3 block w-full text-xs text-black/60"
              onChange={async (e) => {
                const url = await uploadImage(e.target.files?.[0] ?? null)
                if (url) setMasterProfile((p) => ({ ...p, [field]: [...p[field], url] }))
              }}
            />
            <div className="mt-3 grid grid-cols-2 gap-2">
              {masterProfile[field].map((url) => (
                <div key={url} className="overflow-hidden rounded-2xl border border-black/10">
                  <img src={url} className="aspect-square w-full object-cover" />
                  <button
                    type="button"
                    className="w-full border-t border-black/10 py-2 text-xs font-semibold text-red-700"
                    onClick={() => setMasterProfile((p) => ({ ...p, [field]: p[field].filter((item) => item !== url) }))}
                  >
                    Удалить
                  </button>
                </div>
              ))}
            </div>
          </div>
        ))}
      </div>
    </div>
  )

  const renderProcedures = () => (
    <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_360px]">
      <div className="space-y-3">
        {procedures.map((p) => (
          <article key={p.id} className="rounded-3xl border border-black/10 bg-white p-5">
            <div className="flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
              <div>
                <div className="text-sm text-black/50">{p.category || 'Процедура'}</div>
                <div className="mt-1 text-lg font-semibold">{p.title}</div>
                <div className="mt-2 text-sm text-black/70">{p.description}</div>
                <div className="mt-2 text-sm text-black/55">
                  {p.price} ₽ · {p.duration_str || `${p.duration_mins} мин`} · {p.is_active ? 'активна' : 'скрыта'}
                </div>
              </div>
              <div className="flex gap-2">
                <button
                  type="button"
                  className="h-10 rounded-2xl border border-black/15 px-4 text-sm font-semibold"
                  onClick={() => setProcedureForm(p)}
                >
                  Редактировать
                </button>
                <button
                  type="button"
                  className="h-10 rounded-2xl border border-red-200 px-4 text-sm font-semibold text-red-700"
                  onClick={async () => {
                    const confirmed = window.confirm(`Удалить процедуру "${p.title}"? Это действие нельзя отменить.`)
                    if (!confirmed) return
                    setError(null)
                    setInfo(null)
                    try {
                      await apiDelete<{ ok: boolean }>(`/admin/procedures/${p.id}`)
                      setInfo('Процедура удалена.')
                      await load()
                    } catch (e) {
                      setError((e as Error).message)
                    }
                  }}
                >
                  Удалить
                </button>
              </div>
            </div>
          </article>
        ))}
        <div className="rounded-3xl border border-black/10 bg-white p-5">
          <div className="text-sm font-semibold">Отзывы</div>
          <div className="mt-3 space-y-3">
            {procedureReviews.map((review) => (
              <div key={review.id} className="rounded-2xl border border-black/10 p-4 text-sm">
                <div className="flex items-start justify-between gap-3">
                  <div>
                    <div className="font-semibold">{review.procedure_title || `Процедура #${review.procedure_id}`}</div>
                    <div className="mt-1 text-black/60">
                      {review.user_name || 'Клиент'} · {review.rating}/5
                    </div>
                    <div className="mt-2 text-black/70">{review.text || 'Без текста'}</div>
                  </div>
                  <button
                    type="button"
                    className="rounded-2xl border border-red-200 px-3 py-2 text-xs font-semibold text-red-700"
                    onClick={async () => {
                      const confirmed = window.confirm('Удалить отзыв? Это действие нельзя отменить.')
                      if (!confirmed) return
                      setError(null)
                      setInfo(null)
                      try {
                        await apiDelete<{ ok: boolean }>(`/admin/reviews/${review.id}`)
                        setInfo('Отзыв удалён.')
                        await load()
                      } catch (e) {
                        setError((e as Error).message)
                      }
                    }}
                  >
                    Удалить
                  </button>
                </div>
              </div>
            ))}
            {!procedureReviews.length && <div className="text-sm text-black/50">Отзывов пока нет.</div>}
          </div>
        </div>
      </div>

      <div className="rounded-3xl border border-black/10 bg-white p-5">
        <div className="text-sm font-semibold">{procedureForm.id ? 'Редактировать процедуру' : 'Новая процедура'}</div>
        <div className="mt-4 space-y-3">
          {[
            ['title', 'Название'],
            ['category', 'Категория'],
            ['duration_str', 'Длительность текстом'],
          ].map(([key, label]) => (
            <label key={key} className="block">
              <span className="text-xs font-semibold text-black/60">{label}</span>
              <input
                className="mt-1 h-10 w-full rounded-2xl border border-black/10 px-3 text-sm outline-none focus:border-black/30"
                value={String(procedureForm[key as keyof Procedure] ?? '')}
                onChange={(e) => setProcedureForm((p) => ({ ...p, [key]: e.target.value }))}
              />
            </label>
          ))}
          <label className="block">
            <span className="text-xs font-semibold text-black/60">Картинка</span>
            <input
              className="mt-1 h-10 w-full rounded-2xl border border-black/10 px-3 text-sm outline-none focus:border-black/30"
              value={procedureForm.image_url}
              onChange={(e) => setProcedureForm((p) => ({ ...p, image_url: e.target.value }))}
              placeholder="/uploads/procedures/images/..."
            />
            <input
              type="file"
              accept="image/jpeg,image/png,image/webp,image/heic,image/heif,.heic,.heif"
              className="mt-2 block w-full text-xs text-black/60"
              onChange={(e) => void uploadProcedureImage(e.target.files?.[0] ?? null)}
            />
            {uploadingImage && <div className="mt-1 text-xs text-black/50">Оптимизируем картинку...</div>}
          </label>
          <div className="grid grid-cols-3 gap-2">
            {[
              ['duration_mins', 'Минуты'],
              ['price', 'Цена'],
              ['bonus_earned', 'Бонусы'],
            ].map(([key, label]) => (
              <label key={key} className="block">
                <span className="text-xs font-semibold text-black/60">{label}</span>
                <input
                  type="number"
                  className="mt-1 h-10 w-full rounded-2xl border border-black/10 px-3 text-sm outline-none focus:border-black/30"
                  value={Number(procedureForm[key as keyof Procedure] ?? 0)}
                  onChange={(e) => setProcedureForm((p) => ({ ...p, [key]: Number(e.target.value) }))}
                />
              </label>
            ))}
          </div>
          <label className="block">
            <span className="text-xs font-semibold text-black/60">Описание</span>
            <textarea
              className="mt-1 min-h-24 w-full rounded-2xl border border-black/10 p-3 text-sm outline-none focus:border-black/30"
              value={procedureForm.description}
              onChange={(e) => setProcedureForm((p) => ({ ...p, description: e.target.value }))}
            />
          </label>
          <label className="block">
            <span className="text-xs font-semibold text-black/60">Что входит</span>
            <textarea
              className="mt-1 min-h-20 w-full rounded-2xl border border-black/10 p-3 text-sm outline-none focus:border-black/30"
              value={procedureForm.services}
              onChange={(e) => setProcedureForm((p) => ({ ...p, services: e.target.value }))}
            />
          </label>
          <div className="flex gap-3">
            <label className="flex items-center gap-2 text-sm">
              <input
                type="checkbox"
                checked={procedureForm.popular}
                onChange={(e) => setProcedureForm((p) => ({ ...p, popular: e.target.checked }))}
              />
              Популярная
            </label>
            <label className="flex items-center gap-2 text-sm">
              <input
                type="checkbox"
                checked={procedureForm.is_active}
                onChange={(e) => setProcedureForm((p) => ({ ...p, is_active: e.target.checked }))}
              />
              Активна
            </label>
          </div>
          <div className="flex gap-2">
            <button type="button" className="h-11 flex-1 rounded-2xl bg-black text-sm font-semibold text-white" onClick={() => void saveProcedure()}>
              Сохранить
            </button>
            <button type="button" className="h-11 rounded-2xl border border-black/15 px-4 text-sm font-semibold" onClick={() => setProcedureForm(emptyProcedure)}>
              Сброс
            </button>
          </div>
        </div>
      </div>
    </div>
  )

  const renderBonus = () => (
    <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_380px]">
      <div className="rounded-3xl border border-black/10 bg-white p-5">
        <div className="text-sm font-semibold">Найти аккаунт</div>
        <div className="mt-4 flex flex-col gap-3 sm:flex-row">
          <div className="flex-1">
            <label className="text-xs font-semibold text-black/60">Телефон пользователя</label>
            <input
              className="mt-1 h-11 w-full rounded-2xl border border-black/10 px-3 text-sm outline-none focus:border-black/30"
              value={bonusPhone}
              onChange={(e) => {
                setBonusPhone(e.target.value)
                setBonusUser(null)
              }}
              placeholder="+79001234567"
            />
          </div>
          <button
            type="button"
            disabled={searchingBonusUser || !bonusPhone.trim()}
            className="mt-5 h-11 rounded-2xl bg-black px-5 text-sm font-semibold text-white disabled:opacity-40 sm:mt-6"
            onClick={async () => {
              setSearchingBonusUser(true)
              setError(null)
              setInfo(null)
              setBonusUser(null)
              try {
                const resp = await apiGet<{ user: User; bonus_points: number }>(
                  `/admin/bonus/user?phone=${encodeURIComponent(bonusPhone)}`,
                )
                setBonusUser(resp.user)
                setBonusAmount(0)
                setBonusComment('')
              } catch (e) {
                setError((e as Error).message)
              } finally {
                setSearchingBonusUser(false)
              }
            }}
          >
            {searchingBonusUser ? 'Ищем...' : 'Найти'}
          </button>
        </div>

        {bonusUser && (
          <div className="mt-5 rounded-3xl border border-black/10 p-5">
            <div className="grid gap-4 md:grid-cols-4">
              <div>
                <div className="text-xs text-black/50">Фамилия и имя</div>
                <div className="mt-1 text-sm font-semibold">{bonusUser.name || 'Не указано'}</div>
              </div>
              <div>
                <div className="text-xs text-black/50">Телефон</div>
                <div className="mt-1 text-sm font-semibold">{bonusUser.phone || 'Не указан'}</div>
              </div>
              <div>
                <div className="text-xs text-black/50">Email</div>
                <div className="mt-1 break-all text-sm font-semibold">{bonusUser.email || 'Не указан'}</div>
              </div>
              <div>
                <div className="text-xs text-black/50">Бонусы</div>
                <div className="mt-1 text-2xl font-semibold">{bonusUser.bonus_points}</div>
              </div>
            </div>

            <div className="mt-5 grid gap-4 md:grid-cols-[180px_minmax(0,1fr)_auto_auto] md:items-end">
              <div>
                <label className="text-xs font-semibold text-black/60">Количество бонусов</label>
                <input
                  type="number"
                  min={1}
                  className="mt-1 h-11 w-full rounded-2xl border border-black/10 px-3 text-sm outline-none focus:border-black/30"
                  value={bonusAmount}
                  onChange={(e) => setBonusAmount(Number(e.target.value))}
                />
              </div>
              <div>
                <label className="text-xs font-semibold text-black/60">Комментарий</label>
                <input
                  className="mt-1 h-11 w-full rounded-2xl border border-black/10 px-3 text-sm outline-none focus:border-black/30"
                  value={bonusComment}
                  onChange={(e) => setBonusComment(e.target.value)}
                  placeholder="Например: оплата визита"
                />
              </div>
              <button
                type="button"
                disabled={awardingBonus || spendingBonus || bonusAmount <= 0}
                className="h-11 rounded-2xl border border-black/15 px-5 text-sm font-semibold text-black/70 hover:border-black/30 disabled:opacity-40"
                onClick={async () => {
                  setAwardingBonus(true)
                  setError(null)
                  setInfo(null)
                  try {
                    const resp = await apiPost<{ user: User; bonus_points: number }>('/admin/bonus/award', {
                      phone: bonusUser.phone || bonusPhone,
                      amount: bonusAmount,
                      comment: bonusComment,
                    })
                    setBonusUser(resp.user)
                    setInfo(`Зачислено. ${resp.user.name || resp.user.phone}: ${resp.bonus_points} бонусов.`)
                    setBonusAmount(0)
                    setBonusComment('')
                  } catch (e) {
                    setError((e as Error).message)
                  } finally {
                    setAwardingBonus(false)
                  }
                }}
              >
                {awardingBonus ? 'Зачисляем...' : 'Зачислить'}
              </button>
              <button
                type="button"
                disabled={awardingBonus || spendingBonus || bonusAmount <= 0 || bonusAmount > bonusUser.bonus_points}
                className="h-11 rounded-2xl bg-black px-5 text-sm font-semibold text-white disabled:opacity-40"
                onClick={async () => {
                  setSpendingBonus(true)
                  setError(null)
                  setInfo(null)
                  try {
                    const resp = await apiPost<{ user: User; bonus_points: number }>('/admin/bonus/spend', {
                      phone: bonusUser.phone || bonusPhone,
                      amount: bonusAmount,
                      comment: bonusComment,
                    })
                    setBonusUser(resp.user)
                    setInfo(`Списано. ${resp.user.name || resp.user.phone}: ${resp.bonus_points} бонусов.`)
                    setBonusAmount(0)
                    setBonusComment('')
                  } catch (e) {
                    setError((e as Error).message)
                  } finally {
                    setSpendingBonus(false)
                  }
                }}
              >
                {spendingBonus ? 'Списываем...' : 'Списать'}
              </button>
            </div>
          </div>
        )}
      </div>

      <div className="rounded-3xl border border-black/10 bg-white p-5 text-sm text-black/70">
        <div className="text-sm font-semibold text-black">Работа с бонусами</div>
        <div className="mt-2">Введите телефон клиента, откройте найденный аккаунт и измените баланс бонусов.</div>
        <div className="mt-3 rounded-2xl border border-black/10 p-4">
          Списание не пройдет, если бонусов недостаточно или аккаунт заблокирован. Зачисление требует положительное количество бонусов.
        </div>
      </div>
    </div>
  )


  const renderBeforeAfter = () => (
    <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_380px]">
      <div className="space-y-3">
        {beforeAfterResults.map((item) => (
          <article key={item.id} className="rounded-3xl border border-black/10 bg-white p-5">
            <div className="grid gap-4 md:grid-cols-[260px_1fr]">
              <div className="grid grid-cols-2 overflow-hidden rounded-2xl border border-black/10">
                <img src={item.before_url} alt="До" className="aspect-[3/4] w-full object-cover" />
                <img src={item.after_url} alt="После" className="aspect-[3/4] w-full object-cover" />
              </div>
              <div className="flex flex-col justify-between gap-4">
                <div>
                  <div className="text-xs text-black/50">{item.procedure || 'Процедура не указана'}</div>
                  <div className="mt-1 text-lg font-semibold">{item.title || 'Без названия'}</div>
                  {item.description && <div className="mt-2 whitespace-pre-line text-sm text-black/70">{item.description}</div>}
                  <div className="mt-3 flex flex-wrap gap-2 text-xs text-black/55">
                    <span className="rounded-full border border-black/10 px-3 py-1">{item.is_active ? 'Опубликовано' : 'Скрыто'}</span>
                    {item.is_featured && <span className="rounded-full border border-black/10 px-3 py-1">На главной</span>}
                    <span className="rounded-full border border-black/10 px-3 py-1">Порядок: {item.sort_order}</span>
                  </div>
                </div>
                <div className="flex flex-wrap gap-2">
                  <button type="button" className="h-10 rounded-2xl border border-black/15 px-4 text-sm font-semibold" onClick={() => setBeforeAfterForm(item)}>
                    Редактировать
                  </button>
                  <button type="button" className="h-10 rounded-2xl border border-red-200 px-4 text-sm font-semibold text-red-700" onClick={() => void deleteBeforeAfter(item)}>
                    Удалить
                  </button>
                </div>
              </div>
            </div>
          </article>
        ))}
        {!beforeAfterResults.length && (
          <div className="rounded-3xl border border-black/10 bg-white p-10 text-center text-sm text-black/50">
            Результатов пока нет. Добавьте первую пару фото справа.
          </div>
        )}
      </div>

      <div className="rounded-3xl border border-black/10 bg-white p-5">
        <div className="text-sm font-semibold">{beforeAfterForm.id ? 'Редактировать результат' : 'Новый результат'}</div>
        <div className="mt-4 grid grid-cols-2 gap-3">
          <div className="overflow-hidden rounded-2xl border border-black/10 bg-black/[0.02]">
            {beforeAfterForm.before_url ? <img src={beforeAfterForm.before_url} alt="До" className="aspect-[3/4] w-full object-cover" /> : <div className="flex aspect-[3/4] items-center justify-center text-xs text-black/45">До</div>}
          </div>
          <div className="overflow-hidden rounded-2xl border border-black/10 bg-black/[0.02]">
            {beforeAfterForm.after_url ? <img src={beforeAfterForm.after_url} alt="После" className="aspect-[3/4] w-full object-cover" /> : <div className="flex aspect-[3/4] items-center justify-center text-xs text-black/45">После</div>}
          </div>
        </div>

        <div className="mt-4 space-y-3">
          <label className="block">
            <span className="text-xs font-semibold text-black/60">Процедура</span>
            <select
              className="mt-1 h-10 w-full rounded-2xl border border-black/10 px-3 text-sm outline-none focus:border-black/30"
              value={beforeAfterForm.procedure_id || ''}
              onChange={(e) => {
                const id = Number(e.target.value)
                const procedure = procedures.find((p) => p.id === id)
                setBeforeAfterForm((item) => ({ ...item, procedure_id: id || null, procedure: procedure?.title || item.procedure }))
              }}
            >
              <option value="">Без привязки</option>
              {procedures.map((p) => (
                <option key={p.id} value={p.id}>{p.title}</option>
              ))}
            </select>
          </label>
          <label className="block">
            <span className="text-xs font-semibold text-black/60">Название</span>
            <input className="mt-1 h-10 w-full rounded-2xl border border-black/10 px-3 text-sm outline-none focus:border-black/30" value={beforeAfterForm.title} onChange={(e) => setBeforeAfterForm((item) => ({ ...item, title: e.target.value }))} />
          </label>
          <label className="block">
            <span className="text-xs font-semibold text-black/60">Описание</span>
            <textarea className="mt-1 min-h-24 w-full rounded-2xl border border-black/10 p-3 text-sm outline-none focus:border-black/30" value={beforeAfterForm.description} onChange={(e) => setBeforeAfterForm((item) => ({ ...item, description: e.target.value }))} />
          </label>

          {(['before', 'after'] as const).map((field) => (
            <label key={field} className="block">
              <span className="text-xs font-semibold text-black/60">Фото {field === 'before' ? 'до' : 'после'}</span>
              <input
                className="mt-1 h-10 w-full rounded-2xl border border-black/10 px-3 text-sm outline-none focus:border-black/30"
                value={field === 'before' ? beforeAfterForm.before_url : beforeAfterForm.after_url}
                onChange={(e) => setBeforeAfterForm((item) => ({ ...item, [field === 'before' ? 'before_url' : 'after_url']: e.target.value }))}
                placeholder="/uploads/procedures/images/..."
              />
              <input type="file" accept="image/jpeg,image/png,image/webp,image/heic,image/heif,.heic,.heif" className="mt-2 block w-full text-xs text-black/60" onChange={(e) => void uploadBeforeAfterImage(field, e.target.files?.[0] ?? null)} />
              {uploadingBeforeAfter === field && <div className="mt-1 text-xs text-black/50">Оптимизируем фото...</div>}
            </label>
          ))}

          <div className="grid grid-cols-2 gap-3">
            <label className="block">
              <span className="text-xs font-semibold text-black/60">Порядок</span>
              <input type="number" className="mt-1 h-10 w-full rounded-2xl border border-black/10 px-3 text-sm outline-none focus:border-black/30" value={beforeAfterForm.sort_order} onChange={(e) => setBeforeAfterForm((item) => ({ ...item, sort_order: Number(e.target.value) }))} />
            </label>
            <div className="space-y-2 pt-6">
              <label className="flex items-center gap-2 text-sm"><input type="checkbox" checked={beforeAfterForm.is_active} onChange={(e) => setBeforeAfterForm((item) => ({ ...item, is_active: e.target.checked }))} />Опубликовано</label>
              <label className="flex items-center gap-2 text-sm"><input type="checkbox" checked={beforeAfterForm.is_featured} onChange={(e) => setBeforeAfterForm((item) => ({ ...item, is_featured: e.target.checked }))} />Показывать на главной</label>
            </div>
          </div>
        </div>

        <div className="mt-4 flex gap-2">
          <button type="button" className="h-11 rounded-2xl bg-black px-5 text-sm font-semibold text-white" onClick={() => void saveBeforeAfter()}>
            {beforeAfterForm.id ? 'Сохранить' : 'Добавить'}
          </button>
          {beforeAfterForm.id > 0 && (
            <button type="button" className="h-11 rounded-2xl border border-black/15 px-5 text-sm font-semibold" onClick={() => setBeforeAfterForm(emptyBeforeAfter)}>
              Сбросить
            </button>
          )}
        </div>
      </div>
    </div>
  )

  const renderNotifications = () => (
    <div className="rounded-3xl border border-black/10 bg-white p-5">
      <div className="text-sm font-semibold">Уведомления о записях</div>
      <div className="mt-4 grid gap-4 md:grid-cols-2">
        <label className="flex items-center gap-3 rounded-2xl border border-black/10 p-4 text-sm">
          <input
            type="checkbox"
            checked={settings.notify_telegram}
            onChange={(e) => setSettings((s) => ({ ...s, notify_telegram: e.target.checked }))}
          />
          <span>Отправлять админу в Telegram</span>
        </label>
        <label className="flex items-center gap-3 rounded-2xl border border-black/10 p-4 text-sm">
          <input
            type="checkbox"
            checked={settings.notify_sms}
            onChange={(e) => setSettings((s) => ({ ...s, notify_sms: e.target.checked }))}
          />
          <span>Отправлять админу SMS</span>
        </label>
        <div>
          <label className="text-xs font-semibold text-black/60">Телефон админа для SMS</label>
          <input
            className="mt-1 h-11 w-full rounded-2xl border border-black/10 px-3 text-sm outline-none focus:border-black/30"
            value={settings.admin_sms_phone}
            onChange={(e) => setSettings((s) => ({ ...s, admin_sms_phone: e.target.value }))}
            placeholder="+79001234567"
          />
        </div>
      </div>
      <button
        type="button"
        disabled={savingSettings}
        className="mt-4 h-11 rounded-2xl bg-black px-5 text-sm font-semibold text-white disabled:opacity-40"
        onClick={async () => {
          setSavingSettings(true)
          setError(null)
          try {
            const next = await apiPost<NotificationSettings>('/admin/notification-settings', settings)
            setSettings(next)
            setInfo('Настройки уведомлений сохранены.')
          } catch (e) {
            setError((e as Error).message)
          } finally {
            setSavingSettings(false)
          }
        }}
      >
        {savingSettings ? 'Сохраняем...' : 'Сохранить уведомления'}
      </button>
    </div>
  )

  const renderCurrentSection = () => {
    if (section === 'bookings') return renderBookings()
    if (section === 'procedures') return renderProcedures()
    if (section === 'beforeAfter') return renderBeforeAfter()
    if (section === 'master') return renderMaster()
    if (section === 'ai') return renderChats()
    if (section === 'bonus') return renderBonus()
    return renderNotifications()
  }

  return (
    <MotionPage>
      <Helmet>
        <title>Админ-панель - Холодная плазма</title>
        <meta name="robots" content="noindex,nofollow" />
      </Helmet>

      <section className="py-10">
        <Container>
          <div className="flex flex-col gap-4 md:flex-row md:items-end md:justify-between">
            <div>
              <div className="text-3xl font-semibold tracking-tight">Админ-панель</div>
              <div className="mt-2 max-w-2xl text-sm text-black/60">
                {activeMeta?.description || 'Управление проектом'}
              </div>
            </div>
            <button
              type="button"
              onClick={() => void load()}
              disabled={loading}
              className="h-11 rounded-2xl bg-black px-5 text-sm font-semibold text-white disabled:opacity-40"
            >
              {loading ? 'Обновляем...' : 'Обновить'}
            </button>
          </div>

          <div className="mt-6 grid gap-3 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-7">
            {adminSections.map((item) => (
              <button
                key={item.id}
                type="button"
                onClick={() => setSection(item.id)}
                className={[
                  'rounded-3xl border p-4 text-left transition',
                  section === item.id
                    ? 'border-black bg-black text-white'
                    : 'border-black/10 bg-white text-black hover:border-black/25',
                ].join(' ')}
              >
                <div className="text-sm font-semibold">{item.label}</div>
                <div className={['mt-1 text-xs', section === item.id ? 'text-white/65' : 'text-black/50'].join(' ')}>
                  {item.description}
                </div>
              </button>
            ))}
          </div>

          {error && (
            <div className="mt-6 rounded-3xl border border-red-200 bg-red-50 p-5 text-sm text-red-700">
              {error}
            </div>
          )}
          {info && (
            <div className="mt-6 rounded-3xl border border-emerald-200 bg-emerald-50 p-5 text-sm text-emerald-800">
              {info}
            </div>
          )}

          <div className="mt-6">{renderCurrentSection()}</div>
        </Container>
      </section>
    </MotionPage>
  )
}
