import { Helmet } from 'react-helmet-async'
import { AnimatePresence, motion } from 'framer-motion'
import { Link } from 'react-router-dom'
import Container from '../components/Container'
import MotionPage from '../components/MotionPage'
import BeforeAfterSlider from '../components/BeforeAfterSlider'
import ProcedureCard from '../components/ProcedureCard'
import { useEffect, useState } from 'react'
import { useProceduresStore } from '../store/procedures'
import { apiGet } from '../utils/api'
import type { BeforeAfterResult, ProcedureReview } from '../types'

type MasterProfile = {
  name: string
  title: string
  bio: string
  photo_url: string
  certificates: string[]
  gallery: string[]
}

export default function HomePage() {
  const fetchAll = useProceduresStore((s) => s.fetchAll)
  const items = useProceduresStore((s) => s.items)
  const [master, setMaster] = useState<MasterProfile | null>(null)
  const [beforeAfterItems, setBeforeAfterItems] = useState<BeforeAfterResult[]>([])
  const [beforeAfterIndex, setBeforeAfterIndex] = useState(0)
  const beforeAfter = beforeAfterItems[beforeAfterIndex] ?? null
  const [reviews, setReviews] = useState<ProcedureReview[]>([])
  const [reviewIndex, setReviewIndex] = useState(0)
  const currentReview = reviews[reviewIndex] ?? null

  useEffect(() => {
    void fetchAll()
    void apiGet<MasterProfile>('/master-profile')
      .then((profile) => setMaster(profile.name || profile.bio || profile.photo_url ? profile : null))
      .catch(() => setMaster(null))
    void apiGet<BeforeAfterResult[]>('/before-after-results?limit=12')
      .then((items) => {
        const list = Array.isArray(items) ? items : []
        const featured = list.filter((item) => item.is_featured)
        const source = featured.length ? featured : list
        setBeforeAfterItems(source)
        setBeforeAfterIndex(source.length ? Math.floor(Math.random() * source.length) : 0)
      })
      .catch(() => {
        setBeforeAfterItems([])
        setBeforeAfterIndex(0)
      })
    void apiGet<ProcedureReview[]>('/reviews?limit=100')
      .then((items) => {
        const list = Array.isArray(items) ? items : []
        if (list.length) {
          // Shuffle reviews for random order
          const shuffled = [...list].sort(() => Math.random() - 0.5)
          setReviews(shuffled)
          setReviewIndex(0)
        }
      })
      .catch(() => setReviews([]))
  }, [fetchAll])

  // Auto-rotate reviews every 60 seconds
  useEffect(() => {
    if (reviews.length <= 1) return
    const interval = setInterval(() => {
      setReviewIndex((current) => (current + 1) % reviews.length)
    }, 60000) // 60 seconds
    return () => clearInterval(interval)
  }, [reviews.length])

  const showPrevBeforeAfter = () => {
    setBeforeAfterIndex((index) => (beforeAfterItems.length ? (index - 1 + beforeAfterItems.length) % beforeAfterItems.length : 0))
  }

  const showNextBeforeAfter = () => {
    setBeforeAfterIndex((index) => (beforeAfterItems.length ? (index + 1) % beforeAfterItems.length : 0))
  }

  const showPrevReview = () => {
    setReviewIndex((index) => (reviews.length ? (index - 1 + reviews.length) % reviews.length : 0))
  }

  const showNextReview = () => {
    setReviewIndex((index) => (reviews.length ? (index + 1) % reviews.length : 0))
  }

  function stars(rating: number) {
    return '★★★★★'.slice(0, rating) + '☆☆☆☆☆'.slice(0, Math.max(0, 5 - rating))
  }

  return (
    <MotionPage>
      <Helmet>
        <title>Холодная плазма — Северодвинск</title>
        <meta
          name="description"
          content="Холодная плазма в Северодвинске: аккуратный уход, понятные ответы и быстрая онлайн-запись. Бонусы и AI-консультация."
        />
      </Helmet>

      <section className="relative pt-8 pb-12 md:pt-14 md:pb-20">
        <Container>
          <div className="grid items-start gap-14 lg:grid-cols-12">
            <div className="lg:col-span-7">
              <div className="inline-flex items-center gap-2 rounded-full border border-black/10 bg-white/70 px-4 py-1.5 text-xs font-medium tracking-wider text-zinc-300 shadow-sm uppercase">
                <span className="h-1.5 w-1.5 rounded-full bg-black/45 animate-pulse" />
                Стерильный подход • Премиум уход • Онлайн запись
              </div>
              <h1 className="silver-glow mt-6 text-4xl font-light tracking-tight text-white sm:text-5xl md:text-6xl leading-[1.1]">
                Холодная плазма <br />
                <span className="font-semibold text-zinc-400">в Северодвинске</span>
              </h1>
              <p className="mt-6 max-w-xl text-base leading-relaxed text-zinc-400 md:text-lg">
                Инновационная бесконтактная технология омоложения и регенерации кожи. 
                Профессиональный уход без компромиссов, лишнего шума и боли.
              </p>
              <div className="mt-8 flex flex-col gap-4 sm:flex-row">
                <Link
                  to="/booking"
                  className="inline-flex h-12 items-center justify-center rounded-full bg-black px-8 text-xs font-semibold uppercase tracking-wider text-white transition-all duration-300 hover:bg-zinc-800 hover:scale-[1.02] active:scale-[0.98] shadow-[0_16px_36px_rgba(17,24,39,0.18)]"
                >
                  Записаться онлайн
                </Link>
                <Link
                  to="/procedures"
                  className="inline-flex h-12 items-center justify-center rounded-full border border-black/10 bg-white/70 px-8 text-xs font-semibold uppercase tracking-wider text-black transition-all duration-300 hover:bg-white hover:border-black/20 hover:scale-[1.02] active:scale-[0.98] shadow-sm"
                >
                  Смотреть процедуры
                </Link>
              </div>

              <div className="mt-12 grid grid-cols-1 gap-5 sm:grid-cols-3">
                <div className="glass-panel-light rounded-[22px] p-5 transition-all duration-300 hover:-translate-y-0.5 hover:border-black/15">
                  <div className="text-[10px] font-semibold uppercase tracking-wider text-zinc-600">Консультация</div>
                  <div className="mt-2 text-sm font-semibold text-black">Бесплатная консультация</div>
                </div>
                <div className="glass-panel-light rounded-[22px] p-5 transition-all duration-300 hover:-translate-y-0.5 hover:border-black/15">
                  <div className="text-[10px] font-semibold uppercase tracking-wider text-zinc-600">Оповещения о записи</div>
                  <div className="mt-2 text-sm font-semibold text-black">СМС и Telegram оповещения о записи</div>
                </div>
                <div className="glass-panel-light rounded-[22px] p-5 transition-all duration-300 hover:-translate-y-0.5 hover:border-black/15">
                  <div className="text-[10px] font-semibold uppercase tracking-wider text-zinc-600">AI помощник</div>
                  <div className="mt-2 text-sm font-semibold text-black">Ответит на ваши вопросы</div>
                </div>
              </div>
            </div>

            <div className="space-y-7 lg:col-span-5 lg:pl-4">
              <div className="relative px-9 sm:px-10">
                <AnimatePresence mode="wait">
                  <motion.div
                    key={beforeAfter?.id || 'empty'}
                    initial={{ opacity: 0, scale: 0.98 }}
                    animate={{ opacity: 1, scale: 1 }}
                    exit={{ opacity: 0, scale: 0.98 }}
                    transition={{ duration: 0.3, ease: 'easeOut' }}
                  >
                    <BeforeAfterSlider beforeSrc={beforeAfter?.before_url} afterSrc={beforeAfter?.after_url} />
                  </motion.div>
                </AnimatePresence>
                {beforeAfterItems.length > 1 && (
                  <>
                    <button
                      type="button"
                      aria-label="Предыдущий результат"
                      onClick={showPrevBeforeAfter}
                      className="absolute left-0 top-1/2 z-10 flex h-11 w-11 -translate-y-1/2 items-center justify-center rounded-full border border-black/10 bg-white text-2xl font-light leading-none text-black shadow-[0_12px_32px_rgba(17,24,39,0.16)] transition hover:-translate-x-0.5 hover:bg-black hover:text-white"
                    >
                      ‹
                    </button>
                    <button
                      type="button"
                      aria-label="Следующий результат"
                      onClick={showNextBeforeAfter}
                      className="absolute right-0 top-1/2 z-10 flex h-11 w-11 -translate-y-1/2 items-center justify-center rounded-full border border-black/10 bg-white text-2xl font-light leading-none text-black shadow-[0_12px_32px_rgba(17,24,39,0.16)] transition hover:translate-x-0.5 hover:bg-black hover:text-white"
                    >
                      ›
                    </button>
                  </>
                )}
              </div>
              {beforeAfter && (
                <div className="glass-panel-light rounded-[20px] px-4 py-3 text-xs text-zinc-400">
                  <span className="font-semibold text-white uppercase tracking-wider">{beforeAfter.title || beforeAfter.procedure || 'Результат процедуры'}</span>
                  {beforeAfter.description ? `: ${beforeAfter.description}` : ''}
                </div>
              )}
            </div>
          </div>
        </Container>
      </section>

      {master && (
        <section className="py-14">
          <Container>
            <div className="glass-panel rounded-[28px] p-5 md:p-8">
              <div className="grid items-start gap-8 md:grid-cols-[300px_1fr]">
                <div className="relative self-start overflow-hidden rounded-[24px] border border-black/10 bg-white shadow-[0_18px_48px_rgba(17,24,39,0.12)]">
                  {master.photo_url ? (
                    <img src={master.photo_url} alt={master.name || 'Мастер'} className="aspect-[4/5] w-full object-cover transition-transform duration-500 hover:scale-105" />
                  ) : (
                    <div className="aspect-[4/5] flex items-center justify-center bg-zinc-950 text-zinc-600">
                      Нет фото
                    </div>
                  )}
                </div>
                <div className="flex min-w-0 flex-col justify-center py-2 md:py-3">
                  <div className="text-xs font-semibold uppercase tracking-widest text-zinc-500">Специалист студии</div>
                  <div className="mt-2 text-3xl font-semibold tracking-tight text-white">{master.name || 'Специалист студии'}</div>
                  {master.title && <div className="mt-2 text-sm font-medium text-zinc-400">{master.title}</div>}
                  {master.bio && <div className="mt-6 max-w-3xl whitespace-pre-line text-sm leading-7 text-zinc-300">{master.bio}</div>}

                  {!!master.certificates?.length && (
                    <div className="mt-8">
                      <div className="text-xs font-semibold uppercase tracking-widest text-zinc-500">Сертификаты</div>
                      <div className="mt-4 grid grid-cols-2 gap-4 sm:grid-cols-3">
                        {master.certificates.map((url) => (
                          <a key={url} href={url} target="_blank" rel="noreferrer" className="group relative overflow-hidden rounded-[18px] border border-black/10 bg-white shadow-sm">
                            <img src={url} className="aspect-[4/3] w-full object-cover transition-transform duration-500 group-hover:scale-105" />
                          </a>
                        ))}
                      </div>
                    </div>
                )}

                  {!!master.gallery?.length && (
                    <div className="mt-8">
                      <div className="text-xs font-semibold uppercase tracking-widest text-zinc-500">Фотографии</div>
                      <div className="mt-4 grid grid-cols-2 gap-4 sm:grid-cols-3">
                        {master.gallery.map((url) => (
                          <a key={url} href={url} target="_blank" rel="noreferrer" className="group relative overflow-hidden rounded-[18px] border border-black/10 bg-white shadow-sm">
                            <img src={url} className="aspect-square w-full object-cover transition-transform duration-500 group-hover:scale-105" />
                          </a>
                        ))}
                      </div>
                    </div>
                )}
                </div>
              </div>
            </div>
          </Container>
        </section>
      )}

      {reviews.length > 0 && (
        <section className="py-14">
          <Container>
            <div className="flex items-end justify-between gap-4 border-b border-black/10 pb-6">
              <div>
                <div className="text-xs font-semibold uppercase tracking-widest text-zinc-500">Отзывы клиентов</div>
                <div className="mt-2 text-2xl font-semibold tracking-tight text-white sm:text-3xl">Что говорят о нас</div>
              </div>
              <Link to="/procedures" className="text-xs font-semibold uppercase tracking-wider text-zinc-400 hover:text-white transition-colors">
                Все отзывы →
              </Link>
            </div>

            <div className="mt-8 relative">
              <div className="relative px-12 sm:px-16">
                <AnimatePresence mode="wait">
                  <motion.div
                    key={currentReview?.id || 'empty'}
                    initial={{ opacity: 0, y: 20 }}
                    animate={{ opacity: 1, y: 0 }}
                    exit={{ opacity: 0, y: -20 }}
                    transition={{ duration: 0.5, ease: 'easeInOut' }}
                    className="glass-panel rounded-3xl p-8 shadow-xl"
                  >
                    {currentReview ? (
                      <div className="max-w-3xl mx-auto text-center">
                        <div className="text-3xl text-amber-400 mb-4">{stars(currentReview.rating)}</div>
                        <blockquote className="text-lg leading-relaxed text-white font-light italic">
                          "{currentReview.text || 'Отличная процедура, рекомендую!'}"
                        </blockquote>
                        <div className="mt-6 flex flex-col items-center gap-2">
                          <div className="text-sm font-semibold text-white">{currentReview.user_name || 'Клиент'}</div>
                          {currentReview.procedure_title && (
                            <div className="text-xs text-zinc-400">Процедура: {currentReview.procedure_title}</div>
                          )}
                          <div className="text-xs text-zinc-500">
                            {new Date(currentReview.created_at).toLocaleDateString('ru-RU')}
                          </div>
                        </div>
                      </div>
                    ) : (
                      <div className="text-center text-zinc-400">Загружаем отзывы...</div>
                    )}
                  </motion.div>
                </AnimatePresence>

                {reviews.length > 1 && (
                  <>
                    <button
                      type="button"
                      aria-label="Предыдущий отзыв"
                      onClick={showPrevReview}
                      className="absolute left-0 top-1/2 z-10 flex h-12 w-12 -translate-y-1/2 items-center justify-center rounded-full border border-white/10 bg-white/10 backdrop-blur-sm text-2xl font-light leading-none text-white shadow-lg transition hover:-translate-x-0.5 hover:bg-white/20"
                    >
                      ‹
                    </button>
                    <button
                      type="button"
                      aria-label="Следующий отзыв"
                      onClick={showNextReview}
                      className="absolute right-0 top-1/2 z-10 flex h-12 w-12 -translate-y-1/2 items-center justify-center rounded-full border border-white/10 bg-white/10 backdrop-blur-sm text-2xl font-light leading-none text-white shadow-lg transition hover:translate-x-0.5 hover:bg-white/20"
                    >
                      ›
                    </button>
                  </>
                )}
              </div>

              {reviews.length > 1 && (
                <div className="mt-6 flex justify-center gap-2">
                  {reviews.map((_, index) => (
                    <button
                      key={index}
                      type="button"
                      aria-label={`Перейти к отзыву ${index + 1}`}
                      onClick={() => setReviewIndex(index)}
                      className={[
                        'h-2 rounded-full transition-all duration-300',
                        index === reviewIndex ? 'w-8 bg-white' : 'w-2 bg-white/30 hover:bg-white/50',
                      ].join(' ')}
                    />
                  ))}
                </div>
              )}
            </div>
          </Container>
        </section>
      )}

      <section className="py-14">
        <Container>
          <div className="flex items-end justify-between gap-4 border-b border-black/10 pb-6">
            <div>
              <div className="text-xs font-semibold uppercase tracking-widest text-zinc-500">Процедуры</div>
              <div className="mt-2 text-2xl font-semibold tracking-tight text-white sm:text-3xl">Популярные форматы</div>
            </div>
            <Link to="/procedures" className="text-xs font-semibold uppercase tracking-wider text-zinc-400 hover:text-white transition-colors">
              Все процедуры →
            </Link>
          </div>

          <div className="mt-8 grid gap-6 md:grid-cols-2">
            {(items.length ? items.slice(0, 2) : []).map((p) => (
              <ProcedureCard key={p.id} p={p} />
            ))}
            {!items.length && (
              <>
                <div className="glass-panel rounded-3xl p-8 text-center text-sm text-zinc-400">
                  Загружаем процедуры…
                </div>
                <div className="glass-panel rounded-3xl p-8 text-center text-sm text-zinc-400">
                  Если API ещё не запущен — поднимите docker-compose.
                </div>
              </>
            )}
          </div>
        </Container>
      </section>
    </MotionPage>
  )
}
