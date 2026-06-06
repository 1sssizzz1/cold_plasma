import { Link } from 'react-router-dom'
import type { Procedure } from '../types'

function rub(n: number) {
  return new Intl.NumberFormat('ru-RU').format(n) + ' ₽'
}

export default function ProcedureCard({ p }: { p: Procedure }) {
  return (
    <div className="glass-panel rounded-3xl p-6 shadow-xl transition-all duration-300 hover:border-white/15 hover:scale-[1.01]">
      {p.image_url && (
        <div className="relative mb-5 overflow-hidden rounded-2xl border border-white/10 bg-zinc-950">
          <img src={p.image_url} alt={p.title} className="aspect-video w-full object-cover transition-transform duration-500 hover:scale-105" />
        </div>
      )}
      <div className="flex items-start justify-between gap-4">
        <div>
          <div className="text-[10px] font-semibold uppercase tracking-widest text-zinc-500">{p.category || 'Процедура'}</div>
          <div className="mt-1.5 text-xl font-semibold tracking-tight text-white">{p.title}</div>
        </div>
        {p.popular && (
          <div className="rounded-full border border-white/20 bg-white px-3 py-1 text-[10px] font-bold uppercase tracking-wider text-black shadow-[0_0_15px_rgba(255,255,255,0.15)]">
            Популярно
          </div>
        )}
      </div>

      <div className="mt-5 grid grid-cols-3 gap-3 text-sm">
        <div className="rounded-2xl border border-white/5 bg-white/[0.02] px-4 py-3">
          <div className="text-[10px] font-semibold uppercase tracking-wider text-zinc-500">Цена</div>
          <div className="mt-1 font-semibold text-white">{rub(p.price)}</div>
        </div>
        <div className="rounded-2xl border border-white/5 bg-white/[0.02] px-4 py-3">
          <div className="text-[10px] font-semibold uppercase tracking-wider text-zinc-500">Время</div>
          <div className="mt-1 font-semibold text-white">{p.duration_str || `${p.duration_mins} мин`}</div>
        </div>
        <div className="rounded-2xl border border-white/5 bg-white/[0.02] px-4 py-3">
          <div className="text-[10px] font-semibold uppercase tracking-wider text-zinc-500">Бонусы</div>
          <div className="mt-1 font-semibold text-white">+{p.bonus_earned}</div>
        </div>
      </div>

      <div className="mt-4 text-xs leading-relaxed text-zinc-400">{p.description}</div>

      <div className="mt-6 flex gap-3">
        <Link
          to="/booking"
          className="inline-flex h-11 flex-1 items-center justify-center rounded-2xl bg-black text-xs font-semibold uppercase tracking-wider text-white transition-all duration-300 hover:bg-zinc-800 hover:scale-[1.02] active:scale-[0.98] shadow-sm"
        >
          Записаться
        </Link>
        <Link
          to="/procedures"
          className="inline-flex h-11 flex-1 items-center justify-center rounded-2xl border border-white/20 bg-white/10 text-xs font-semibold uppercase tracking-wider text-white transition-all duration-300 hover:bg-white/15 hover:border-white/30 hover:scale-[1.02] active:scale-[0.98]"
        >
          Подробнее
        </Link>
      </div>
    </div>
  )
}
