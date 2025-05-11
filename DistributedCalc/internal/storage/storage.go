package storage

import (
	"DistributedCalc/pkg/logger"
	"database/sql"
	"errors"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteDB struct {
	db   *sql.DB
	logr *logger.Logger
}

type User struct {
	ID       int64
	Login    string
	Password string
}

type Expression struct {
	ID         int64
	UserID     int64
	Expression string
	Result     float64
	Status     string
}

type Task struct {
	ID           int64
	ExpressionID int64
	Arg1         float64
	Arg2         float64
	Operator     string
	Duration     int
	Result       float64
	Status       string
}

func NewSQLiteDB(path string, logr *logger.Logger) (*SQLiteDB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		logr.Error("Failed to open database: %v", err)
		return nil, err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			login TEXT UNIQUE,
			password TEXT
		);
		CREATE TABLE IF NOT EXISTS expressions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER,
			expression TEXT,
			result REAL,
			status TEXT,
			FOREIGN KEY (user_id) REFERENCES users(id)
		);
		CREATE TABLE IF NOT EXISTS tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			expression_id INTEGER,
			arg1 REAL,
			arg2 REAL,
			operator TEXT,
			duration INTEGER,
			result REAL,
			status TEXT,
			FOREIGN KEY (expression_id) REFERENCES expressions(id)
		);
	`)
	if err != nil {
		logr.Error("Failed to create tables: %v", err)
		return nil, err
	}

	return &SQLiteDB{db: db, logr: logr}, nil
}

func (s *SQLiteDB) Close() {
	s.db.Close()
}

func (s *SQLiteDB) CreateUser(login, password string) (int64, error) {
	result, err := s.db.Exec("INSERT INTO users (login, password) VALUES (?, ?)", login, password)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return 0, errors.New("user already exists")
		}
		s.logr.Error("Failed to insert user: %v", err)
		return 0, err
	}
	id, _ := result.LastInsertId()
	return id, nil
}

func (s *SQLiteDB) GetUser(login string) (User, error) {
	var user User
	err := s.db.QueryRow("SELECT id, login, password FROM users WHERE login = ?", login).Scan(&user.ID, &user.Login, &user.Password)
	if err == sql.ErrNoRows {
		return User{}, errors.New("user not found")
	}
	if err != nil {
		s.logr.Error("Failed to get user: %v", err)
		return User{}, err
	}
	return user, nil
}

func (s *SQLiteDB) SaveExpression(userID int64, expr string) (int64, error) {
	result, err := s.db.Exec("INSERT INTO expressions (user_id, expression, result, status) VALUES (?, ?, 0, 'pending')", userID, expr)
	if err != nil {
		s.logr.Error("Failed to insert expression: %v", err)
		return 0, err
	}
	id, _ := result.LastInsertId()
	return id, nil
}

func (s *SQLiteDB) GetUserExpressions(userID int64) ([]Expression, error) {
	rows, err := s.db.Query("SELECT id, user_id, expression, result, status FROM expressions WHERE user_id = ?", userID)
	if err != nil {
		s.logr.Error("Failed to query expressions: %v", err)
		return nil, err
	}
	defer rows.Close()

	var exprs []Expression
	for rows.Next() {
		var expr Expression
		if err := rows.Scan(&expr.ID, &expr.UserID, &expr.Expression, &expr.Result, &expr.Status); err != nil {
			s.logr.Error("Failed to scan expression: %v", err)
			return nil, err
		}
		exprs = append(exprs, expr)
	}
	return exprs, nil
}

func (s *SQLiteDB) GetExpression(id, userID int64) (Expression, error) {
	var expr Expression
	err := s.db.QueryRow("SELECT id, user_id, expression, result, status FROM expressions WHERE id = ? AND user_id = ?", id, userID).
		Scan(&expr.ID, &expr.UserID, &expr.Expression, &expr.Result, &expr.Status)
	if err == sql.ErrNoRows {
		return Expression{}, errors.New("expression not found")
	}
	if err != nil {
		s.logr.Error("Failed to get expression: %v", err)
		return Expression{}, err
	}
	return expr, nil
}

func (s *SQLiteDB) SaveTask(exprID int64, arg1, arg2 float64, op string, duration int) (int64, error) {
	result, err := s.db.Exec("INSERT INTO tasks (expression_id, arg1, arg2, operator, duration, result, status) VALUES (?, ?, ?, ?, ?, 0, 'pending')",
		exprID, arg1, arg2, op, duration)
	if err != nil {
		s.logr.Error("Failed to insert task: %v", err)
		return 0, err
	}
	id, _ := result.LastInsertId()
	return id, nil
}

func (s *SQLiteDB) UpdateTaskResult(taskID int64, result float64, status string) error {
	_, err := s.db.Exec("UPDATE tasks SET result = ?, status = ? WHERE id = ?", result, status, taskID)
	if err != nil {
		s.logr.Error("Failed to update task: %v", err)
		return err
	}
	return nil
}

func (s *SQLiteDB) GetPendingExpressions() ([]Expression, error) {
	rows, err := s.db.Query("SELECT id, user_id, expression, result, status FROM expressions WHERE status = 'pending'")
	if err != nil {
		s.logr.Error("Failed to query pending expressions: %v", err)
		return nil, err
	}
	defer rows.Close()

	var exprs []Expression
	for rows.Next() {
		var expr Expression
		if err := rows.Scan(&expr.ID, &expr.UserID, &expr.Expression, &expr.Result, &expr.Status); err != nil {
			s.logr.Error("Failed to scan expression: %v", err)
			return nil, err
		}
		exprs = append(exprs, expr)
	}
	return exprs, nil
}

func (s *SQLiteDB) UpdateExpression(exprID int64, result float64, status string) error {
	_, err := s.db.Exec("UPDATE expressions SET result = ?, status = ? WHERE id = ?", result, status, exprID)
	if err != nil {
		s.logr.Error("Failed to update expression: %v", err)
		return err
	}
	return nil
}

func (s *SQLiteDB) GetPendingTask() (Task, error) {
	var task Task
	err := s.db.QueryRow("SELECT id, expression_id, arg1, arg2, operator, duration, result, status FROM tasks WHERE status = 'pending' LIMIT 1").
		Scan(&task.ID, &task.ExpressionID, &task.Arg1, &task.Arg2, &task.Operator, &task.Duration, &task.Result, &task.Status)
	if err == sql.ErrNoRows {
		return Task{}, errors.New("no pending tasks")
	}
	if err != nil {
		s.logr.Error("Failed to get pending task: %v", err)
		return Task{}, err
	}
	return task, nil
}
