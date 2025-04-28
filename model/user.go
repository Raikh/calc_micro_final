package model

import (
	"database/sql"
	"errors"
	"time"

	"github.com/raikh/calc_micro_final/internal/database"

	sq "github.com/Masterminds/squirrel"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id        int64      `json:"id" db:"id"`
	Email     string     `json:"email" db:"email"`
	Password  string     `json:"-" db:"password"`
	CreatedAt *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt *time.Time `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"-" db:"deleted_at"`
}

func (u *User) GeneratePasswordHarsh() error {
	bytes, err := bcrypt.GenerateFromPassword([]byte(u.Password), 14)
	u.Password = string(bytes)
	return err
}

func (u *User) CheckPasswordHarsh(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

func GetByEmail(email string) (User, error) {
	var user User

	sql, args, err := sq.Select("*").
		From("users").
		Where(sq.Eq{"email": email}).
		Limit(1).
		ToSql()

	if err != nil {
		return User{}, err
	}

	err = database.GetDB().Get(&user, sql, args...)

	if err != nil {
		return User{}, err
	}

	return user, nil
}

func GetById(id int64) (User, error) {
	var user User

	sql, args, err := sq.Select("*").
		From("users").
		Where(sq.Eq{"id": id}).
		Limit(1).
		ToSql()

	if err != nil {
		return User{}, err
	}

	err = database.GetDB().Get(&user, sql, args...)

	if err != nil {
		return User{}, err
	}

	return user, nil
}

func Create(email string, password string) (*User, error) {
	user, err := GetByEmail(email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	if user.Id != 0 {
		return nil, errors.New("email already exists")
	}

	now := time.Now()
	newUser := &User{
		Email:     email,
		Password:  password,
		CreatedAt: &now,
		UpdatedAt: &now,
	}
	newUser.GeneratePasswordHarsh()

	err = newUser.insert()

	if err != nil {
		return nil, err
	}

	return newUser, nil
}

func (u *User) insert() error {
	sql, args, err := sq.Insert("users").
		Columns("email", "password", "created_at", "updated_at", "deleted_at").
		Values(u.Email, u.Password, u.CreatedAt, u.UpdatedAt, u.DeletedAt).
		ToSql()

	if err != nil {
		return err
	}

	res, err := database.GetDB().Exec(sql, args...)

	if err != nil {
		return err
	}

	id, err := res.LastInsertId()

	if err != nil {
		return err
	}

	u.Id = id

	return nil
}
