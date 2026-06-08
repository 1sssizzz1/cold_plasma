package service

import "time"

// Рабочее окно записи: ежедневно с 12:00 до 19:00 по местному времени салона.
const (
	workdayStartHour = 12
	workdayEndHour   = 19
)

// salonLocation — часовой пояс салона (Северодвинск, MSK = UTC+3).
func salonLocation() *time.Location {
	if loc, err := time.LoadLocation("Europe/Moscow"); err == nil {
		return loc
	}
	return time.FixedZone("MSK", 3*60*60)
}

// Slot — одно окно записи [StartAt, EndAt).
type Slot struct {
	StartAt time.Time `json:"start_at"`
	EndAt   time.Time `json:"end_at"`
}

type interval struct {
	start time.Time
	end   time.Time
}

func overlaps(aStart, aEnd, bStart, bEnd time.Time) bool {
	return aStart.Before(bEnd) && bStart.Before(aEnd)
}

// generateDaySlots строит окна записи на конкретный день, нарезая рабочее
// время по длительности процедуры. Возвращает только окна, целиком
// помещающиеся в рабочий интервал.
func generateDaySlots(day time.Time, durationMins int, loc *time.Location) []Slot {
	if durationMins < 1 {
		return nil
	}
	dayStart := time.Date(day.Year(), day.Month(), day.Day(), workdayStartHour, 0, 0, 0, loc)
	dayEnd := time.Date(day.Year(), day.Month(), day.Day(), workdayEndHour, 0, 0, 0, loc)
	step := time.Duration(durationMins) * time.Minute

	out := make([]Slot, 0)
	for start := dayStart; !start.Add(step).After(dayEnd); start = start.Add(step) {
		out = append(out, Slot{StartAt: start, EndAt: start.Add(step)})
	}
	return out
}
