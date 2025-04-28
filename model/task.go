package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/raikh/calc_micro_final/internal/database"

	sq "github.com/Masterminds/squirrel"
)

type StringArray []string

func (a *StringArray) Scan(src interface{}) error {
	if src == nil {
		*a = nil
		return nil
	}
	var source []byte
	switch s := src.(type) {
	case []byte:
		source = s
	case string:
		source = []byte(s)
	default:
		return fmt.Errorf("unsupported type: %T", src)
	}
	return json.Unmarshal(source, a)
}

func (a StringArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	return json.Marshal(a)
}

type Task struct {
	Id            string      `json:"id" db:"id"`
	ExpressionId  string      `json:"expression_id" db:"expression_id"`
	Arg1          *float64    `json:"arg1" db:"arg1"`
	Arg2          *float64    `json:"arg2" db:"arg2"`
	Operation     string      `json:"operation" db:"operation"`
	OperationTime int64       `json:"operation_time" db:"operation_time"`
	Dependencies  StringArray `json:"-" db:"dependencies"`
	Result        *float64    `json:"result" db:"result"`
	Completed     bool        `json:"-" db:"completed"`
	IsProcessing  bool        `json:"-" db:"is_processing"`
	CreatedAt     *time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     *time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt     *time.Time  `json:"-" db:"deleted_at"`
}

func (e *Task) buildInsertExpression() (string, []interface{}, error) {
	sql, args, err := sq.Insert("tasks").
		Columns("id", "expression_id", "arg1", "arg2", "operation", "completed", "is_processing", "operation_time", "dependencies", "created_at", "updated_at").
		Values(e.Id, e.ExpressionId, e.Arg1, e.Arg2, e.Operation, e.Completed, e.IsProcessing, e.OperationTime, e.Dependencies, e.CreatedAt, e.UpdatedAt).
		ToSql()
	if err != nil {
		return "", nil, err
	}

	return sql, args, nil
}

func (e *Task) InsertTx(tx *sqlx.Tx) error {
	sql, args, err := e.buildInsertExpression()

	if err != nil {
		return err
	}

	res, err := tx.Exec(sql, args...)
	if err != nil {
		return err
	}
	_, err = res.LastInsertId()
	if err != nil {
		return err
	}

	return nil
}

func (e *Task) Insert() error {
	sql, args, err := e.buildInsertExpression()

	res, err := database.GetDB().Exec(sql, args...)

	if err != nil {
		return err
	}

	_, err = res.LastInsertId()

	if err != nil {
		return err
	}

	return nil
}

func (e *Task) Update() error {
	now := time.Now()
	sql, args, err := sq.Update("tasks").
		Set("arg1", e.Arg1).
		Set("arg2", e.Arg2).
		Set("result", e.Result).
		Set("completed", e.Completed).
		Set("is_processing", e.IsProcessing).
		Set("updated_at", &now).
		Where(sq.Eq{"id": e.Id}).
		ToSql()
	if err != nil {
		return err
	}

	_, err = database.GetDB().Exec(sql, args...)

	if err != nil {
		return err
	}

	return nil
}

func GetTaskById(id string) (*Task, error) {
	var task Task

	sql, args, err := sq.Select("*").
		From("tasks").
		Where(sq.Eq{"id": id}).
		Limit(1).
		ToSql()

	if err != nil {
		return nil, err
	}

	err = database.GetDB().Get(&task, sql, args...)

	if err != nil {
		return nil, err
	}

	return &task, nil
}

func GetTasksByExpressionId(expressionId string) ([]Task, error) {
	var tasks []Task

	sql, args, err := sq.Select("*").
		From("tasks").
		Where(sq.Eq{"expression_id": expressionId}).
		ToSql()

	if err != nil {
		return nil, err
	}

	err = database.GetDB().Select(&tasks, sql, args...)

	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func IsAllTasksCompleted(expressionId string) bool {
	sql, args, err := sq.Select("count(*)").
		From("tasks").
		Where(sq.And{
			sq.Eq{"expression_id": expressionId},
			sq.Eq{"completed": false},
		}).
		ToSql()

	if err != nil {
		return false
	}
	var count int
	err = database.GetDB().Get(&count, sql, args...)

	if err != nil {
		return false
	}

	return count == 0
}

func GetTasksForProcessing(redistributionDelay int) ([]Task, error) {
	var tasks []Task

	sql, args, err := sq.Select("*").
		From("tasks").
		Where(sq.And{
			sq.Or{
				sq.Eq{"is_processing": false},
				sq.Expr("strftime('%s','now') - strftime('%s', updated_at) > ?", redistributionDelay),
			},
			sq.Eq{"completed": false},
		}).
		ToSql()

	if err != nil {
		return nil, err
	}

	err = database.GetDB().Select(&tasks, sql, args...)

	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func GetTasksByIds(ids []string) ([]Task, error) {
	var tasks []Task

	sql, args, err := sq.Select("*").
		From("tasks").
		Where(sq.Eq{"id": ids}).
		ToSql()

	if err != nil {
		return nil, err
	}

	err = database.GetDB().Select(&tasks, sql, args...)

	if err != nil {
		return nil, err
	}

	return tasks, nil
}
