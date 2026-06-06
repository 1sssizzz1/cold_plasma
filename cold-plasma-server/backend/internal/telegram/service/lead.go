package service

import (
	"context"
	"fmt"
	"strings"
)

type LeadAnalysis struct {
	Action            string `json:"action"`
	Temperature       string `json:"temperature"`
	Priority          string `json:"priority"`
	ManagerReplyDraft string `json:"manager_reply_draft"`
	Reason            string `json:"reason"`
}

func (s *Service) analyzeLead(ctx context.Context, payload map[string]any) LeadAnalysis {
	return localLeadAnalysis(payload)
}

func localLeadAnalysis(payload map[string]any) LeadAnalysis {
	contact := strings.TrimSpace(fmt.Sprint(payload["phone"]))
	if contact == "" || contact == "<nil>" {
		return LeadAnalysis{
			Action:      "notify_admin",
			Temperature: "notify_admin",
			Priority:    "high",
			Reason:      "контакт не найден - требуется ручная обработка",
		}
	}

	msg := strings.ToLower(fmt.Sprint(payload["message"]))
	datesRaw := strings.ToLower(fmt.Sprint(payload["requested_datetimes"]))
	combined := msg + " " + datesRaw

	hotWords := []string{
		"срочно", "сегодня", "завтра", "хочу записаться",
		"запишите", "свободно ли", "есть ли место",
		"цена", "стоимость", "сколько стоит",
		"когда можно", "удобно",
	}
	for _, word := range hotWords {
		if strings.Contains(combined, word) {
			return LeadAnalysis{
				Action:      "hot",
				Temperature: "hot",
				Priority:    "high",
				Reason:      fmt.Sprintf("ключевое слово %q - явный интерес к записи", word),
			}
		}
	}

	warmWords := []string{
		"хочу", "интересует", "расскажите", "подробнее",
		"записаться", "запись", "процедур",
	}
	procedureID := strings.TrimSpace(fmt.Sprint(payload["procedure_id"]))
	hasProcedure := procedureID != "" && procedureID != "0" && procedureID != "<nil>"

	for _, word := range warmWords {
		if strings.Contains(combined, word) {
			return LeadAnalysis{
				Action:      "warm",
				Temperature: "warm",
				Priority:    "normal",
				Reason:      fmt.Sprintf("ключевое слово %q", word),
			}
		}
	}
	if hasProcedure {
		return LeadAnalysis{
			Action:      "warm",
			Temperature: "warm",
			Priority:    "normal",
			Reason:      "выбрана конкретная процедура",
		}
	}
	if datesRaw != "" && datesRaw != "<nil>" && datesRaw != "[]" {
		return LeadAnalysis{
			Action:      "warm",
			Temperature: "warm",
			Priority:    "normal",
			Reason:      "указаны предпочтительные даты",
		}
	}

	return LeadAnalysis{
		Action:      "cold",
		Temperature: "cold",
		Priority:    "low",
		Reason:      "мало информации для классификации",
	}
}
