import { useState } from 'react'
import { Auth, Config, ConfigResponseMode, type AuthResponse } from '@vkid/sdk'
import { apiPost } from '../../utils/api'
import { useAuthStore } from '../../store/auth'
import type { User } from '../../types'

function randomVerifier() {
  const bytes = new Uint8Array(32)
  crypto.getRandomValues(bytes)
  return Array.from(bytes, (b) => b.toString(16).padStart(2, '0')).join('')
}

export default function VKLoginButton() {
  const setAuth = useAuthStore((s) => s.setAuth)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const app = Number(import.meta.env.VITE_VK_CLIENT_ID || 0)
  const redirectUrl = (import.meta.env.VITE_VK_REDIRECT_URI as string | undefined) || `${window.location.origin}/account`

  if (!app) return null

  return (
    <div className="mt-4">
      {error && (
        <div className="mb-3 rounded-2xl border border-red-200 bg-red-50 p-3 text-sm text-red-700">
          {error}
        </div>
      )}
      <button
        type="button"
        disabled={loading}
        className="h-11 w-full rounded-2xl border border-[#2787f5] bg-[#2787f5] text-sm font-semibold text-white disabled:opacity-40"
        onClick={async () => {
          setError(null)
          setLoading(true)
          const codeVerifier = randomVerifier()
          try {
            Config.init({
              app,
              redirectUrl,
              scope: 'phone email',
              responseMode: ConfigResponseMode.Callback,
              codeVerifier,
            })
            const vkResp = (await Auth.login()) as AuthResponse
            const resp = await apiPost<{ user: User; token: string }>('/auth/vk/exchange', {
              code: vkResp.code,
              code_verifier: codeVerifier,
              device_id: vkResp.device_id,
              redirect_uri: redirectUrl,
            })
            setAuth(resp.token, resp.user)
          } catch (e) {
            setError((e as Error).message || 'Не удалось войти через VK')
          } finally {
            setLoading(false)
          }
        }}
      >
        {loading ? 'Открываем VK…' : 'Войти через VK'}
      </button>
    </div>
  )
}
