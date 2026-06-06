import type { PropsWithChildren } from 'react'
import { useEffect } from 'react'
import Navbar from './Navbar'
import Footer from './Footer'
import ChatWidget from './chat/ChatWidget'
import { useAuthStore } from '../store/auth'
import Analytics from './analytics/Analytics'

export default function SiteShell({ children }: PropsWithChildren) {
  const hydrate = useAuthStore((s) => s.hydrate)
  useEffect(() => {
    hydrate()
  }, [hydrate])

  return (
    <div className="relative min-h-dvh overflow-x-hidden bg-[#f7f7f4] text-[#111111] antialiased">
      <div className="pointer-events-none fixed inset-0 z-0 overflow-hidden">
        <div className="absolute inset-0 bg-[#f7f7f4]" />
        <div className="absolute inset-0 bg-[linear-gradient(180deg,rgba(255,255,255,0.62)_0%,rgba(247,247,244,0.08)_34%,rgba(239,239,234,0.32)_100%)]" />
        <div className="mono-lava-background absolute inset-0" />
      </div>

      <div className="relative z-10 flex min-h-dvh flex-col justify-between">
        <Analytics />
        <Navbar />
        <main className="flex-grow pt-24 pb-12">
          {children}
        </main>
        <Footer />
        <ChatWidget />
      </div>
    </div>
  )
}
