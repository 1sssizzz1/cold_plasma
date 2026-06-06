import { useCallback, useMemo, useRef, useState } from 'react'

function svgData(label: string) {
  const svg = encodeURIComponent(
    `<svg xmlns="http://www.w3.org/2000/svg" width="1200" height="1500">
      <rect width="1200" height="1500" fill="#09090b"/>
      <rect x="90" y="90" width="1020" height="1320" rx="44" fill="#050505" stroke="#ffffff" stroke-opacity="0.08" stroke-width="6"/>
      <text x="600" y="780" font-family="Arial" font-size="72" fill="#ffffff" fill-opacity="0.7" text-anchor="middle">${label}</text>
    </svg>`,
  )
  return `data:image/svg+xml,${svg}`
}

function clamp(value: number) {
  return Math.min(100, Math.max(0, value))
}

export default function BeforeAfterSlider({
  beforeSrc,
  afterSrc,
}: {
  beforeSrc?: string
  afterSrc?: string
}) {
  const [v, setV] = useState(50)
  const frameRef = useRef<HTMLDivElement | null>(null)
  const before = useMemo(() => beforeSrc || svgData('До'), [beforeSrc])
  const after = useMemo(() => afterSrc || svgData('После'), [afterSrc])

  const setFromClientX = useCallback((clientX: number) => {
    const rect = frameRef.current?.getBoundingClientRect()
    if (!rect || rect.width <= 0) return
    setV(clamp(((clientX - rect.left) / rect.width) * 100))
  }, [])

  return (
    <div className="glass-panel rounded-[28px] p-4 shadow-[0_24px_70px_rgba(17,24,39,0.16)]">
      <div
        ref={frameRef}
        role="slider"
        aria-label="Сравнение до и после"
        aria-valuemin={0}
        aria-valuemax={100}
        aria-valuenow={Math.round(v)}
        tabIndex={0}
        className="relative aspect-[4/5] touch-none overflow-hidden rounded-[22px] bg-zinc-950 outline-none ring-white/10 focus:ring-2"
        onPointerDown={(e) => {
          e.currentTarget.setPointerCapture(e.pointerId)
          setFromClientX(e.clientX)
        }}
        onPointerMove={(e) => {
          if (e.currentTarget.hasPointerCapture(e.pointerId)) setFromClientX(e.clientX)
        }}
        onKeyDown={(e) => {
          if (e.key === 'ArrowLeft') setV((value) => clamp(value - 2))
          if (e.key === 'ArrowRight') setV((value) => clamp(value + 2))
        }}
      >
        <img className="absolute inset-0 h-full w-full select-none object-contain" src={after} alt="После" draggable={false} />
        <div className="absolute inset-0 will-change-[clip-path]" style={{ clipPath: `inset(0 ${100 - v}% 0 0)` }}>
          <img className="absolute inset-0 h-full w-full select-none object-contain" src={before} alt="До" draggable={false} />
        </div>

        <div
          className="pointer-events-none absolute inset-y-0 flex -translate-x-1/2 items-center"
          style={{ left: `${v}%` }}
        >
          <div className="h-full w-[1px] bg-white/40 shadow-[0_0_10px_rgba(255,255,255,0.5)]" />
          <div className="absolute left-1/2 flex h-8 w-8 -translate-x-1/2 items-center justify-center rounded-full border border-white/20 bg-black text-white shadow-2xl backdrop-blur-md">
            <span className="h-3.5 w-[1px] rounded-full bg-white/60" />
            <span className="ml-1 h-3.5 w-[1px] rounded-full bg-white/60" />
          </div>
        </div>
      </div>

      <div className="mt-4 flex items-center gap-4">
        <div className="w-8 text-xs font-semibold uppercase tracking-wider text-zinc-500">До</div>
        <input
          className="w-full accent-white bg-zinc-800 h-1 rounded-lg appearance-none cursor-pointer"
          type="range"
          min={0}
          max={100}
          value={v}
          style={{ backgroundSize: `${v}% 100%` }}
          onChange={(e) => setV(Number(e.target.value))}
        />
        <div className="w-10 text-right text-xs font-semibold uppercase tracking-wider text-zinc-500">После</div>
      </div>
    </div>
  )
}
