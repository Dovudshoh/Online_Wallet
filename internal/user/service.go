package user

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type UserService struct {
	repo   *UserRepository
	apiKey string
}

func (s *UserService) GetAllUsersExcept(excludeID int) ([]*User, error) {
	return s.repo.GetAllUsersExcept(excludeID)
}


func NewUserService(repo *UserRepository, apiKey string) *UserService {
	return &UserService{repo: repo, apiKey: apiKey}
}

func (s *UserService) Register(name, email, password string) error {
	existing, _ := s.repo.GetByEmail(email)
	if existing != nil {
		return errors.New("user already exists")
	}
	return s.repo.CreateUser(name, email, password)
}

func (s *UserService) Login(email, password string) (string, error) {
	u, err := s.repo.GetByEmail(email)
	if err != nil {
		return "", err
	}
	if u == nil {
		return "", errors.New("user not found")
	}

	if !s.repo.CheckPassword(u, password) {
		return "", errors.New("invalid password")
	}

	token := generateToken()

	// Сохраняем токен
	err = s.repo.SaveToken(u.ID, token)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *UserService) GetUserIDByToken(token string) (int, error) {
	return s.repo.GetUserIDByToken(token)
}
func (s *UserService) GetBalance(userID int) (*User, error) {
	return s.repo.GetUserByID(userID)
}

func (s *UserService) Deposit(userID int, amount float64) error {
	if amount <= 0 {
		return errors.New("amount must be positive")
	}
	return s.repo.Deposit(userID, amount)
}

func (s *UserService) Transfer(fromID, toID int, amount float64) error {
	if amount <= 0 {
		return errors.New("amount must be positive")
	}
	return s.repo.Transfer(fromID, toID, amount)
}


func (s *UserService) ConvertCurrency(userID int, from string, to string, amount float64, rate float64) error {
	if amount <= 0 {
		return errors.New("amount must be positive")
	}
	if rate <= 0 {
		return errors.New("rate must be positive")
	}
	return s.repo.ConvertCurrency(userID, from, to, amount, rate)
}

func (s *UserService) GetCurrencyRate(from, to string) (float64, error) {

    if from == to {
        return 1.0, nil
    }


    const tjsToUsd = 0.092
    const usdToTjs = 1 / tjsToUsd

    if from == "TJS" && to != "TJS" {
        usdRate, err := s.getRateFromAPI("USD", to)
        if err != nil {
            return 0, err
        }
        return tjsToUsd * usdRate, nil
    }
    if from != "TJS" && to == "TJS" {
        usdRate, err := s.getRateFromAPI(from, "USD")
        if err != nil {
            return 0, err
        }
        return usdRate * usdToTjs, nil
    }

    if from == "TJS" && to == "TJS" {
        return 1.0, nil
    }

    return s.getRateFromAPI(from, to)
}


func (s *UserService) getRateFromAPI(from, to string) (float64, error) {
    if s.apiKey == "" {
        return 0, fmt.Errorf("open exchange rates API key not set")
    }

    url := fmt.Sprintf("https://openexchangerates.org/api/latest.json?app_id=%s&symbols=%s,%s", s.apiKey, from, to)
    resp, err := http.Get(url)
    if err != nil {
        return 0, err
    }
    defer resp.Body.Close()

    var data struct {
        Rates map[string]float64 `json:"rates"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
        return 0, err
    }

    rateFrom, ok1 := data.Rates[from]
    rateTo, ok2 := data.Rates[to]
    if !ok1 || !ok2 {
        return 0, fmt.Errorf("currency not found in API response")
    }

    return rateTo / rateFrom, nil
}


func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}
