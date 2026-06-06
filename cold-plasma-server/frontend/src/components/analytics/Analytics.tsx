import { useEffect } from 'react'

function loadScriptOnce(id: string, src: string) {
  if (document.getElementById(id)) return
  const s = document.createElement('script')
  s.id = id
  s.async = true
  s.src = src
  document.head.appendChild(s)
}

export default function Analytics() {
  useEffect(() => {
    const ymId = (import.meta.env.VITE_YM_ID as string | undefined)?.trim()
    if (ymId) {
      // Яндекс.Метрика (облегчённая загрузка). В проде можно заменить на официальный сниппет.
      loadScriptOnce('ym-script', `https://mc.yandex.ru/metrika/tag.js`)
      ;(window as any).ym =
        (window as any).ym ||
        function (...args: any[]) {
          ;((window as any).ym.a = (window as any).ym.a || []).push(args)
        }
      ;(window as any).ym.l = +new Date()
      ;(window as any).ym(Number(ymId), 'init', { clickmap: true, trackLinks: true, accurateTrackBounce: true })
    }

    const gaId = (import.meta.env.VITE_GA_ID as string | undefined)?.trim()
    if (gaId) {
      loadScriptOnce('ga-gtag', `https://www.googletagmanager.com/gtag/js?id=${encodeURIComponent(gaId)}`)
      ;(window as any).dataLayer = (window as any).dataLayer || []
      function gtag(...args: any[]) {
        ;(window as any).dataLayer.push(args)
      }
      gtag('js', new Date())
      gtag('config', gaId)
    }
  }, [])

  return null
}

