package data

import (
	"database/sql"
	"time"
)

type PGTestRepository struct {
	Conn *sql.DB
}

func NewPGTestRepository(db *sql.DB) *PGTestRepository {
	return &PGTestRepository{db}
}

func (u *PGTestRepository) GetAll() ([]*User, error) {
	users := []*User{}
	return users, nil
}

func (u *PGTestRepository) GetByEmail(email string) (*User, error) {
	user := User{
		ID:        1,
		FirstName: "First",
		LastName:  "Last",
		Email:     "me@here.com",
		Password:  "password",
		Active:    1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return &user, nil
}

func (u *PGTestRepository) GetOne(id int) (*User, error) {
	user := User{
		ID:        1,
		FirstName: "First",
		LastName:  "Last",
		Email:     "me@here.com",
		Password:  "password",
		Active:    1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return &user, nil
}

func (u *PGTestRepository) Update(user User) error {
	return nil
}

func (u *PGTestRepository) DeleteById(id int) error {
	return nil
}

func (u *PGTestRepository) Insert(user User) (int, error) {
	return 2, nil
}

func (u *PGTestRepository) ResetPassword(password string, user User) error {
	return nil
}

func (u *PGTestRepository) PasswordMatches(plainText string, user User) (bool, error) {
	return true, nil
}
