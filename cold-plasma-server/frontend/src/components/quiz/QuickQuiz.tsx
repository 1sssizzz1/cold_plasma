import { useMemo, useState } from 'react'

type Answer = 'yes' | 'no'

type Q = {
  key: string
  q: string
  help?: string
  riskIfYes?: boolean
}

const questions: Q[] = [
  { key: 'pregnant', q: 'Вы беременны / кормите грудью?', help: 'Если да — лучше сначала обсудить со специалистом.', riskIfYes: true },
  { key: 'acute', q: 'Есть активное воспаление/обострение в зоне?', help: 'При обострении обычно ждём ремиссию.', riskIfYes: true },
  { key: 'scalp', q: 'Интересует уход за кожей головы/волосами?', help: 'Подберём протокол именно под запрос.' },
  { key: 'sensitive', q: 'Кожа/кожа головы очень чувствительная?', help: 'Можно — но важно аккуратно подобрать уход.' },
]

export default function QuickQuiz() {
  const [step, setStep] = useState(0)
  const [answers, setAnswers] = useState<Record<string, Answer | null>>(() =>
    Object.fromEntries(questions.map((q) => [q.key, null])),
  )

  const current = questions[step]
  const done = step >= questions.length

  const result = useMemo(() => {
    if (!done) return null
    const riskYes = questions.some((q) => q.riskIfYes && answers[q.key] === 'yes')
    const wantsScalp = answers['scalp'] === 'yes'
    if (riskYes) {
      return {
        title: 'Нужна короткая консультация перед процедурой',
        body: 'Это нормально: мы уточним детали и подберём максимально бережный вариант. Напишите в чат — поможем.',
      }
    }
    return {
      title: wantsScalp ? 'Похоже, вам подойдёт вариант для кожи головы' : 'Похоже, вам подойдёт процедура',
      body: 'Точный ответ — после 2–3 уточняющих вопросов. Нажмите «Записаться» или задайте вопрос в чате.',
    }
  }, [answers, done])

  return (
    <div className="glass-panel rounded-[26px] p-6 shadow-[0_18px_52px_rgba(17,24,39,0.1)]">
      <div className="text-[10px] font-semibold uppercase tracking-widest text-zinc-500">Квиз</div>
      <div className="mt-1.5 text-xl font-semibold tracking-tight text-white">Подходит ли вам процедура?</div>
      <div className="mt-2 text-xs text-zinc-400">Ответ займёт меньше минуты.</div>

      {!done ? (
        <div className="mt-6">
          <div className="text-[10px] font-semibold uppercase tracking-wider text-zinc-500">
            Вопрос {step + 1} из {questions.length}
          </div>
          <div className="mt-3 text-base font-medium text-white">{current.q}</div>
          {current.help && <div className="mt-1.5 text-xs text-zinc-400">{current.help}</div>}

          <div className="mt-6 flex gap-3">
            <button
              className="h-11 flex-1 rounded-2xl border border-white/10 bg-white/5 text-xs font-semibold uppercase tracking-wider text-white transition-all duration-300 hover:bg-white/10 hover:border-white/20 hover:scale-[1.02] active:scale-[0.98]"
              onClick={() => {
                setAnswers((a) => ({ ...a, [current.key]: 'no' }))
                setStep((s) => s + 1)
              }}
            >
              Нет
            </button>
            <button
              className="h-11 flex-1 rounded-2xl bg-white text-xs font-semibold uppercase tracking-wider text-black transition-all duration-300 hover:bg-zinc-200 hover:scale-[1.02] active:scale-[0.98]"
              onClick={() => {
                setAnswers((a) => ({ ...a, [current.key]: 'yes' }))
                setStep((s) => s + 1)
              }}
            >
              Да
            </button>
          </div>
        </div>
      ) : (
        <div className="mt-6">
          <div className="text-base font-semibold text-white">{result?.title}</div>
          <div className="mt-2 text-xs leading-relaxed text-zinc-400">{result?.body}</div>
          <div className="mt-6 flex gap-3">
            <a
              href="/booking"
              className="inline-flex h-11 flex-1 items-center justify-center rounded-2xl bg-white text-xs font-semibold uppercase tracking-wider text-black transition-all duration-300 hover:bg-zinc-200 hover:scale-[1.02] active:scale-[0.98]"
            >
              Записаться
            </a>
            <button
              className="h-11 flex-1 rounded-2xl border border-white/10 bg-white/5 text-xs font-semibold uppercase tracking-wider text-white transition-all duration-300 hover:bg-white/10 hover:border-white/20 hover:scale-[1.02] active:scale-[0.98]"
              onClick={() => {
                setAnswers(Object.fromEntries(questions.map((q) => [q.key, null])))
                setStep(0)
              }}
            >
              Пройти снова
            </button>
          </div>
        </div>
      )}
    </div>
  )
}
