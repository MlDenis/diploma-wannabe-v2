package mocks

import (
	"github.com/MlDenis/diploma-wannabe-v2/internal/db"
	"github.com/MlDenis/diploma-wannabe-v2/internal/errors"
	"github.com/MlDenis/diploma-wannabe-v2/internal/models"
	"go.uber.org/zap"

	"github.com/go-chi/chi/v5"
)

type MockDB struct {
	db.IDBInterface
	storage     map[string]string
	sessions    map[string]models.Session
	orders      map[string][]*models.Order
	balance     map[string]*models.Balance
	withdrawals map[string][]*models.Withdrawal
}

type TestHandler struct {
	*chi.Mux
	DBAPI  *db.Cursor
	Cursor *db.Cursor
}

func NewMock() *MockDB {
	return &MockDB{
		storage:     make(map[string]string),
		sessions:    make(map[string]models.Session),
		orders:      make(map[string][]*models.Order),
		balance:     make(map[string]*models.Balance),
		withdrawals: make(map[string][]*models.Withdrawal),
	}
}

func (mock *MockDB) SaveSession(id string, session *models.Session, l *zap.Logger) error {
	mock.sessions[id] = *session
	return nil
}

func (mock *MockDB) SaveUserInfo(info *models.UserInfo, l *zap.Logger) error {

	for k := range mock.storage {
		if k == info.Username {
			return errors.ErrDatabaseSQLQuery
		}
	}

	mock.storage[info.Username] = info.Password
	return nil
}

func (mock *MockDB) GetUserInfo(info *models.UserInfo, l *zap.Logger) (*models.UserInfo, error) {
	for k, v := range mock.storage {
		if k == info.Username {
			return &models.UserInfo{
				Username: k,
				Password: v,
			}, nil
		}
	}
	return nil, errors.ErrValidation
}

func (mock *MockDB) GetOrder(username string, number string, l *zap.Logger) (*models.Order, error) {
	for user, orders := range mock.orders {
		if user == username {
			for _, order := range orders {
				if order.Number == number {
					return order, nil
				}
			}
		}
	}

	return nil, nil
}

func (mock *MockDB) SaveOrder(order *models.Order, l *zap.Logger) error {
	mock.orders[order.Username] = append(mock.orders[order.Username], order)
	return nil
}

func (mock *MockDB) GetOrders(username string, l *zap.Logger) ([]*models.Order, error) {
	if len(mock.orders[username]) == 0 {
		return nil, nil
	}
	return mock.orders[username], nil
}

func (mock *MockDB) GetUsernameByToken(token string, l *zap.Logger) (string, error) {
	session, ok := mock.sessions[token]
	if !ok {
		return "", errors.ErrValidation
	}
	return session.Username, nil
}

func (mock *MockDB) GetUserBalance(username string, l *zap.Logger) (*models.Balance, error) {
	balance, ok := mock.balance[username]
	if !ok {
		return nil, errors.ErrValidation
	}
	return balance, nil
}

func (mock *MockDB) UpdateUserBalance(username string, newBalance *models.Balance, l *zap.Logger) (*models.Balance, error) {
	mock.balance[username] = newBalance
	return newBalance, nil
}

func (mock *MockDB) GetWithdrawals(username string, l *zap.Logger) ([]*models.Withdrawal, error) {
	return mock.withdrawals[username], nil
}

func (mock *MockDB) SaveWithdrawal(withdrawal *models.Withdrawal, l *zap.Logger) error {
	mock.withdrawals[withdrawal.User] = append(mock.withdrawals[withdrawal.User], withdrawal)
	return nil
}

func (mock *MockDB) UpdateOrder(username string, from *models.AccrualResponse, l *zap.Logger) error {
	orders := mock.orders[username]
	for _, order := range orders {
		if order.Number == from.Order {
			if from.Status == "REGISTERED" {
				order.Status = "PROCESSING"
				break
			}
			order.Accrual = from.Accrual
			order.Status = from.Status
			break
		}
	}
	return nil
}

func (mock *MockDB) GetSession(token string, l *zap.Logger) (*models.Session, error) {
	session, ok := mock.sessions[token]
	if !ok {
		return nil, errors.ErrDatabaseSQLQuery
	}
	return &session, nil
}

func (mock *MockDB) GetAllOrders() ([]*models.Order, error) {
	result := make([]*models.Order, 0)
	for _, orders := range mock.orders {
		result = append(result, orders...)
	}
	return result, nil
}
