package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type Storage interface {
	CreateAccount(*Account) (int, error)
	UpdateAccount(*Account) error
	DeleteAccount(int) error
	GetAccountByNumber(int) (*Account, error)
	GetAccountByID(int) (*Account, error)
	GetAllAccount() ([]*Account, error)
}

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage() (*PostgresStorage, error) {
	connStr := "user=postgres dbname=postgres password=godb sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresStorage{
		db: db,
	}, nil
}

func (s *PostgresStorage) Init() error {
	return s.createAccountTable()
}

func (s *PostgresStorage) createAccountTable() error {
	createAccountTableQuery := `CREATE TABLE IF NOT EXISTS ACCOUNT (
	  id serial primary key,
	  first_name varchar(50),
	  last_name varchar(50),
		password varchar(100) not null,
	  account_number serial UNIQUE,
	  account_balance numeric(20,2),
	  created_at timestamp
	)`

	_, err := s.db.Exec(createAccountTableQuery)
	return err
}

func (s *PostgresStorage) CreateAccount(acc *Account) (int, error) {
	query := `
	INSERT INTO ACCOUNT
	(first_name, last_name, password, account_balance, created_at)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id
	`

	hashedPassword, bcryptErr := bcrypt.GenerateFromPassword([]byte(acc.Password), bcrypt.DefaultCost)
	if bcryptErr != nil {
		return 0, fmt.Errorf("error hashing password")
	}

	acc.Password = string(hashedPassword)

	var id int

	err := s.db.QueryRow(
		query,
		acc.FirstName,
		acc.LastName,
		acc.Password,
		acc.Balance,
		acc.CreatedAt,
	).Scan(&id)

	return id, err
}

func (s *PostgresStorage) UpdateAccount(acc *Account) error {
	updates := []string{}
	args := []interface{}{}
	argID := 1

	if acc.FirstName != "" {
		updates = append(updates, "first_name = $"+fmt.Sprint(argID))
		args = append(args, acc.FirstName)
		argID++
	}
	if acc.LastName != "" {
		updates = append(updates, "last_name = $"+fmt.Sprint(argID))
		args = append(args, acc.LastName)
		argID++
	}

	if len(updates) == 0 {
		return nil // No updates to make
	}

	query := fmt.Sprintf(`
		UPDATE ACCOUNT
		SET %s
		WHERE id = $%d
	`, strings.Join(updates, ", "), argID)

	args = append(args, acc.ID)

	_, err := s.db.Exec(query, args...)
	if err != nil {
		log.Println("Error executing query:", err)
	}
	return err
}

func (s *PostgresStorage) DeleteAccount(id int) error {
	query := `DELETE FROM ACCOUNT WHERE id = $1`
	_, err := s.db.Exec(query, id)
	if err != nil {
		return err
	}

	return nil
}

func (s *PostgresStorage) GetAccountByID(id int) (*Account, error) {
	query := `
	SELECT * FROM ACCOUNT WHERE id = $1
	`

	row, err := s.db.Query(query, id)
	if err != nil {
		return nil, err
	}

	for row.Next() {
		return scanIntoAccount(row)
	}

	return nil, fmt.Errorf("Account not found")
}

func (s *PostgresStorage) GetAccountByNumber(number int) (*Account, error) {
	query := `
	SELECT * FROM ACCOUNT
	WHERE account_number = $1
	`

	row, err := s.db.Query(query, number)
	if err != nil {
		return nil, err
	}

	for row.Next() {
		return scanIntoAccount(row)
	}

	return nil, fmt.Errorf("Account not found")
}

func (s *PostgresStorage) GetAllAccount() ([]*Account, error) {
	query := `
	SELECT * FROM ACCOUNT
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}

	accounts := []*Account{}
	for rows.Next() {
		account, err := scanIntoAccount(rows)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}

	return accounts, nil
}

func scanIntoAccount(rows *sql.Rows) (*Account, error) {
	account := &Account{}
	if err := rows.Scan(
		&account.ID,
		&account.FirstName,
		&account.LastName,
		&account.Password,
		&account.Number,
		&account.Balance,
		&account.CreatedAt,
	); err != nil {
		return nil, err
	}

	return account, nil
}
