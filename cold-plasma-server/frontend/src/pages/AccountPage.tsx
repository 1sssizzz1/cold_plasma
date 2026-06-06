import { Helmet } from 'react-helmet-async'
import { useEffect, useMemo, useState } from 'react'
import Container from '../components/Container'
import MotionPage from '../components/MotionPage'
import VKLoginButton from '../components/auth/VKLoginButton'
import { useAuthStore } from '../store/auth'
import { apiGet, apiPost } from '../utils/api'
import type { Booking, BonusLog, User } from '../types'

type Mode = 'login' | 'register' | 'forgot' | 'reset'

function emailLooksOk(v: string) {
  return v.trim().includes('@')
}

export default function AccountPage() {
  const token = useAuthStore((s) => s.token)
  const user = useAuthStore((s) => s.user)
  const setAuth = useAuthStore((s) => s.setAuth)
  const logout = useAuthStore((s) => s.logout)

  const [mode, setMode] = useState<Mode>('login')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [resetToken, setResetToken] = useState('')
  const [name, setName] = useState('')
  const [lastName, setLastName] = useState('')
  const [phone, setPhone] = useState('')
  const [phoneCode, setPhoneCode] = useState('')
  const [accountPhone, setAccountPhone] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [info, setInfo] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)

  const [bookings, setBookings] = useState<Booking[]>([])
  const [bonusBalance, setBonusBalance] = useState<number | null>(null)
  const [bonusLogs, setBonusLogs] = useState<BonusLog[]>([])
  const [telegramLinked, setTelegramLinked] = useState(false)
  const [telegramLinkURL, setTelegramLinkURL] = useState('')
  const [procedures, setProcedures] = useState<Record<number, string>>({})

  const apiBase = ((import.meta.env.VITE_API_URL as string | undefined) ?? '/api/v1') as string

  const statusTranslations: Record<string, string> = {
    'new': 'Новая заявка',
    'confirmed': 'Подтверждена',
    'completed': 'Завершена',
    'cancelled': 'Отменена',
  }

  const bonusTypeTranslations: Record<string, string> = {
    'earn': 'Начисление',
    'spend': 'Списание',
    'award': 'Награда',
    'refund': 'Возврат',
  }

  const formatDateTime = (dateString: string) => {
    const date = new Date(dateString)
    return date.toLocaleString('ru-RU', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    })
  }

  useEffect(() => {
    const params = new URLSearchParams(window.location.search)
    const verified = params.get('verified')
    if (verified === '1') {
      setInfo('Email подтверждён — теперь можно войти.')
    } else if (verified === '0') {
      setError('Не удалось подтвердить email. Попробуйте зарегистрироваться или войти позже.')
    }
    const token = params.get('reset_token')
    const resetEmail = params.get('email')
    if (token) {
      setResetToken(token)
      if (resetEmail) setEmail(resetEmail)
      setMode('reset')
      setInfo('Введите новый пароль.')
    }
  }, [])

  useEffect(() => {
    if (!token) return
    void apiGet<User>('/me')
      .then((u) => setAuth(token, u))
      .catch(() => {})
    void apiGet<Booking[]>('/bookings')
      .then((items) => setBookings(Array.isArray(items) ? items : []))
      .catch(() => setBookings([]))
    void apiGet<{ bonus_points: number }>('/bonus/balance')
      .then((d) => setBonusBalance(d.bonus_points))
      .catch(() => setBonusBalance(null))
    void apiGet<BonusLog[]>('/bonus/logs?limit=50')
      .then((items) => setBonusLogs(Array.isArray(items) ? items : []))
      .catch(() => setBonusLogs([]))
    void apiGet<{ linked: boolean }>('/telegram/status')
      .then((d) => setTelegramLinked(Boolean(d.linked)))
      .catch(() => setTelegramLinked(false))
    void apiGet<any[]>('/procedures')
      .then((items) => {
        const map: Record<number, string> = {}
        if (Array.isArray(items)) {
          items.forEach((proc) => {
            map[proc.id] = proc.title
          })
        }
        setProcedures(map)
      })
      .catch(() => setProcedures({}))
  }, [token, setAuth])

  useEffect(() => {
    if (user?.phone) setAccountPhone(user.phone)
  }, [user?.phone])

  const title = useMemo(() => (token ? 'Личный кабинет' : 'Войти / регистрация'), [token])

  return (
    <MotionPage>
      <Helmet>
        <title>{title} — Холодная плазма (Северодвинск)</title>
        <meta name="description" content="Личный кабинет: записи, бонусы, памятка по уходу и быстрый вход." />
      </Helmet>

      <section className="py-12">
        <Container>
          <div className="border-b border-white/5 pb-6">
            <div className="text-xs font-semibold uppercase tracking-widest text-zinc-500">Аккаунт</div>
            <div className="mt-2 text-3xl font-semibold tracking-tight text-white sm:text-4xl">Личный кабинет</div>
          </div>

          {!token ? (
            <div className="mt-8 grid gap-6 md:grid-cols-2">
              <div className="glass-panel rounded-3xl p-6 shadow-xl">
                <div className="flex gap-2">
                  <button
                    className={[
                      'h-10 flex-1 rounded-2xl text-xs font-semibold uppercase tracking-wider transition',
                      mode === 'login' ? 'bg-black text-white' : 'border border-black/20 bg-white text-zinc-700 hover:text-black hover:border-black/30',
                    ].join(' ')}
                    onClick={() => setMode('login')}
                  >
                    Вход
                  </button>
                  <button
                    className={[
                      'h-10 flex-1 rounded-2xl text-xs font-semibold uppercase tracking-wider transition',
                      mode === 'register' ? 'bg-black text-white' : 'border border-black/20 bg-white text-zinc-700 hover:text-black hover:border-black/30',
                    ].join(' ')}
                    onClick={() => setMode('register')}
                  >
                    Регистрация
                  </button>
                </div>

                {error && (
                  <div className="mt-4 rounded-2xl border border-red-200 bg-red-50 p-3 text-sm font-medium text-red-700">
                    {error}
                  </div>
                )}
                {info && (
                  <div className="mt-4 rounded-2xl border border-emerald-200 bg-emerald-50 p-3 text-sm font-medium text-emerald-800">
                    {info}
                  </div>
                )}

                <div className="mt-4">
                  <label className="text-xs font-semibold uppercase tracking-wider text-zinc-500">Email</label>
                  <input
                    type="email"
                    className="mt-2 h-11 w-full rounded-2xl border border-white/10 bg-zinc-950 px-3 text-sm text-white outline-none focus:border-white/30"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                  />
                  {mode === 'register' && email && !emailLooksOk(email) && (
                    <div className="mt-1 text-xs text-red-300">Email должен содержать символ “@”.</div>
                  )}
                </div>

                {mode === 'reset' && (
                  <div className="mt-4">
                    <label className="text-xs font-semibold uppercase tracking-wider text-zinc-500">Токен восстановления</label>
                    <input
                      className="mt-2 h-11 w-full rounded-2xl border border-white/10 bg-zinc-950 px-3 text-sm text-white outline-none focus:border-white/30"
                      value={resetToken}
                      onChange={(e) => setResetToken(e.target.value)}
                    />
                  </div>
                )}

                {mode !== 'forgot' && (
                <div className="mt-4">
                  <label className="text-xs font-semibold uppercase tracking-wider text-zinc-500">Пароль</label>
                  <input
                    type="password"
                    className="mt-2 h-11 w-full rounded-2xl border border-white/10 bg-zinc-950 px-3 text-sm text-white outline-none focus:border-white/30"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                  />
                </div>
                )}

                {mode === 'register' && (
                  <>
                    <div className="mt-4">
                      <label className="text-xs font-semibold uppercase tracking-wider text-zinc-500">Имя</label>
                      <input
                        className="mt-2 h-11 w-full rounded-2xl border border-white/10 bg-zinc-950 px-3 text-sm text-white outline-none focus:border-white/30"
                        value={name}
                        onChange={(e) => setName(e.target.value)}
                      />
                    </div>
                    <div className="mt-4">
                      <label className="text-xs font-semibold uppercase tracking-wider text-zinc-500">Фамилия</label>
                      <input
                        className="mt-2 h-11 w-full rounded-2xl border border-white/10 bg-zinc-950 px-3 text-sm text-white outline-none focus:border-white/30"
                        value={lastName}
                        onChange={(e) => setLastName(e.target.value)}
                      />
                    </div>
                    <div className="mt-4">
                      <label className="text-xs font-semibold uppercase tracking-wider text-zinc-500">Телефон</label>
                      <input
                        className="mt-2 h-11 w-full rounded-2xl border border-white/10 bg-zinc-950 px-3 text-sm text-white outline-none placeholder:text-zinc-600 focus:border-white/30"
                        value={phone}
                        onChange={(e) => setPhone(e.target.value)}
                        placeholder="+7 900 000-00-00"
                      />
                      <div className="mt-1 text-xs text-zinc-500">Только российский мобильный номер.</div>
                    </div>
                  </>
                )}

                <button
                  disabled={
                    loading || (mode === 'register' && (!emailLooksOk(email) || !name.trim() || !lastName.trim()))
                  }
                  className="mt-6 h-11 w-full rounded-2xl bg-black text-xs font-semibold uppercase tracking-wider text-white transition hover:bg-zinc-800 disabled:opacity-40"
                  onClick={async () => {
                    setError(null)
                    setInfo(null)
                    setLoading(true)
                    try {
                      if (mode === 'login') {
                        const resp = await apiPost<{ user: User; token: string }>('/auth/login', { email, password })
                        setAuth(resp.token, resp.user)
                      } else if (mode === 'register') {
                        if (!emailLooksOk(email)) throw new Error('Укажите корректный email (должен содержать “@”)')
                        if (!name.trim() || !lastName.trim()) throw new Error('Укажите имя и фамилию')
                        const resp = await apiPost<{
                          user: User
                          verification_required: boolean
                          verification_sent: boolean
                        }>('/auth/register', {
                          email,
                          password,
                          name: name.trim(),
                          last_name: lastName.trim(),
                          phone,
                        })
                        if (resp.verification_sent) {
                          setInfo('Мы отправили письмо для подтверждения. Перейдите по ссылке из письма и затем войдите.')
                        } else {
                          setInfo('Аккаунт создан, но письмо не удалось отправить. Попробуйте войти позже.')
                        }
                        setMode('login')
                      } else if (mode === 'forgot') {
                        if (!emailLooksOk(email)) throw new Error('Укажите email с “@”')
                        await apiPost<{ ok: boolean }>('/auth/forgot-password', { email })
                        setInfo('Если email зарегистрирован, мы отправили ссылку для восстановления.')
                      } else {
                        if (!emailLooksOk(email)) throw new Error('Укажите email с “@”')
                        await apiPost<{ ok: boolean }>('/auth/reset-password', {
                          email,
                          token: resetToken,
                          password,
                        })
                        setInfo('Пароль изменён. Теперь можно войти.')
                        setMode('login')
                      }
                    } catch (e) {
                      setError(mode === 'login' ? 'Неверный email или пароль' : (e as Error).message)
                    } finally {
                      setLoading(false)
                    }
                  }}
                >
                  {mode === 'login'
                    ? 'Войти'
                    : mode === 'register'
                      ? 'Создать аккаунт'
                      : mode === 'forgot'
                        ? 'Отправить ссылку'
                        : 'Сменить пароль'}
                </button>

                {mode === 'login' && (
                  <button
                    type="button"
                    className="mt-3 h-11 w-full rounded-2xl border border-white/10 bg-white/5 text-xs font-semibold uppercase tracking-wider text-zinc-300 transition hover:border-white/30 hover:text-white"
                    onClick={() => {
                      setMode('forgot')
                      setError(null)
                      setInfo(null)
                    }}
                  >
                    Забыли пароль?
                  </button>
                )}
                {(mode === 'forgot' || mode === 'reset') && (
                  <button
                    type="button"
                    className="mt-3 h-11 w-full rounded-2xl border border-white/10 bg-white/5 text-xs font-semibold uppercase tracking-wider text-zinc-300 transition hover:border-white/30 hover:text-white"
                    onClick={() => setMode('login')}
                  >
                    Вернуться ко входу
                  </button>
                )}

                {!token && <VKLoginButton />}
              </div>

              <div className="glass-panel rounded-3xl p-6 text-sm leading-relaxed text-zinc-400 shadow-xl">
                <div className="text-xs font-semibold uppercase tracking-widest text-white">Зачем нужен кабинет?</div>
                <div className="mt-3">— История записей</div>
                <div>— Бонусная система</div>
                <div>— Памятка по уходу (PDF)</div>
                <div className="mt-5 rounded-2xl border border-white/10 bg-white/[0.03] p-4">
                  Не хотите регистрироваться сейчас? Можно задать вопрос в чате — и мы подскажем, с чего начать.
                </div>
              </div>
            </div>
          ) : (
            <div className="mt-6 grid gap-6 md:grid-cols-2">
              <div className="rounded-3xl border border-black/10 bg-white p-6">
                <div className="flex items-start justify-between gap-4">
                  <div>
                    <div className="text-sm text-black/60">Здравствуйте</div>
                    <div className="mt-1 text-xl font-semibold">{user?.name || 'Пользователь'}</div>
                    <div className="mt-2 text-sm text-black/60">{user?.email}</div>
                  </div>
                  <button
                    className="rounded-2xl border border-black/15 px-4 py-2 text-sm font-semibold text-black/70 hover:border-black/30"
                    onClick={logout}
                  >
                    Выйти
                  </button>
                </div>

                <div className="mt-6 grid grid-cols-2 gap-3">
                  <div className="rounded-3xl border border-black/10 p-4">
                    <div className="text-xs text-black/50">Бонусы</div>
                    <div className="mt-1 text-lg font-semibold">{bonusBalance == null ? '—' : bonusBalance}</div>
                  </div>
                  <div className="rounded-3xl border border-black/10 p-4">
                    <div className="text-xs text-black/50">Телефон</div>
                    <div className="mt-1 text-sm font-semibold">{user?.phone || '—'}</div>
                    <div className="mt-1 text-xs text-black/50">
                      {user?.phone_verified ? 'подтверждён' : 'нужен код'}
                    </div>
                  </div>
                  <a
                    className="rounded-3xl border border-black/10 p-4 hover:border-black/25"
                    href={apiBase + '/pdf/care-memo'}
                    target="_blank"
                    rel="noreferrer"
                  >
                    <div className="text-xs text-black/50">Памятка</div>
                    <div className="mt-1 text-lg font-semibold">PDF</div>
                  </a>
                </div>

                <div className="mt-6 rounded-3xl border border-black/10 p-4">
                  <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                    <div>
                      <div className="text-sm font-semibold">Telegram-уведомления</div>
                      <div className="mt-1 text-sm text-black/60">
                        {telegramLinked || user?.telegram_chat_id
                          ? 'Telegram подключён.'
                          : 'Подключите Telegram, чтобы получать статус записи и напоминания.'}
                      </div>
                    </div>
                    <button
                      type="button"
                      className="h-11 rounded-2xl bg-black px-5 text-sm font-semibold text-white disabled:opacity-40"
                      disabled={telegramLinked || Boolean(user?.telegram_chat_id)}
                      onClick={async () => {
                        setError(null)
                        setInfo(null)
                        try {
                          const resp = await apiPost<{ linked: boolean; url: string }>('/telegram/link')
                          setTelegramLinked(Boolean(resp.linked))
                          if (resp.url) {
                            setTelegramLinkURL(resp.url)
                            window.open(resp.url, '_blank', 'noopener,noreferrer')
                          }
                          if (resp.linked) setInfo('Telegram уже подключён.')
                        } catch (e) {
                          setError((e as Error).message)
                        }
                      }}
                    >
                      {telegramLinked || user?.telegram_chat_id ? 'Подключён' : 'Подключить'}
                    </button>
                  </div>
                  {telegramLinkURL && (
                    <a
                      className="mt-3 block break-all text-sm font-semibold text-black underline"
                      href={telegramLinkURL}
                      target="_blank"
                      rel="noreferrer"
                    >
                      Открыть ссылку подключения
                    </a>
                  )}
                </div>

                {!user?.phone_verified && (
                  <div className="mt-6 rounded-3xl border border-amber-200 bg-amber-50 p-4">
                    <div className="text-sm font-semibold text-amber-950">Подтвердите телефон</div>
                    <div className="mt-1 text-sm text-amber-900">
                      Для записи нужен подтверждённый российский номер. Код в dev-режиме появится в логах backend.
                    </div>
                    <div className="mt-4 flex flex-col gap-3 sm:flex-row">
                      <input
                        className="h-11 flex-1 rounded-2xl border border-black/10 bg-white px-3 text-sm outline-none focus:border-black/30"
                        value={accountPhone}
                        onChange={(e) => setAccountPhone(e.target.value)}
                        placeholder="+7 900 000-00-00"
                      />
                      <button
                        type="button"
                        className="h-11 rounded-2xl border border-black/15 bg-white px-5 text-sm font-semibold text-black/70 hover:border-black/30"
                        onClick={async () => {
                          setError(null)
                          setInfo(null)
                          try {
                            const resp = await apiPost<{ user: User }>('/auth/phone/update', { phone: accountPhone })
                            if (token) setAuth(token, resp.user)
                            setInfo('Телефон сохранён. Теперь можно отправить код.')
                          } catch (e) {
                            setError((e as Error).message)
                          }
                        }}
                      >
                        Сохранить
                      </button>
                    </div>
                    <div className="mt-3 flex flex-col gap-3 sm:flex-row">
                      <button
                        type="button"
                        className="h-11 rounded-2xl bg-black px-5 text-sm font-semibold text-white"
                        onClick={async () => {
                          setError(null)
                          setInfo(null)
                          try {
                            await apiPost<{ ok: boolean; phone: string }>('/auth/phone/send-code')
                            setInfo('SMS-код отправлен.')
                          } catch (e) {
                            setError((e as Error).message)
                          }
                        }}
                      >
                        Отправить код
                      </button>
                      <input
                        className="h-11 flex-1 rounded-2xl border border-black/10 bg-white px-3 text-sm outline-none focus:border-black/30"
                        value={phoneCode}
                        onChange={(e) => setPhoneCode(e.target.value)}
                        placeholder="Код из SMS"
                      />
                      <button
                        type="button"
                        className="h-11 rounded-2xl border border-black/15 bg-white px-5 text-sm font-semibold text-black/70 hover:border-black/30"
                        onClick={async () => {
                          setError(null)
                          setInfo(null)
                          try {
                            const resp = await apiPost<{ user: User }>('/auth/phone/verify', { code: phoneCode })
                            if (token) setAuth(token, resp.user)
                            setPhoneCode('')
                            setInfo('Телефон подтверждён.')
                          } catch (e) {
                            setError((e as Error).message)
                          }
                        }}
                      >
                        Подтвердить
                      </button>
                    </div>
                  </div>
                )}

                <div className="mt-6">
                  <div className="text-sm font-semibold">Мои записи</div>
                  <div className="mt-3 space-y-2">
                    {bookings.map((b) => (
                      <div key={b.id} className="rounded-2xl border border-black/10 p-4 text-sm">
                        <div className="font-semibold text-black">
                          {procedures[b.procedure_id] || `Процедура #${b.procedure_id}`}
                        </div>
                        <div className="mt-1 text-black/70">
                          {formatDateTime(b.datetime)}
                        </div>
                        <div className="mt-1 flex items-center gap-2">
                          <span className="text-xs text-black/50">Статус:</span>
                          <span className={[
                            'text-xs font-semibold',
                            b.status === 'confirmed' ? 'text-green-700' :
                            b.status === 'completed' ? 'text-blue-700' :
                            b.status === 'cancelled' ? 'text-red-700' :
                            'text-amber-700'
                          ].join(' ')}>
                            {statusTranslations[b.status] || b.status}
                          </span>
                        </div>
                      </div>
                    ))}
                    {!bookings.length && (
                      <div className="rounded-2xl border border-black/10 p-4 text-sm text-black/70">
                        Пока нет записей. Создайте первую на странице «Запись».
                      </div>
                    )}
                  </div>
                </div>
              </div>

              <div className="rounded-3xl border border-black/10 bg-white p-6">
                <div className="text-sm font-semibold">История бонусов</div>
                <div className="mt-3 space-y-2">
                  {bonusLogs.map((l) => (
                    <div key={l.id} className="rounded-2xl border border-black/10 p-4 text-sm">
                      <div className="flex items-center justify-between">
                        <div className="font-semibold text-black">
                          {bonusTypeTranslations[l.type] || l.type}
                        </div>
                        <div className={[
                          'font-semibold text-lg',
                          l.type === 'spend' ? 'text-red-600' : 'text-green-600'
                        ].join(' ')}>
                          {l.type === 'spend' ? '-' : '+'}{l.amount}
                        </div>
                      </div>
                      {l.comment && <div className="mt-1 text-xs text-black/60">{l.comment}</div>}
                      <div className="mt-1 text-xs text-black/50">{formatDateTime(l.created_at)}</div>
                    </div>
                  ))}
                  {!bonusLogs.length && (
                    <div className="rounded-2xl border border-black/10 p-4 text-sm text-black/70">
                      Пока нет операций.
                    </div>
                  )}
                </div>
              </div>
            </div>
          )}
        </Container>
      </section>
    </MotionPage>
  )
}
