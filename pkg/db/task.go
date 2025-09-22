package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Task — как задача выглядит в коде и при обмене JSON.
type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

// AddTask вставляет задачу в таблицу scheduler и возвращает id записи.
func AddTask(t *Task) (int64, error) {
	if DB == nil {
		return 0, errors.New("db not initialized")
	}
	const q = `
INSERT INTO scheduler (date, title, comment, repeat)
VALUES (?, ?, ?, ?);
`
	res, err := DB.Exec(q, t.Date, t.Title, t.Comment, t.Repeat)
	if err != nil {
		return 0, fmt.Errorf("insert: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("last id: %w", err)
	}
	return id, nil
}

// Tasks выбирает задачи из БД.
// Если search == "" -> просто ближайшие задачи.
// Если search похож на "02.01.2006" -> точное совпадение по этой дате.
// Иначе -> подстрочный поиск в title/comment (LIKE).
func Tasks(limit int, search string) ([]*Task, error) {
	if DB == nil {
		return nil, fmt.Errorf("db not initialized")
	}
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	search = strings.TrimSpace(search)

	const base = `SELECT id, date, title, comment, repeat FROM scheduler`
	var (
		rows *sql.Rows
		err  error
	)

	// поиск по дате формата 02.01.2006 → конвертируем в 20060102
	if yyyymmdd, ok := tryParseRuDate(search); ok {
		q := base + ` WHERE date = ? ORDER BY date, id LIMIT ?`
		rows, err = DB.Query(q, yyyymmdd, limit)
	} else if search != "" {
		// подстрока (LIKE) по title/comment — без конкатенации SQL
		p := "%" + search + "%"
		q := base + ` WHERE title LIKE ? OR comment LIKE ? ORDER BY date, id LIMIT ?`
		rows, err = DB.Query(q, p, p, limit)
	} else {
		// без фильтра
		q := base + ` ORDER BY date, id LIMIT ?`
		rows, err = DB.Query(q, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*Task
	for rows.Next() {
		var (
			idInt int64
			t     Task
		)
		if err := rows.Scan(&idInt, &t.Date, &t.Title, &t.Comment, &t.Repeat); err != nil {
			return nil, err
		}
		t.ID = fmt.Sprintf("%d", idInt) // хотим строку в JSON
		out = append(out, &t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if out == nil {
		out = make([]*Task, 0)
	}
	return out, nil
}

// tryParseRuDate: "02.01.2006" -> "20060102"
func tryParseRuDate(s string) (string, bool) {
	const (
		ruFmt = "02.01.2006"
		goFmt = "20060102"
	)
	d, err := time.Parse(ruFmt, s)
	if err != nil {
		return "", false
	}
	return d.Format(goFmt), true
}

// GetTask возвращает задачу по id.
func GetTask(id string) (*Task, error) {
	if DB == nil {
		return nil, fmt.Errorf("db not initialized")
	}
	n, err := strconv.ParseInt(strings.TrimSpace(id), 10, 64)
	if err != nil || n <= 0 {
		return nil, fmt.Errorf("bad id")
	}

	const q = `SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?`
	var (
		idInt int64
		t     Task
	)
	err = DB.QueryRow(q, n).Scan(&idInt, &t.Date, &t.Title, &t.Comment, &t.Repeat)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("task not found")
		}
		return nil, err
	}
	t.ID = fmt.Sprintf("%d", idInt)
	return &t, nil
}

// UpdateTask обновляет поля задачи по её id.
func UpdateTask(t *Task) error {
	if DB == nil {
		return fmt.Errorf("db not initialized")
	}
	n, err := strconv.ParseInt(strings.TrimSpace(t.ID), 10, 64)
	if err != nil || n <= 0 {
		return fmt.Errorf("bad id")
	}
	const q = `UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`
	res, err := DB.Exec(q, t.Date, t.Title, t.Comment, t.Repeat, n)
	if err != nil {
		return err
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if aff == 0 {
		return fmt.Errorf("incorrect id for updating task")
	}
	return nil
}

/************* шаг 7: удалить и поменять дату *************/

func DeleteTask(id string) error {
	if DB == nil {
		return fmt.Errorf("db not initialized")
	}
	n, err := strconv.ParseInt(strings.TrimSpace(id), 10, 64)
	if err != nil || n <= 0 {
		return fmt.Errorf("bad id")
	}
	const q = `DELETE FROM scheduler WHERE id = ?`
	res, err := DB.Exec(q, n)
	if err != nil {
		return err
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if aff == 0 {
		return fmt.Errorf("task not found")
	}
	return nil
}

func UpdateDate(next string, id string) error {
	if DB == nil {
		return fmt.Errorf("db not initialized")
	}
	// можно довериться уже проверенному next (его считаем в api),
	// но на всякий случай приведём
	next = strings.TrimSpace(next)
	if len(next) != 8 {
		return fmt.Errorf("bad date format")
	}
	n, err := strconv.ParseInt(strings.TrimSpace(id), 10, 64)
	if err != nil || n <= 0 {
		return fmt.Errorf("bad id")
	}
	const q = `UPDATE scheduler SET date = ? WHERE id = ?`
	res, err := DB.Exec(q, next, n)
	if err != nil {
		return err
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if aff == 0 {
		return fmt.Errorf("task not found")
	}
	return nil
}
