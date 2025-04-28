package model

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/raikh/calc_micro_final/internal/database"

	sq "github.com/Masterminds/squirrel"
)

type Expression struct {
	Id         string     `json:"id" db:"id"`
	UserId     int64      `json:"-" db:"user_id"`
	Expression string     `json:"expression" db:"expression"`
	Status     string     `json:"status" db:"status"`
	Result     *float64   `json:"result" db:"result"`
	CreatedAt  *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  *time.Time `json:"updated_at" db:"updated_at"`
	DeletedAt  *time.Time `json:"-" db:"deleted_at"`
}

func BeginTx() (*sqlx.Tx, error) {
	return database.GetDB().Beginx()
}

func (e *Expression) buildInsertExpression() (string, []interface{}, error) {
	sql, args, err := sq.Insert("expressions").
		Columns("id", "user_id", "expression", "status", "result", "created_at", "updated_at").
		Values(e.Id, e.UserId, e.Expression, e.Status, e.Result, e.CreatedAt, e.UpdatedAt).
		ToSql()

	if err != nil {
		return "", nil, err
	}

	return sql, args, nil
}

func (e *Expression) InsertTx(tx *sqlx.Tx) error {
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

func (e *Expression) Insert() error {
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

func (e *Expression) Update() error {
	now := time.Now()
	sql, args, err := sq.Update("expressions").
		Set("expression", e.Expression).
		Set("status", e.Status).
		Set("result", e.Result).
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

func GetExpressionById(id string) (Expression, error) {
	var expression Expression

	query, args, err := sq.Select("*").
		From("expressions").
		Where(sq.Eq{"id": id}).
		Limit(1).
		ToSql()

	if err != nil {
		return Expression{}, fmt.Errorf("failed to build query: %w", err)
	}

	err = database.GetDB().Get(&expression, query, args...)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Expression{}, fmt.Errorf("expression with id %s not found", id)
		}
		return Expression{}, fmt.Errorf("failed to get expression: %w", err)
	}

	return expression, nil
}

func GetExpressionByIdForUser(id string, userId int64) (Expression, error) {
	var expression Expression

	query, args, err := sq.Select("*").
		From("expressions").
		Where(sq.Eq{"id": id}).
		Where(sq.Eq{"user_id": userId}).
		Limit(1).
		ToSql()

	if err != nil {
		return Expression{}, fmt.Errorf("failed to build query: %w", err)
	}

	err = database.GetDB().Get(&expression, query, args...)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Expression{}, fmt.Errorf("expression with id %s not found", id)
		}
		return Expression{}, fmt.Errorf("failed to get expression: %w", err)
	}

	return expression, nil
}

func GetExpressionsByUserId(userId int64) ([]Expression, error) {
	var expressions []Expression

	query, args, err := sq.Select("*").
		From("expressions").
		Where(sq.Eq{"user_id": userId}).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	err = database.GetDB().Select(&expressions, query, args...)

	if err != nil {
		return nil, fmt.Errorf("failed to get expressions: %w", err)
	}

	return expressions, nil
}
