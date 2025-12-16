package user

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}



func (r *UserRepository) CreateUser(name, email, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(`
		INSERT INTO users (name, email, password, balance_tjs, balance_usd, balance_eur, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, name, email, string(hash), 100.0, 0.0, 0.0, time.Now())
	if err != nil{
		return err
	}
	user_ID, err := r.GetByEmail(email)
	if err != nil{
		return err
	}
	err = r.CreateProfile(user_ID.ID, name)
	if err != nil{
		return err
	}
	return nil
}

func (r *UserRepository) GetByEmail(email string) (*User, error) {
	u := &User{}
	row := r.db.QueryRow(`
		SELECT id, name, email, password, balance_tjs, balance_usd, balance_eur, created_at 
		FROM users WHERE email = $1
	`, email)
	err := row.Scan(&u.ID, &u.Name, &u.Email, &u.Password, &u.BalanceTJS, &u.BalanceUSD, &u.BalanceEUR, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return u, nil
}

func (r *UserRepository) CheckPassword(u *User, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

func (r *UserRepository) SaveToken(userID int, token string) error {
	_, err := r.db.Exec(`
		INSERT INTO user_tokens(token, user_id, created_at)
		VALUES($1, $2, $3)
	`, token, userID, time.Now())
	return err
}

func (r *UserRepository) GetUserIDByToken(token string) (int, error) {
	var userID int
	err := r.db.QueryRow(`SELECT user_id FROM user_tokens WHERE token=$1`, token).Scan(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}
func (r *UserRepository) CreateProfile(id int, name string) error{
	_, err := r.db.Exec(`
		INSERT INTO profiles(user_id, full_name, bio, avatar_path, updated_at)
		VALUES($1, $2, $3, $4, $5)
	`, id, name, "Расскажите о себе", "uploads/default-avatar.jpg", time.Now())
	return err
}

func (r *UserRepository)UpdateProfile(name, bio, avatar_path string, id int) error{
	
	_, err := r.db.Exec(`
		UPDATE profiles SET full_name = $1, bio = $2, avatar_path = $3, updated_at = $4
		WHERE user_id=$5
	`, name, bio, avatar_path, time.Now(), id)
	if err != nil{
		return err
	}
	_, err = r.db.Exec(`
		UPDATE users SET name = $1
		WHERE id=$2
	`, name, id)
	if err != nil{
		return err
	}
	return nil
}

func(r *UserRepository) GetProfile(id int) (*AboutPerson, error){
	p := &AboutPerson{}
	row := r.db.QueryRow(`
	SELECT full_name, bio, avatar_path
	FROM profiles
	WHERE user_id=$1
	`, id)
	err := row.Scan(&p.Full_name, &p.Bio, &p.Avatar_path)
	if err != nil{
		return nil, err
	}
	return p, nil
}
func(r *UserRepository) GetAvatar_path(id int) (string, error){
	var p string
	row := r.db.QueryRow(`
	SELECT avatar_path
	FROM profiles
	WHERE user_id=$1
	`, id)
	_ = row.Scan(&p)
	return p, nil
}

func (r *UserRepository) GetUserByID(id int) (*User, error) {
	u := &User{}
	row := r.db.QueryRow(`
		SELECT id, name, email, balance_tjs, balance_usd, balance_eur 
		FROM users 
		WHERE id=$1
	`, id)
	err := row.Scan(&u.ID, &u.Name, &u.Email, &u.BalanceTJS, &u.BalanceUSD, &u.BalanceEUR)
	if err != nil {
		return nil, err
	}
	return u, nil
}
func (r *UserRepository) GetTransactionsByID(userID int) ([]*Transactions, error) {
	rows, err := r.db.Query(`
		SELECT type, amount, currency, description, created_at
		FROM transactions
		WHERE user_id = $1
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var transactions []*Transactions
	for rows.Next() {
		t := &Transactions{}
		if err := rows.Scan(&t.TType, &t.Amount, &t.Currency, &t.Description, &t.CreatedAt); err != nil {
			return nil, err
		}
		transactions = append(transactions, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	
	return transactions, nil
}



func (r *UserRepository) GetAllUsersExcept(excludeID int) ([]*User, error) {
	rows, err := r.db.Query("SELECT id, name FROM users WHERE id != $1", excludeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		u := &User{}
		if err := rows.Scan(&u.ID, &u.Name); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *UserRepository) Deposit(userID int, amount float64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(`UPDATE users SET balance_tjs = balance_tjs + $1 WHERE id=$2`, amount, userID)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(`
		INSERT INTO transactions (user_id, type, amount, currency, description, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, userID, "deposit", amount, "TJS", "Пополнение счета", time.Now())
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (r *UserRepository) Transfer(fromID, toID int, amount float64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	var senderBalance float64
	err = tx.QueryRow(`SELECT balance_tjs FROM users WHERE id=$1`, fromID).Scan(&senderBalance)
	if err != nil {
		tx.Rollback()
		return err
	}
	if senderBalance < amount {
		tx.Rollback()
		return errors.New("insufficient funds")
	}

	_, err = tx.Exec(`UPDATE users SET balance_tjs = balance_tjs - $1 WHERE id=$2`, amount, fromID)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(`UPDATE users SET balance_tjs = balance_tjs + $1 WHERE id=$2`, amount, toID)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(`
		INSERT INTO transactions (user_id, type, amount, currency, description, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, fromID, "transfer", -amount, "TJS", "Перевод пользователю "+fmt.Sprint(toID), time.Now())
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(`
		INSERT INTO transactions (user_id, type, amount, currency, description, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, toID, "transfer", amount, "TJS", "Получено от пользователя "+fmt.Sprint(fromID), time.Now())
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}


func (r *UserRepository) ConvertCurrency(userID int, from, to string, amount, rate float64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(`UPDATE users SET balance_`+from+` = balance_`+from+` - $1 WHERE id=$2`, amount, userID)
	if err != nil {
		tx.Rollback()
		return err
	}

	converted := amount * rate
	_, err = tx.Exec(`UPDATE users SET balance_`+to+` = balance_`+to+` + $1 WHERE id=$2`, converted, userID)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(`
		INSERT INTO transactions (user_id, type, amount, currency, description, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, userID, "conversion", amount, from, "Конвертация в "+to, time.Now())
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}
