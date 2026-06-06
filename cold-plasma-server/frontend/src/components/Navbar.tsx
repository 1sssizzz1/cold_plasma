import { useEffect, useState } from 'react'
import { Link, NavLink, useLocation } from 'react-router-dom'
import Container from './Container'
import { useAuthStore } from '../store/auth'

const nav = [
  { to: '/', label: 'Главная' },
  { to: '/procedures', label: 'Процедуры' },
  { to: '/booking', label: 'Запись' },
  { to: '/before-after', label: 'До/После' },
  { to: '/account', label: 'Кабинет' },
  { to: '/help', label: 'Помощь' },
]

export default function Navbar() {
  const token = useAuthStore((s) => s.token)
  const user = useAuthStore((s) => s.user)
  const logout = useAuthStore((s) => s.logout)
  const location = useLocation()
  const [open, setOpen] = useState(false)
  const visibleNav = user?.is_admin ? [...nav, { to: '/admin', label: 'Админ' }] : nav

  useEffect(() => {
    setOpen(false)
  }, [location.pathname])

  return (
    <header className="fixed inset-x-0 top-0 z-40 border-b border-black/10 bg-white/80 backdrop-blur-md transition-all duration-300">
      <Container>
        <div className="relative flex h-20 items-center justify-between gap-3">
          <Link to="/" className="group flex min-w-0 items-center gap-3.5">
            <div className="relative h-11 w-11 shrink-0 overflow-hidden rounded-full border border-black/10 bg-white p-[1px] transition-transform duration-500 group-hover:scale-105">
              <img src="/plasma-glow.jpg" alt="" className="h-full w-full rounded-full object-cover opacity-90 transition-opacity group-hover:opacity-100" />
              <div className="absolute inset-0 rounded-full ring-1 ring-black/10 group-hover:ring-black/25 transition-all" />
            </div>
            <div className="min-w-0 leading-tight">
              <div className="truncate text-sm font-semibold tracking-[0.12em] text-black uppercase sm:text-base">Холодная плазма</div>
              <div className="text-[10px] tracking-wider text-zinc-600 uppercase">Северодвинск</div>
            </div>
          </Link>

          <nav className="hidden items-center gap-1.5 md:flex">
            {visibleNav.map((n) => (
              <NavLink
                key={n.to}
                to={n.to}
                className={({ isActive }) =>
                  [
                    'relative rounded-full px-4 py-1.5 text-xs font-medium tracking-wide uppercase transition-all duration-300',
                    isActive
                      ? 'bg-black text-white font-semibold shadow-sm'
                      : 'text-zinc-500 hover:bg-black/5 hover:text-black',
                  ].join(' ')
                }
              >
                {n.label}
              </NavLink>
            ))}
          </nav>

          <div className="hidden md:block">
            {token ? (
              <button
                type="button"
                onClick={logout}
                className="rounded-full border border-black/10 bg-black px-5 py-2 text-xs font-semibold uppercase tracking-wider text-white transition hover:bg-zinc-800 shadow-sm"
              >
                Выйти
              </button>
            ) : (
              <Link
                to="/booking"
                className="rounded-full bg-black px-5 py-2 text-xs font-semibold uppercase tracking-wider text-white transition hover:bg-zinc-800 shadow-sm"
              >
                Записаться
              </Link>
            )}
          </div>

          <button
            type="button"
            aria-expanded={open}
            aria-controls="mobile-navigation"
            onClick={() => setOpen((value) => !value)}
            className="inline-flex h-11 shrink-0 items-center gap-2 rounded-full border border-black/10 bg-black/5 px-4 text-xs font-semibold uppercase tracking-wider text-black hover:bg-black/10 transition md:hidden"
          >
            <span className="grid gap-1" aria-hidden="true">
              <span className={`block h-0.5 w-4 rounded-full bg-black transition-transform ${open ? 'translate-y-1.5 rotate-45' : ''}`} />
              <span className={`block h-0.5 w-4 rounded-full bg-black transition-opacity ${open ? 'opacity-0' : ''}`} />
              <span className={`block h-0.5 w-4 rounded-full bg-black transition-transform ${open ? '-translate-y-1.5 -rotate-45' : ''}`} />
            </span>
            Меню
          </button>

          {open && (
            <div
              id="mobile-navigation"
              className="absolute right-0 top-[calc(100%+0.5rem)] w-[min(20rem,calc(100vw-2rem))] rounded-2xl border border-black/10 bg-white/95 p-3 shadow-2xl backdrop-blur-lg md:hidden"
            >
              <nav className="grid gap-1">
                {visibleNav.map((n) => (
                  <NavLink
                    key={n.to}
                    to={n.to}
                    className={({ isActive }) =>
                      [
                        'rounded-xl px-4 py-2.5 text-xs font-medium tracking-wider uppercase transition-all',
                        isActive ? 'bg-black text-white' : 'text-zinc-500 hover:bg-black/5 hover:text-black',
                      ].join(' ')
                    }
                  >
                    {n.label}
                  </NavLink>
                ))}
              </nav>
              <div className="mt-2 border-t border-black/10 pt-2">
                {token ? (
                  <button
                    type="button"
                    onClick={() => {
                      logout()
                      setOpen(false)
                    }}
                    className="h-10 w-full rounded-xl border border-black/10 bg-black text-xs font-semibold uppercase tracking-wider text-white hover:bg-zinc-800 transition"
                  >
                    Выйти
                  </button>
                ) : (
                  <Link
                    to="/booking"
                    className="flex h-10 w-full items-center justify-center rounded-xl bg-black text-xs font-semibold uppercase tracking-wider text-white hover:bg-zinc-800 transition"
                  >
                    Записаться
                  </Link>
                )}
              </div>
            </div>
          )}
        </div>
      </Container>
    </header>
  )
}
