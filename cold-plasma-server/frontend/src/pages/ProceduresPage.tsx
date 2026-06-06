import { Helmet } from 'react-helmet-async'
import { useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import Container from '../components/Container'
import MotionPage from '../components/MotionPage'
import { useProceduresStore } from '../store/procedures'
import { apiGet, apiPost } from '../utils/api'
import { useAuthStore } from '../store/auth'
import type { Procedure, ProcedureReview } from '../types'

function rub(n: number) {
  return new Intl.NumberFormat('ru-RU').format(n) + ' ₽'
}

function stars(rating: number) {
  return '★★★★★'.slice(0, rating) + '☆☆☆☆☆'.slice(0, Math.max(0, 5 - rating))
}

export default function ProceduresPage() {
  const fetchAll = useProceduresStore((s) => s.fetchAll)
  const items = useProceduresStore((s) => s.items)
  const loading = useProceduresStore((s) => s.loading)
  const error = useProceduresStore((s) => s.error)
  const token = useAuthStore((s) => s.token)
  const [openedID, setOpenedID] = useState<number | null>(null)
  const [reviews, setReviews] = useState<Record<number, ProcedureReview[]>>({})
  const [ratingByProcedure, setRatingByProcedure] = useState<Record<number, number>>({})
  const [textByProcedure, setTextByProcedure] = useState<Record<number, string>>({})
  const [reviewError, setReviewError] = useState<Record<number, string>>({})
  const [reviewInfo, setReviewInfo] = useState<Record<number, string>>({})

  useEffect(() => {
    void fetchAll()
  }, [fetchAll])

  useEffect(() => {
    for (const item of items) {
      void loadReviews(item.id)
    }
  }, [items])

  const reviewStats = useMemo(() => {
    const out: Record<number, { avg: number; count: number }> = {}
    for (const item of items) {
      const list = reviews[item.id] || []
      const avg = list.length ? list.reduce((sum, review) => sum + review.rating, 0) / list.length : 0
      out[item.id] = { avg, count: list.length }
    }
    return out
  }, [items, reviews])

  const loadReviews = async (procedureID: number) => {
    try {
      const list = await apiGet<ProcedureReview[]>(`/procedures/${procedureID}/reviews`)
      setReviews((prev) => ({ ...prev, [procedureID]: list }))
    } catch {
      setReviews((prev) => ({ ...prev, [procedureID]: [] }))
    }
  }

  const submitReview = async (procedureID: number) => {
    setReviewError((prev) => ({ ...prev, [procedureID]: '' }))
    setReviewInfo((prev) => ({ ...prev, [procedureID]: '' }))
    try {
      await apiPost<ProcedureReview>('/reviews', {
        procedure_id: procedureID,
        rating: ratingByProcedure[procedureID] ?? 5,
        text: textByProcedure[procedureID] ?? '',
      })
      setReviewInfo((prev) => ({ ...prev, [procedureID]: 'Отзыв сохранён.' }))
      setTextByProcedure((prev) => ({ ...prev, [procedureID]: '' }))
      await loadReviews(procedureID)
    } catch (e) {
      setReviewError((prev) => ({ ...prev, [procedureID]: (e as Error).message }))
    }
  }

  return (
    <MotionPage>
      <Helmet>
        <title>Процедуры — Холодная плазма</title>
        <meta name="description" content="Каталог процедур холодной плазмы: цена, длительность, описание и отзывы." />
      </Helmet>

      <section className="py-12">
        <Container>
          <div className="border-b border-white/5 pb-6">
            <div className="text-xs font-semibold uppercase tracking-widest text-zinc-500">Каталог</div>
            <div className="mt-2 text-3xl font-semibold tracking-tight text-white sm:text-4xl">Процедуры</div>
            <div className="mt-3 max-w-2xl text-sm text-zinc-400">
              Раскройте карточку, чтобы посмотреть детали, состав услуги и отзывы.
            </div>
          </div>

          {error && (
            <div className="mt-6 rounded-2xl border border-red-500/20 bg-red-500/10 p-4 text-sm text-red-200">
              {error}
            </div>
          )}

          <div className="mt-8 space-y-5">
            {items.map((p) => (
              <ProcedureProductCard
                key={p.id}
                procedure={p}
                opened={openedID === p.id}
                onToggle={() => setOpenedID(openedID === p.id ? null : p.id)}
                reviews={reviews[p.id] || []}
                stats={reviewStats[p.id] || { avg: 0, count: 0 }}
                token={token}
                rating={ratingByProcedure[p.id] ?? 5}
                text={textByProcedure[p.id] ?? ''}
                info={reviewInfo[p.id] || ''}
                error={reviewError[p.id] || ''}
                onRatingChange={(rating) => setRatingByProcedure((prev) => ({ ...prev, [p.id]: rating }))}
                onTextChange={(text) => setTextByProcedure((prev) => ({ ...prev, [p.id]: text }))}
                onSubmitReview={() => void submitReview(p.id)}
              />
            ))}
            {loading && !items.length && (
              <div className="glass-panel rounded-3xl p-8 text-center text-sm text-zinc-400">
                Загружаем каталог...
              </div>
            )}
          </div>
        </Container>
      </section>
    </MotionPage>
  )
}

type ProductCardProps = {
  procedure: Procedure
  opened: boolean
  onToggle: () => void
  reviews: ProcedureReview[]
  stats: { avg: number; count: number }
  token: string | null
  rating: number
  text: string
  info: string
  error: string
  onRatingChange: (rating: number) => void
  onTextChange: (text: string) => void
  onSubmitReview: () => void
}

function ProcedureProductCard({
  procedure,
  opened,
  onToggle,
  reviews,
  stats,
  token,
  rating,
  text,
  info,
  error,
  onRatingChange,
  onTextChange,
  onSubmitReview,
}: ProductCardProps) {
  const [showAllReviews, setShowAllReviews] = useState(false)
  const services = procedure.services
    .split(/\n|;/)
    .map((item) => item.trim())
    .filter(Boolean)

  const displayedReviews = showAllReviews ? reviews : reviews.slice(0, 1)

  return (
    <article className="glass-panel overflow-hidden rounded-3xl shadow-xl transition-all duration-300 hover:border-white/15">
      <button
        type="button"
        className="grid w-full gap-5 p-5 text-left transition hover:bg-white/[0.03] md:grid-cols-[220px_1fr_auto]"
        aria-expanded={opened}
        onClick={onToggle}
      >
        <div className="aspect-video overflow-hidden rounded-2xl border border-white/10 bg-zinc-950">
          {procedure.image_url ? (
            <img src={procedure.image_url} alt={procedure.title} className="h-full w-full object-cover transition-transform duration-500 hover:scale-105" />
          ) : (
            <div className="flex h-full items-center justify-center text-xs text-zinc-600">Нет изображения</div>
          )}
        </div>
        <div className="min-w-0">
          <div className="flex flex-wrap items-center gap-2">
            <span className="text-[10px] font-semibold uppercase tracking-widest text-zinc-500">{procedure.category || 'Процедура'}</span>
            {procedure.popular && (
              <span className="rounded-full border border-white/20 bg-white px-2.5 py-1 text-[10px] font-bold uppercase tracking-wider text-black">Популярно</span>
            )}
          </div>
          <div className="mt-1.5 text-xl font-semibold tracking-tight text-white">{procedure.title}</div>
          <div className="mt-2 line-clamp-2 text-sm leading-relaxed text-zinc-400">{procedure.description}</div>
          <div className="mt-4 flex flex-wrap gap-2 text-xs text-zinc-300">
            <span className="rounded-2xl border border-white/10 bg-white/[0.03] px-3 py-1.5 font-semibold text-white">{rub(procedure.price)}</span>
            <span className="rounded-2xl border border-white/10 bg-white/[0.03] px-3 py-1.5">
              {procedure.duration_str || `${procedure.duration_mins} мин`}
            </span>
            <span className="rounded-2xl border border-white/10 bg-white/[0.03] px-3 py-1.5">
              {stats.count ? `${stats.avg.toFixed(1)}/5 · ${stats.count}` : 'Нет отзывов'}
            </span>
          </div>
        </div>
        <div className="flex items-center justify-end text-2xl font-light leading-none text-zinc-500">{opened ? '−' : '+'}</div>
      </button>

      {opened && (
        <div className="border-t border-white/5 p-5">
          <div className="grid gap-5 lg:grid-cols-[minmax(0,1fr)_360px]">
            <div>
              <div className="text-xs font-semibold uppercase tracking-widest text-zinc-500">О процедуре</div>
              <div className="mt-3 whitespace-pre-line text-sm leading-relaxed text-zinc-300">{procedure.description}</div>

              {services.length > 0 && (
                <div className="mt-5">
                  <div className="text-xs font-semibold uppercase tracking-widest text-zinc-500">Что входит</div>
                  <div className="mt-2 grid gap-2 sm:grid-cols-2">
                    {services.map((item) => (
                      <div key={item} className="rounded-2xl border border-white/10 bg-white/[0.03] px-3 py-2 text-sm text-zinc-300">
                        {item}
                      </div>
                    ))}
                  </div>
                </div>
              )}

              <div className="mt-5 grid gap-3 sm:grid-cols-3">
                <div className="rounded-2xl border border-white/10 bg-white/[0.03] p-4">
                  <div className="text-[10px] font-semibold uppercase tracking-wider text-zinc-500">Цена</div>
                  <div className="mt-1 text-sm font-semibold text-white">{rub(procedure.price)}</div>
                </div>
                <div className="rounded-2xl border border-white/10 bg-white/[0.03] p-4">
                  <div className="text-[10px] font-semibold uppercase tracking-wider text-zinc-500">Длительность</div>
                  <div className="mt-1 text-sm font-semibold text-white">
                    {procedure.duration_str || `${procedure.duration_mins} мин`}
                  </div>
                </div>
                <div className="rounded-2xl border border-white/10 bg-white/[0.03] p-4">
                  <div className="text-[10px] font-semibold uppercase tracking-wider text-zinc-500">Отзывы</div>
                  <div className="mt-1 text-sm font-semibold text-white">
                    {stats.count ? `${stars(Math.round(stats.avg))} ${stats.avg.toFixed(1)}` : 'Пока нет'}
                  </div>
                </div>
              </div>

              <Link
                to="/booking"
                className="mt-5 inline-flex h-11 items-center justify-center rounded-2xl bg-black px-5 text-xs font-semibold uppercase tracking-wider text-white transition hover:bg-zinc-800 shadow-sm"
              >
                Записаться
              </Link>

              {/* Отзывы под описанием на десктопе */}
              <div className="mt-8">
                <div className="flex items-center justify-between gap-3 mb-4">
                  <div className="text-xs font-semibold uppercase tracking-widest text-zinc-500">Отзывы клиентов</div>
                  <div className="rounded-full border border-white/10 bg-white/[0.03] px-2.5 py-1 text-[10px] font-semibold uppercase tracking-wider text-zinc-500">
                    {reviews.length}
                  </div>
                </div>

                <div className="space-y-3">
                  {displayedReviews.map((review) => (
                    <div key={review.id} className="rounded-[20px] border border-white/10 bg-white/[0.03] p-4 text-sm shadow-sm">
                      <div className="font-semibold text-white">{review.user_name || 'Клиент'}</div>
                      <div className="mt-1 text-xs text-zinc-400">{stars(review.rating)} · {review.rating}/5</div>
                      <div className="mt-2 text-zinc-300">{review.text || 'Без текста'}</div>
                    </div>
                  ))}
                  {!reviews.length && (
                    <div className="rounded-[20px] border border-dashed border-white/20 bg-white/[0.02] p-4 text-sm text-zinc-400">
                      Отзывов пока нет. Станьте первым!
                    </div>
                  )}
                  {reviews.length > 1 && !showAllReviews && (
                    <button
                      type="button"
                      onClick={() => setShowAllReviews(true)}
                      className="w-full h-10 rounded-2xl border border-white/10 bg-white/5 text-xs font-semibold uppercase tracking-wider text-white transition hover:bg-white/10"
                    >
                      Смотреть все отзывы ({reviews.length})
                    </button>
                  )}
                  {showAllReviews && reviews.length > 1 && (
                    <button
                      type="button"
                      onClick={() => setShowAllReviews(false)}
                      className="w-full h-10 rounded-2xl border border-white/10 bg-white/5 text-xs font-semibold uppercase tracking-wider text-white transition hover:bg-white/10"
                    >
                      Свернуть отзывы
                    </button>
                  )}
                </div>
              </div>
            </div>

            {/* Форма отзыва справа */}
            <aside className="rounded-[26px] border border-black/10 bg-white/75 p-5 shadow-[0_18px_52px_rgba(17,24,39,0.08)] backdrop-blur-md">
              <div className="text-sm font-semibold text-black mb-1">Оставить отзыв</div>
              <div className="text-xs text-zinc-600">Доступно после завершённого сеанса.</div>
              {error && <div className="mt-3 rounded-2xl border border-red-200 bg-red-50 p-3 text-sm text-red-700">{error}</div>}
              {info && <div className="mt-3 rounded-2xl border border-emerald-200 bg-emerald-50 p-3 text-sm text-emerald-800">{info}</div>}

              <div className="mt-4">
                <div className="text-xs font-semibold text-zinc-600 mb-2">Ваша оценка</div>
                <div className="flex items-center gap-1">
                  {[1, 2, 3, 4, 5].map((star) => (
                    <button
                      key={star}
                      type="button"
                      onClick={() => onRatingChange(star)}
                      className="text-3xl transition-all hover:scale-110 focus:outline-none"
                      aria-label={`Оценка ${star} из 5`}
                    >
                      <span className={star <= rating ? 'text-amber-500' : 'text-zinc-300'}>★</span>
                    </button>
                  ))}
                  <span className="ml-2 text-sm text-zinc-600">{rating}/5</span>
                </div>
              </div>

              <textarea
                className="mt-3 min-h-24 w-full resize-y rounded-2xl border border-black/10 bg-white p-3 text-sm text-black outline-none placeholder:text-zinc-500 transition focus:border-black/35"
                value={text}
                onChange={(e) => onTextChange(e.target.value)}
                placeholder="Расскажите о своих впечатлениях..."
              />
              <button
                type="button"
                disabled={!token}
                className="mt-3 h-11 w-full rounded-2xl bg-black px-4 text-xs font-semibold uppercase tracking-wider text-white transition hover:bg-zinc-800 disabled:cursor-not-allowed disabled:bg-black/20 disabled:text-zinc-500"
                onClick={onSubmitReview}
              >
                {token ? 'Отправить отзыв' : 'Войдите, чтобы оставить отзыв'}
              </button>
            </aside>
          </div>
        </div>
      )}
    </article>
  )
}
