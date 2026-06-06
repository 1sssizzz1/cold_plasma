import { Helmet } from 'react-helmet-async'
import { Link } from 'react-router-dom'
import Container from '../components/Container'
import MotionPage from '../components/MotionPage'

export default function NotFoundPage() {
  return (
    <MotionPage>
      <Helmet>
        <title>Страница не найдена — Холодная плазма</title>
      </Helmet>
      <section className="py-16">
        <Container>
          <div className="glass-panel mx-auto max-w-lg rounded-3xl p-10 text-center shadow-xl">
            <div className="silver-glow text-4xl font-semibold text-white">404</div>
            <div className="mt-3 text-sm text-zinc-400">Такой страницы нет.</div>
            <Link
              to="/"
              className="mt-6 inline-flex h-11 items-center justify-center rounded-2xl bg-white px-6 text-xs font-semibold uppercase tracking-wider text-black transition hover:bg-zinc-200"
            >
              На главную
            </Link>
          </div>
        </Container>
      </section>
    </MotionPage>
  )
}
