import Container from './Container'

export default function Footer() {
  return (
    <footer className="relative z-10 border-t border-black/10 bg-white/70 py-12 backdrop-blur-md">
      <Container>
        <div className="grid gap-8 md:grid-cols-3">
          <div>
            <div className="text-sm font-semibold uppercase tracking-wider text-white">Холодная плазма — Северодвинск</div>
            <div className="mt-3 text-xs leading-relaxed text-zinc-400">
              Сайт для записи, процедур и консультаций. Премиальный, минималистичный дизайн — без лишнего шума.
            </div>
          </div>
          <div className="text-xs text-zinc-400">
            <div className="font-semibold uppercase tracking-wider text-white">Контакты</div>
            <div className="mt-3">Адрес: укажите в .env</div>
            <div className="mt-1">Режим: ежедневно 10:00–20:00</div>
          </div>
          <div className="text-xs text-zinc-400">
            <div className="font-semibold uppercase tracking-wider text-white">Юридически</div>
            <div className="mt-3">Информация не является медицинской рекомендацией.</div>
            <div className="mt-4 text-[10px] text-zinc-600">© {new Date().getFullYear()}</div>
          </div>
        </div>
      </Container>
    </footer>
  )
}
