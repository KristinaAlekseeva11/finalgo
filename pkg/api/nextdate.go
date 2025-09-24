package api

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// формат даты в проекте один и тот же
const dateFmt = "20060102"

// NextDate считает ближайшую дату > now по правилу repeat, начиная отсчет от dstart.
// ВАЖНО: сначала делаем минимум ОДИН шаг по правилу, потом сравниваем с now.
// Это соответствует подсказке в задании (цикл с AddDate перед проверкой).
func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	repeat = strings.TrimSpace(repeat)
	if repeat == "" {
		return "", errors.New("empty repeat: no rule")
	}
	start, err := time.Parse(dateFmt, dstart)
	if err != nil {
		return "", fmt.Errorf("bad dstart %q: %w", dstart, err)
	}

	// удобная функция: строго после now (без времени)
	afterNow := func(d time.Time, n time.Time) bool {
		d = dateOnly(d)
		n = dateOnly(n)
		return d.After(n)
	}

	parts := splitSpace(repeat)
	switch parts[0] {
	case "y":
		// ежегодно, без параметров
		if len(parts) != 1 {
			return "", errors.New("bad repeat: 'y' takes no args")
		}
		// делаем хотя бы один шаг на год, потом пока не перелезем через now
		cur := start.AddDate(1, 0, 0)
		for !afterNow(cur, now) {
			cur = cur.AddDate(1, 0, 0)
		}
		return cur.Format(dateFmt), nil

	case "d":
		// d <N> — шаг в днях, 1..400
		if len(parts) != 2 {
			return "", errors.New("bad repeat: 'd <N>' required")
		}
		n, err := strconv.Atoi(parts[1])
		if err != nil || n <= 0 || n > 400 {
			return "", errors.New("bad repeat: days must be 1..400")
		}
		// минимум один шаг вперёд
		cur := start.AddDate(0, 0, n)
		for !afterNow(cur, now) {
			cur = cur.AddDate(0, 0, n)
		}
		return cur.Format(dateFmt), nil

	case "w":
		// w <list of 1..7> — 1=пн ... 7=вс
		// пример: w 1,4,5
		if len(parts) != 2 {
			return "", errors.New("bad repeat: 'w <list>' required")
		}
		week, err := parseWeekdays(parts[1])
		if err != nil {
			return "", err
		}
		// идём со следующего дня (минимум один шаг)
		cur := start.AddDate(0, 0, 1)
		limit := 366 * 5
		for i := 0; i < limit; i++ {
			if afterNow(cur, now) && weekMatch(cur, week) {
				return cur.Format(dateFmt), nil
			}
			cur = cur.AddDate(0, 0, 1)
		}
		return "", errors.New("no next date for 'w' within limit")

	case "m":
		// m <days list> [months list]
		// дни: 1..31, также -1 (последний), -2 (предпоследний)
		// месяцы (опционально): 1..12
		if len(parts) < 2 || len(parts) > 3 {
			return "", errors.New("bad repeat: 'm <days> [months]'")
		}
		days, err := parseMonthDays(parts[1])
		if err != nil {
			return "", err
		}
		var months map[int]bool
		if len(parts) == 3 {
			months, err = parseMonths(parts[2])
			if err != nil {
				return "", err
			}
		}
		// начинаем со следующего дня (чтобы минимум один шаг)
		cur := start.AddDate(0, 0, 1)
		limit := 366 * 6
		for i := 0; i < limit; i++ {
			if afterNow(cur, now) && monthMatch(cur, days, months) {
				return cur.Format(dateFmt), nil
			}
			cur = cur.AddDate(0, 0, 1)
		}
		return "", errors.New("no next date for 'm' within limit")

	default:
		return "", errors.New("bad repeat: unknown rule")
	}
}

// --------------------- утилиты ниже ---------------------

// убираем время
func dateOnly(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.Local)
}

func splitSpace(s string) []string {
	s = strings.TrimSpace(s)
	fields := strings.Fields(s)
	return fields
}

func parseWeekdays(s string) (map[int]bool, error) {
	m := map[int]bool{}
	for _, p := range splitCSV(s) {
		n, err := strconv.Atoi(p)
		if err != nil || n < 1 || n > 7 {
			return nil, errors.New("bad 'w': day must be 1..7")
		}
		m[n] = true
	}
	if len(m) == 0 {
		return nil, errors.New("bad 'w': empty list")
	}
	return m, nil
}

func weekMatch(d time.Time, week map[int]bool) bool {
	wd := int(d.Weekday()) // Go: 0=вс, 1=пн, ... 6=сб
	if wd == 0 {
		wd = 7 // нам нужно 1..7, где 7=вс
	}
	return week[wd]
}

func parseMonthDays(s string) (map[int]bool, error) {
	m := map[int]bool{}
	for _, p := range splitCSV(s) {
		n, err := strconv.Atoi(p)
		if err != nil {
			return nil, errors.New("bad 'm': day must be int")
		}
		// допустимо 1..31, а также -1 и -2
		if !((n >= 1 && n <= 31) || n == -1 || n == -2) {
			return nil, errors.New("bad 'm': day must be 1..31 or -1 or -2")
		}
		m[n] = true
	}
	if len(m) == 0 {
		return nil, errors.New("bad 'm': empty day list")
	}
	return m, nil
}

func parseMonths(s string) (map[int]bool, error) {
	m := map[int]bool{}
	for _, p := range splitCSV(s) {
		n, err := strconv.Atoi(p)
		if err != nil || n < 1 || n > 12 {
			return nil, errors.New("bad 'm': month must be 1..12")
		}
		m[n] = true
	}
	if len(m) == 0 {
		return nil, errors.New("bad 'm': empty month list")
	}
	return m, nil
}

func monthMatch(d time.Time, days map[int]bool, months map[int]bool) bool {
	// если список месяцев задан — месяц должен попадать
	if months != nil {
		if !months[int(d.Month())] {
			return false
		}
	}
	day := d.Day()
	// проверяем обычные числа
	if days[day] {
		return true
	}
	// -1 — последний день месяца, -2 — предпоследний
	last := lastDayOfMonth(d.Year(), d.Month())
	if days[-1] && day == last {
		return true
	}
	if days[-2] && day == last-1 {
		return true
	}
	return false
}

func lastDayOfMonth(y int, m time.Month) int {
	// берём 1 число след. месяца и отматываем день назад
	t := time.Date(y, m+1, 1, 0, 0, 0, 0, time.Local)
	t = t.AddDate(0, 0, -1)
	return t.Day()
}

func splitCSV(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	// сортировка не обязательна, просто приятно
	sort.Strings(parts)
	return parts
}

// --------------------- HTTP: /api/nextdate ---------------------

func nextDateHandler(w http.ResponseWriter, r *http.Request) {
	// читаем query: now, date, repeat
	q := r.URL.Query()
	nowStr := q.Get("now")
	dateStr := q.Get("date")
	repeat := q.Get("repeat")

	// если now не передали — берём "сегодня"
	var now time.Time
	var err error
	if strings.TrimSpace(nowStr) == "" {
		now = dateOnly(time.Now())
	} else {
		now, err = time.Parse(dateFmt, nowStr)
		if err != nil {
			http.Error(w, "bad now", http.StatusBadRequest)
			return
		}
	}

	if strings.TrimSpace(dateStr) == "" || strings.TrimSpace(repeat) == "" {
		http.Error(w, "date/repeat required", http.StatusBadRequest)
		return
	}

	next, err := NextDate(now, dateStr, repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintln(w, next)
}

func mustEsc(s string) string { return url.QueryEscape(s) }
