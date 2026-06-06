import { Helmet } from 'react-helmet-async'
import Container from '../components/Container'
import MotionPage from '../components/MotionPage'
import QuickQuiz from '../components/quiz/QuickQuiz'

const faq = [
  {
    q: 'Больно ли?',
    a: 'Обычно ощущения комфортные. Если вы очень чувствительны — скажите об этом, подберём максимально мягкий режим и уход.',
  },
  {
    q: 'Сколько держится эффект?',
    a: 'Зависит от исходного состояния и домашнего ухода. Мы подскажем, как поддержать результат простыми шагами.',
  },
  {
    q: 'Можно ли летом?',
    a: 'Как правило — да, но важно соблюдать SPF и базовый уход. В чате уточним детали и подскажем безопасный вариант.',
  },
  {
    q: 'Есть ли противопоказания?',
    a: 'Есть ситуации, когда лучше сначала проконсультироваться. Напишите, что именно беспокоит — зададим пару вопросов и сориентируем.',
  },
]

export default function HelpPage() {
  return (
    <MotionPage>
      <Helmet>
        <title>Помощь — Холодная плазма (Северодвинск)</title>
        <meta name="description" content="Ответы на популярные вопросы и быстрый квиз. Если нужно — задайте вопрос в AI-чате." />
      </Helmet>

      <section className="py-12">
        <Container>
          <div className="grid gap-8 md:grid-cols-2">
            <div>
              <div className="border-b border-white/5 pb-6">
                <div className="text-xs font-semibold uppercase tracking-widest text-zinc-500">Вопросы</div>
                <div className="mt-2 text-3xl font-semibold tracking-tight text-white sm:text-4xl">Помощь</div>
              </div>
              <div className="mt-5 text-sm leading-relaxed text-zinc-400">
                Спросите в чате внизу справа — отвечаем кратко и по теме (холодная плазма + уход).
              </div>

              <div className="mt-6 space-y-4">
                {faq.map((f) => (
                  <div key={f.q} className="glass-panel rounded-3xl p-5 shadow-xl transition-all duration-300 hover:border-white/15">
                    <div className="text-sm font-semibold text-white">{f.q}</div>
                    <div className="mt-2 text-sm leading-relaxed text-zinc-400">{f.a}</div>
                  </div>
                ))}
              </div>
            </div>
            <QuickQuiz />
          </div>
        </Container>
      </section>
    </MotionPage>
  )
}
