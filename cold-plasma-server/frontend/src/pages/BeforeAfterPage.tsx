import { useEffect, useState } from 'react'
import { Helmet } from 'react-helmet-async'
import Container from '../components/Container'
import MotionPage from '../components/MotionPage'
import BeforeAfterSlider from '../components/BeforeAfterSlider'
import { apiGet } from '../utils/api'
import type { BeforeAfterResult } from '../types'

export default function BeforeAfterPage() {
  const [items, setItems] = useState<BeforeAfterResult[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    setLoading(true)
    void apiGet<BeforeAfterResult[]>('/before-after-results?limit=100')
      .then((data) => setItems(Array.isArray(data) ? data : []))
      .catch(() => setItems([]))
      .finally(() => setLoading(false))
  }, [])

  return (
    <MotionPage>
      <Helmet>
        <title>До/После — Холодная плазма (Северодвинск)</title>
        <meta name="description" content="Реальные фото до и после процедур холодной плазмы." />
      </Helmet>

      <section className="py-12">
        <Container>
          <div className="border-b border-white/5 pb-6">
            <div className="text-xs font-semibold uppercase tracking-widest text-zinc-500">Результаты</div>
            <div className="mt-2 text-3xl font-semibold tracking-tight text-white sm:text-4xl">До / После</div>
            <div className="mt-3 max-w-2xl text-sm text-zinc-400">
              Результаты процедур с аккуратным сравнением. Фото добавляются администратором и обновляются автоматически.
            </div>
          </div>

          <div className="mt-10 grid gap-8 md:grid-cols-2">
            {items.map((item) => (
              <article key={item.id} className="space-y-4">
                <BeforeAfterSlider beforeSrc={item.before_url} afterSrc={item.after_url} />
                <div className="glass-panel rounded-3xl p-6 shadow-xl">
                  <div className="text-[10px] font-semibold uppercase tracking-widest text-zinc-500">{item.procedure || 'Процедура'}</div>
                  <div className="mt-2 text-lg font-semibold text-white">{item.title || 'Результат процедуры'}</div>
                  {item.description && <div className="mt-3 whitespace-pre-line text-xs leading-relaxed text-zinc-400">{item.description}</div>}
                </div>
              </article>
            ))}
          </div>

          {!loading && !items.length && (
            <div className="mt-10 glass-panel rounded-3xl p-12 text-center text-sm text-zinc-500">
              Результаты скоро появятся.
            </div>
          )}
        </Container>
      </section>
    </MotionPage>
  )
}
