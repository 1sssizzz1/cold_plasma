import type { PropsWithChildren } from 'react'
import { useEffect, useState } from 'react'
import Container from '../Container'
import MotionPage from '../MotionPage'
import { useAuthStore } from '../../store/auth'
import { apiGet } from '../../utils/api'
import type { User } from '../../types'
import NotFoundPage from '../../pages/NotFoundPage'

export default function RequireAdmin({ children }: PropsWithChildren) {
  const token = useAuthStore((s) => s.token)
  const user = useAuthStore((s) => s.user)
  const hydrated = useAuthStore((s) => s.hydrated)
  const setAuth = useAuthStore((s) => s.setAuth)
  const logout = useAuthStore((s) => s.logout)
  const [checking, setChecking] = useState(false)
  const [checked, setChecked] = useState(false)

  useEffect(() => {
    if (!hydrated || !token || checked) return

    let cancelled = false
    setChecking(true)

    void apiGet<User>('/me')
      .then((freshUser) => {
        if (!cancelled) setAuth(token, freshUser)
      })
      .catch(() => {
        if (!cancelled) logout()
      })
      .finally(() => {
        if (!cancelled) {
          setChecking(false)
          setChecked(true)
        }
      })

    return () => {
      cancelled = true
    }
  }, [checked, hydrated, logout, setAuth, token])

  if (!hydrated || (token && (!checked || checking))) {
    return (
      <MotionPage>
        <section className="py-10">
          <Container>
            <div className="rounded-3xl border border-black/10 bg-white p-6 text-sm text-black/60">
              Загружаем страницу…
            </div>
          </Container>
        </section>
      </MotionPage>
    )
  }

  if (!token || !user?.is_admin) {
    return <NotFoundPage />
  }

  return <>{children}</>
}
