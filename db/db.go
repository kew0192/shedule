package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"task_logic/objects"

	_ "github.com/lib/pq"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(connStr string) (*PostgresRepository, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	return &PostgresRepository{db: db}, nil
}

func (r *PostgresRepository) Close() error {
	return r.db.Close()
}

func (r *PostgresRepository) InitTables(ctx context.Context) error {
	teachersTable := `
	CREATE TABLE IF NOT EXISTS teachers (
		id SERIAL PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		lesson JSONB NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`

	if _, err := r.db.ExecContext(ctx, teachersTable); err != nil {
		return fmt.Errorf("create teachers table: %w", err)
	}

	classesTable := `
	CREATE TABLE IF NOT EXISTS classes (
		id SERIAL PRIMARY KEY,
		name VARCHAR(100) NOT NULL UNIQUE,
		week JSONB NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`

	if _, err := r.db.ExecContext(ctx, classesTable); err != nil {
		return fmt.Errorf("create classes table: %w", err)
	}

	return nil
}

// ===== КЛАССЫ =====

func (r *PostgresRepository) GetClass(ctx context.Context, id int) (*objects.Class, error) {
	var class objects.Class
	var weekJSON []byte

	query := `SELECT id, name, week FROM classes WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(&class.ID, &class.Name, &weekJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get class: %w", err)
	}

	err = json.Unmarshal(weekJSON, &class.Week)
	if err != nil {
		return nil, fmt.Errorf("parse week: %w", err)
	}

	return &class, nil
}

func (r *PostgresRepository) GetAllClasses(ctx context.Context) ([]*objects.Class, error) {
	query := `SELECT id, name, week FROM classes ORDER BY name`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get all classes: %w", err)
	}
	defer rows.Close()

	var classes []*objects.Class

	for rows.Next() {
		var class objects.Class
		var weekJSON []byte

		err := rows.Scan(&class.ID, &class.Name, &weekJSON)
		if err != nil {
			return nil, fmt.Errorf("scan class: %w", err)
		}

		err = json.Unmarshal(weekJSON, &class.Week)
		if err != nil {
			return nil, fmt.Errorf("parse week: %w", err)
		}

		classes = append(classes, &class)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return classes, nil
}

func (r *PostgresRepository) GetClassByName(ctx context.Context, name string) (*objects.Class, error) {
	var class objects.Class
	var weekJSON []byte

	query := `SELECT id, name, week FROM classes WHERE name = $1`
	err := r.db.QueryRowContext(ctx, query, name).Scan(&class.ID, &class.Name, &weekJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get class by name: %w", err)
	}

	err = json.Unmarshal(weekJSON, &class.Week)
	if err != nil {
		return nil, fmt.Errorf("parse week: %w", err)
	}

	return &class, nil
}

func (r *PostgresRepository) UpdateClass(ctx context.Context, class *objects.Class) error {
	weekJSON, err := json.Marshal(class.Week)
	if err != nil {
		return fmt.Errorf("marshal week: %w", err)
	}

	query := `UPDATE classes SET week = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	_, err = r.db.ExecContext(ctx, query, weekJSON, class.ID)
	if err != nil {
		return fmt.Errorf("update class: %w", err)
	}

	return nil
}

func (r *PostgresRepository) CreateClass(ctx context.Context, name string) (*objects.Class, error) {
	log.Printf("📦 Создание класса: %s", name)

	// Проверяем, существует ли уже такой класс
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM classes WHERE name = $1)`
	err := r.db.QueryRowContext(ctx, checkQuery, name).Scan(&exists)
	if err != nil {
		log.Printf("❌ Ошибка проверки существования: %v", err)
		return nil, fmt.Errorf("check exists: %w", err)
	}

	if exists {
		log.Printf("⚠️ Класс %s уже существует", name)
		return nil, fmt.Errorf("class %s already exists", name)
	}

	// Создаём объект
	class, err := objects.NewClass(ctx, name)
	if err != nil {
		log.Printf("❌ Ошибка NewClass: %v", err)
		return nil, err
	}

	// Маршалим неделю
	weekJSON, err := json.Marshal(class.Week)
	if err != nil {
		log.Printf("❌ Ошибка Marshal: %v", err)
		return nil, fmt.Errorf("marshal week: %w", err)
	}

	log.Printf("📦 JSON недели: %s", string(weekJSON))

	// Вставляем в БД
	query := `INSERT INTO classes (name, week) VALUES ($1, $2) RETURNING id`
	err = r.db.QueryRowContext(ctx, query, class.Name, weekJSON).Scan(&class.ID)
	if err != nil {
		log.Printf("❌ Ошибка INSERT: %v", err)
		return nil, fmt.Errorf("create class: %w", err)
	}

	log.Printf("✅ Класс создан с ID: %d", class.ID)
	return class, nil
}

func (r *PostgresRepository) DeleteClass(ctx context.Context, id int) error {
	query := `DELETE FROM classes WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete class: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("class with id %d not found", id)
	}

	return nil
}

// ===== УЧИТЕЛЯ =====

func (r *PostgresRepository) GetTeacher(ctx context.Context, id int) (*objects.Teacher, error) {
	var teacher objects.Teacher
	var lessonJSON []byte

	query := `SELECT id, name, lesson FROM teachers WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(&teacher.ID, &teacher.Name, &lessonJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get teacher: %w", err)
	}

	err = json.Unmarshal(lessonJSON, &teacher.Lesson)
	if err != nil {
		return nil, fmt.Errorf("parse lesson: %w", err)
	}

	return &teacher, nil
}

func (r *PostgresRepository) GetTeacherByName(ctx context.Context, name string) (*objects.Teacher, error) {
	var teacher objects.Teacher
	var lessonJSON []byte

	query := `SELECT id, name, lesson FROM teachers WHERE name = $1`
	err := r.db.QueryRowContext(ctx, query, name).Scan(&teacher.ID, &teacher.Name, &lessonJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get teacher by name: %w", err)
	}

	err = json.Unmarshal(lessonJSON, &teacher.Lesson)
	if err != nil {
		return nil, fmt.Errorf("parse lesson: %w", err)
	}

	return &teacher, nil
}
func (r *PostgresRepository) GetTeacherByNameAndSubject(ctx context.Context, name string, subject string) (*objects.Teacher, error) {
	var teacher objects.Teacher
	var lessonJSON []byte

	query := `SELECT id, name, lesson FROM teachers WHERE name = $1 AND lesson->>'name' = $2`
	err := r.db.QueryRowContext(ctx, query, name, subject).Scan(&teacher.ID, &teacher.Name, &lessonJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get teacher by name and subject: %w", err)
	}

	err = json.Unmarshal(lessonJSON, &teacher.Lesson)
	if err != nil {
		return nil, fmt.Errorf("parse lesson: %w", err)
	}

	return &teacher, nil
}

func (r *PostgresRepository) CreateTeacher(ctx context.Context, teacher *objects.Teacher) error {
	lessonJSON, err := json.Marshal(teacher.Lesson)
	if err != nil {
		return fmt.Errorf("marshal lesson: %w", err)
	}

	query := `INSERT INTO teachers (name, lesson) VALUES ($1, $2) RETURNING id`
	err = r.db.QueryRowContext(ctx, query, teacher.Name, lessonJSON).Scan(&teacher.ID)
	if err != nil {
		return fmt.Errorf("create teacher: %w", err)
	}

	return nil
}

func (r *PostgresRepository) DeleteTeacher(ctx context.Context, id int) error {
	query := `DELETE FROM teachers WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete teacher: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("teacher with id %d not found", id)
	}

	return nil
}
