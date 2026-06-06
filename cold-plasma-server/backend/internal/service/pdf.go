package service

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"cold-plasma-server/config"

	"github.com/jung-kurt/gofpdf"
)

type PDFService struct {
	fontPath string
	cfg      *config.Config
}

func NewPDFService(cfg *config.Config) *PDFService {
	return &PDFService{cfg: cfg, fontPath: cfg.PDFFontPath}
}

func (s *PDFService) CareMemo(ctx context.Context, userName string) ([]byte, error) {
	_ = ctx // сейчас генерация синхронная и не блокирующая, но сигнатуру оставляем для единообразия

	userName = strings.TrimSpace(userName)

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetTitle("Памятка по уходу — холодная плазма", false)
	pdf.AddPage()

	// UTF-8 шрифт для кириллицы
	pdf.AddUTF8Font("Roboto", "", s.fontPath)
	if pdf.Error() != nil {
		return nil, fmt.Errorf("не удалось загрузить шрифт для PDF (%s): %w", s.fontPath, pdf.Error())
	}
	pdf.SetFont("Roboto", "", 18)
	pdf.Cell(0, 10, "Памятка по уходу после процедуры «холодная плазма»")
	pdf.Ln(12)

	pdf.SetFont("Roboto", "", 12)
	if userName != "" {
		pdf.MultiCell(0, 6, fmt.Sprintf("%s, сохраните эту памятку — так эффект будет ровнее и комфортнее.", userName), "", "L", false)
	} else {
		pdf.MultiCell(0, 6, "Сохраните эту памятку — так эффект будет ровнее и комфортнее.", "", "L", false)
	}
	pdf.Ln(4)

	pdf.SetFont("Roboto", "", 13)
	pdf.Cell(0, 7, "Первые 24 часа")
	pdf.Ln(8)
	pdf.SetFont("Roboto", "", 12)
	pdf.MultiCell(0, 6, "• Не трогайте обработанную зону грязными руками.\n• Избегайте активных тренировок, бани/сауны, горячих ванн.\n• Не используйте скрабы/кислоты/ретиноиды.\n• Макияж — по самочувствию, лучше отложить до следующего дня.", "", "L", false)
	pdf.Ln(2)

	pdf.SetFont("Roboto", "", 13)
	pdf.Cell(0, 7, "3–7 дней")
	pdf.Ln(8)
	pdf.SetFont("Roboto", "", 12)
	pdf.MultiCell(0, 6, "• Мягкое очищение, увлажнение.\n• SPF 30+ каждый день (особенно при солнце).\n• Не сдирайте корочки/шелушение, если появится — дайте коже восстановиться.\n• Если есть дискомфорт — напишите нам, подскажем уход.", "", "L", false)
	pdf.Ln(2)

	pdf.SetFont("Roboto", "", 13)
	pdf.Cell(0, 7, "Когда лучше написать или прийти на осмотр")
	pdf.Ln(8)
	pdf.SetFont("Roboto", "", 12)
	pdf.MultiCell(0, 6, "• Сильное покраснение/отёк, которые не уменьшаются.\n• Выраженный зуд, пузырьки, необычная реакция.\n\nВажно: памятка не заменяет консультацию. Если есть хронические заболевания или вы принимаете лекарства — уточните у специалиста индивидуальные рекомендации.", "", "L", false)
	pdf.Ln(4)

	pdf.SetFont("Roboto", "", 11)
	pdf.MultiCell(0, 5, fmt.Sprintf("Салон: %s\nАдрес: %s\nРежим работы: %s\n\nХотите записаться? Напишите нам — подберём удобное время.", s.cfg.SalonName, s.cfg.SalonAddress, s.cfg.SalonWorkHours), "", "L", false)

	pdf.SetFont("Roboto", "", 9)
	pdf.Ln(6)
	pdf.Cell(0, 5, "Сформировано: "+time.Now().Format("02.01.2006 15:04"))

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("pdf output: %w", err)
	}
	return buf.Bytes(), nil
}
