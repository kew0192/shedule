package DB

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"task_auth/users"

	_ "github.com/lib/pq"
)

type PostgreSQL struct {
	sql *sql.DB
}

func OpenPOSTGRESQL(connStr string) (*PostgreSQL, error) {
	sql, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := sql.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}
	createTableSQL := `
    CREATE TABLE IF NOT EXISTS users (
        id SERIAL PRIMARY KEY,
        First_name TEXT NOT NULL,
        Last_name TEXT NOT NULL,
		Role TEXT NOT NULL,
		Code TEXT NOT NULL,
		Access_token TEXT,
		Refresh_token TEXT
    )`

	if _, err := sql.Exec(createTableSQL); err != nil {
		return nil, fmt.Errorf("create table: %w", err)
	}
	return &PostgreSQL{sql: sql}, nil
}
func (s *PostgreSQL) GetAllUsersDB(ctx context.Context) ([]*users.User, error) {
	query := `SELECT id, First_name, Last_name, Role, Code FROM users ORDER BY id`

	rows, err := s.sql.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get all users: %w", err)
	}
	defer rows.Close()

	var userList []*users.User
	for rows.Next() {
		var user users.User
		err := rows.Scan(
			&user.ID,
			&user.First_name,
			&user.Last_name,
			&user.Role,
			&user.Code,
		)
		if err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		userList = append(userList, &user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	log.Printf("📋 Получено %d пользователей", len(userList))
	return userList, nil
}
func (s *PostgreSQL) SaveUserDB(ctx context.Context, user *users.User) error {
	log.Printf("📌 SaveUserDB: code=%s, name=%s", user.Code, user.First_name)

	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM users WHERE code = $1)`
	err := s.sql.QueryRowContext(ctx, checkQuery, user.Code).Scan(&exists)
	if err != nil {
		log.Printf("❌ Ошибка проверки exists: %v", err)
		return fmt.Errorf("check exists: %w", err)
	}

	if exists {
		query := `
            UPDATE users 
            SET first_name = $1, 
                last_name = $2, 
                role = $3, 
                access_token = $4, 
                refresh_token = $5
            WHERE code = $6
            RETURNING id
        `
		return s.sql.QueryRowContext(
			ctx,
			query,
			user.First_name,
			user.Last_name,
			user.Role,
			user.AccessToken,
			user.RefreshToken,
			user.Code,
		).Scan(&user.ID)
	}

	query := `
        INSERT INTO users (first_name, last_name, role, code, access_token, refresh_token) 
        VALUES ($1, $2, $3, $4, $5, $6) 
        RETURNING id
    `
	err = s.sql.QueryRowContext(
		ctx,
		query,
		user.First_name,
		user.Last_name,
		user.Role,
		user.Code,
		user.AccessToken,
		user.RefreshToken,
	).Scan(&user.ID)
	if err != nil {
		log.Printf("❌ Ошибка INSERT: %v", err)
		return fmt.Errorf("insert user: %w", err)
	}

	log.Printf("✅ Вставлен новый пользователь: ID=%d, code=%s", user.ID, user.Code)
	return nil
}

func (s *PostgreSQL) FindUserDB(ctx context.Context, id int) (*users.User, error) {
	query := `SELECT id, First_name, Last_name, Role, Code, Access_token, Refresh_token FROM users WHERE id = $1`

	var user users.User
	err := s.sql.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.First_name, &user.Last_name, &user.Role, &user.Code, &user.AccessToken, &user.RefreshToken,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}

	return &user, nil
}
func (s *PostgreSQL) FindUserCodeDB(ctx context.Context, code string) (*users.User, error) {
	// ✅ Не читаем Access_token и Refresh_token
	query := `SELECT id, First_name, Last_name, Role, Code FROM users WHERE Code = $1`

	var user users.User
	err := s.sql.QueryRowContext(ctx, query, code).Scan(
		&user.ID, &user.First_name, &user.Last_name, &user.Role, &user.Code,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find user by code: %w", err)
	}

	return &user, nil
}
func (s *PostgreSQL) DeleteUserDB(ctx context.Context, id int) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := s.sql.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("user with id %d not found", id)
	}

	return nil
}

func (s *PostgreSQL) UpdateUserDB(ctx context.Context, user *users.User) error {
	query := `UPDATE users SET First_name = $1, Last_name = $2, Role = $3, Code = $4, Access_token = $5, Refresh_token = $6 WHERE id = $7`

	result, err := s.sql.ExecContext(ctx, query, user.First_name, user.Last_name, user.Role, user.Code, user.AccessToken, user.RefreshToken, user.ID)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("user with id %d not found", user.ID)
	}

	return nil
}
