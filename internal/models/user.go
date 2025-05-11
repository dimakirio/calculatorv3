package models

import (
	"database/sql"
	"errors"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       int64
	Login    string
	Password string
}

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(login, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = r.db.Exec("INSERT INTO users (login, password_hash) VALUES (?, ?)",
		login, string(hashedPassword))
	if err != nil {
		return err
	}

	return nil
}

func (r *UserRepository) GetByLogin(login string) (*User, error) {
	var user User
	var hashedPassword string
	err := r.db.QueryRow("SELECT id, login, password_hash FROM users WHERE login = ?",
		login).Scan(&user.ID, &user.Login, &hashedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	user.Password = hashedPassword
	return &user, nil
}

func (r *UserRepository) ValidatePassword(user *User, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	return err == nil
}
