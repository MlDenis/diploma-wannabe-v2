package db

import (
	"context"
	"database/sql"
	"github.com/MlDenis/diploma-wannabe-v2/internal/configuration"
	"go.uber.org/zap"
	"time"

	"github.com/MlDenis/diploma-wannabe-v2/internal/errors"
	"github.com/MlDenis/diploma-wannabe-v2/internal/models"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type IDBInterface interface {
	SaveUserInfo(*models.UserInfo, *zap.Logger) error
	GetUserInfo(*models.UserInfo, *zap.Logger) (*models.UserInfo, error)
	SaveSession(string, *models.Session, *zap.Logger) error
	GetSession(string, *zap.Logger) (*models.Session, error)
	GetOrder(string, string, *zap.Logger) (*models.Order, error)
	SaveOrder(*models.Order, *zap.Logger) error
	GetOrders(string, *zap.Logger) ([]*models.Order, error)
	GetUsernameByToken(string, *zap.Logger) (string, error)
	GetUserBalance(string, *zap.Logger) (*models.Balance, error)
	UpdateUserBalance(string, *models.Balance, *zap.Logger) (*models.Balance, error)
	GetWithdrawals(string, *zap.Logger) ([]*models.Withdrawal, error)
	SaveWithdrawal(*models.Withdrawal, *zap.Logger) error
	SaveUserBalance(string, *models.Balance, *zap.Logger) (*models.Balance, error)
	UpdateOrder(string, *models.AccrualResponse, *zap.Logger) error
	GetAllOrders() ([]*models.Order, error)
}

type Cursor struct {
	IDBInterface
}

func GetCursor(url string, logger *zap.Logger) (*Cursor, error) {
	cursor, err := NewCursor(url, logger)
	if err != nil {
		return nil, err
	}
	return &Cursor{cursor}, nil
}

type IDBCursor struct {
	IDBInterface
	DB      *sql.DB
	Context context.Context
	Logger  *zap.Logger
}

func RunMigrations(databaseURL string, logger *zap.Logger) error {
	m, err := migrate.New(
		"file://./migrations",
		databaseURL)
	if err != nil {
		logger.Info("Error creating migration: ", zap.String("", err.Error()))
		return errors.ErrDatabaseMigration
	}
	if err := m.Up(); err != nil {
		logger.Info("Error executing migration: ", zap.String("", err.Error()))
		return errors.ErrDatabaseMigration
	}
	logger.Info("Migrations successfully executed")
	return nil
}

func NewCursor(IDBURL string, logger *zap.Logger) (*IDBCursor, error) {
	db, err := sql.Open("pgx", IDBURL)
	if err != nil {
		logger.Info("Unable to connect to database: ", zap.String("", err.Error()))
		return nil, errors.ErrDatabaseUnreachable
	}
	n := &IDBCursor{
		DB:      db,
		Context: context.Background(),
	}
	if err := n.Ping(logger); err != nil {
		logger.Info("DB ping error", zap.String("", err.Error()))
		return nil, err
	}
	err = RunMigrations(IDBURL, logger)
	if err != nil {
		return nil, err
	}
	return n, nil
}

func (c *IDBCursor) Close() {
	err := c.DB.Close()
	if err != nil {
		return
	}
}

func (c *IDBCursor) Ping(logger *zap.Logger) error {
	ctx, cancel := context.WithTimeout(c.Context, configuration.IDBTIMEOUT*time.Second)
	defer cancel()
	if err := c.DB.PingContext(ctx); err != nil {
		logger.Info("Ping error, database unreachable?", zap.String("", err.Error()))
		return errors.ErrDatabaseUnreachable
	}
	return nil
}

func (c *IDBCursor) SaveSession(id string, session *models.Session, logger *zap.Logger) error {
	_, err := c.DB.ExecContext(c.Context, SaveSession, session.Username, session.Token, session.ExpiresAt)
	if err != nil {
		logger.Info("error inserting row to db: ", zap.String("", err.Error()))
		return err
	}
	return nil
}

func (c *IDBCursor) SaveUserInfo(info *models.UserInfo, logger *zap.Logger) error {
	_, err := c.DB.ExecContext(c.Context, SaveUserInfo, info.Username, info.Password)
	if err != nil {
		logger.Info("error inserting row into Userinfo: %e", zap.String("", err.Error()))
		return err
	}
	return nil
}

func (c *IDBCursor) GetUserInfo(info *models.UserInfo, logger *zap.Logger) (*models.UserInfo, error) {
	var row *sql.Row
	if row = c.DB.QueryRowContext(c.Context, GetUserInfo, info.Username); row.Err() != nil {
		logger.Info("error during getting user info from db: %e", zap.Error(row.Err()))
		return nil, row.Err()
	}
	foundInfo := &models.UserInfo{}
	err := row.Scan(&foundInfo.Username, &foundInfo.Password)
	if err != nil {
		logger.Info("error scanning userinfo from db", zap.String("", err.Error()))
		return nil, err
	}
	return foundInfo, nil
}

func (c *IDBCursor) GetOrder(username string, number string, logger *zap.Logger) (*models.Order, error) {
	var row *sql.Row
	if row = c.DB.QueryRowContext(c.Context, GetOrder, username, number); row.Err() != nil {
		logger.Info("error during getting order from db", zap.Error(row.Err()))
		return nil, row.Err()
	}
	foundOrder := &models.Order{}
	err := row.Scan(&foundOrder.Username, &foundOrder.Number, &foundOrder.Status, &foundOrder.Accrual, &foundOrder.UploadedAt)
	if err == sql.ErrNoRows {
		logger.Info("No rows found for order and user", zap.String("", number+""+username))
		return nil, nil
	}
	if err != nil {
		logger.Error("error scanning single order from db", zap.Error(err))
		return nil, err
	}
	return foundOrder, nil
}

func (c *IDBCursor) SaveOrder(order *models.Order, logger *zap.Logger) error {
	_, err := c.DB.ExecContext(c.Context, SaveOrder, order.Username, order.Number, order.Status, order.Accrual, order.UploadedAt)
	if err != nil {
		logger.Error("error during saving order to db", zap.Error(err))
		return err
	}
	return nil
}

func (c *IDBCursor) GetOrders(username string, logger *zap.Logger) ([]*models.Order, error) {
	rows, err := c.DB.QueryContext(c.Context, GetOrders, username)
	if err != nil {
		logger.Error("error during getting orders from db", zap.Error(err))
		return nil, err
	}
	if rows.Err() != nil {
		logger.Error("error during getting orders from db", zap.Error(rows.Err()))
		return nil, rows.Err()
	}
	foundOrders := []*models.Order{}
	for rows.Next() {
		var o models.Order
		if err = rows.Scan(&o.Username, &o.Number, &o.Status, &o.Accrual, &o.UploadedAt); err != nil {
			logger.Error("error scanning order from db", zap.Error(err))
			return foundOrders, err
		}
		foundOrders = append(foundOrders, &o)
	}
	return foundOrders, nil
}

func (c *IDBCursor) GetUsernameByToken(token string, logger *zap.Logger) (string, error) {
	var row *sql.Row
	if row = c.DB.QueryRowContext(c.Context, GetSessionUser, token); row.Err() != nil {
		logger.Error("error during getting current session user from db", zap.Error(row.Err()))
		return "", row.Err()
	}
	foundSession := &models.Session{}
	err := row.Scan(&foundSession.Username)
	if err != nil {
		logger.Error("error scanning session username from db", zap.Error(err))
		return "", err
	}
	return foundSession.Username, nil
}

func (c *IDBCursor) GetUserBalance(username string, logger *zap.Logger) (*models.Balance, error) {
	var row *sql.Row
	if row = c.DB.QueryRowContext(c.Context, GetBalance, username); row.Err() != nil {
		logger.Error("error during getting user balance from db", zap.Error(row.Err()))
		return nil, row.Err()
	}
	logger.Info("Getting balance for user", zap.String("", username))
	foundBalance := &models.Balance{}
	err := row.Scan(&foundBalance.User, &foundBalance.Current, &foundBalance.Withdrawn)
	if err == sql.ErrNoRows {
		return foundBalance, nil
	}
	if err != nil {
		logger.Error("error scanning balance from db", zap.Error(err))
		return nil, err
	}
	return foundBalance, nil
}

func (c *IDBCursor) SaveUserBalance(username string, newBalance *models.Balance, logger *zap.Logger) (*models.Balance, error) {
	_, err := c.DB.ExecContext(c.Context, SaveBalance, username, newBalance.Current, newBalance.Withdrawn)
	if err != nil {
		logger.Error("error during saving balance for user", zap.Error(err))
		return nil, err
	}
	logger.Info("Saved balance for", zap.String("", username))
	newBalance.User = username
	return newBalance, nil
}

func (c *IDBCursor) UpdateUserBalance(username string, newBalance *models.Balance, logger *zap.Logger) (*models.Balance, error) {
	_, err := c.DB.ExecContext(c.Context, UpdateBalance, newBalance.Current, newBalance.Withdrawn, username)
	if err != nil {
		logger.Error("error during updating balance", zap.Error(err))
		return nil, err
	}
	logger.Info("Balance updated for user", zap.String("", username))
	return newBalance, nil
}

func (c *IDBCursor) GetWithdrawals(username string, logger *zap.Logger) ([]*models.Withdrawal, error) {
	rows, err := c.DB.QueryContext(c.Context, GetWithdrawals, username)

	if err != nil {
		logger.Error("error during getting withdrawals from db: %e", zap.Error(err))
		return nil, err
	}
	if rows.Err() != nil {
		logger.Error("error during getting withdrawals from db: %e", zap.Error(rows.Err()))
		return nil, rows.Err()
	}
	foundWithdrawals := []*models.Withdrawal{}
	for rows.Next() {
		var w models.Withdrawal
		if err := rows.Scan(&w.User, &w.Order, &w.Sum, &w.ProcessedAt); err != nil {
			logger.Error("error scanning withdrawal from db", zap.Error(err))
			return foundWithdrawals, err
		}
		foundWithdrawals = append(foundWithdrawals, &w)
	}
	if err = rows.Err(); err != nil {
		return foundWithdrawals, err
	}
	return foundWithdrawals, nil
}

func (c *IDBCursor) SaveWithdrawal(withdrawal *models.Withdrawal, logger *zap.Logger) error {
	_, err := c.DB.ExecContext(c.Context, SaveWithdrawal, withdrawal.User, withdrawal.Order, withdrawal.Sum, withdrawal.ProcessedAt)
	if err != nil {
		logger.Error("error during saving withdrawal to db", zap.Error(err))
		return err
	}
	return nil
}

func (c *IDBCursor) UpdateOrder(username string, from *models.AccrualResponse, logger *zap.Logger) error {
	var status string
	if from.Status == configuration.REGISTERED {
		status = configuration.PROCESSING
	} else {
		status = from.Status
	}
	_, err := c.DB.ExecContext(c.Context, UpdateOrder, status, from.Accrual, username, from.Order)
	if err != nil {
		logger.Error("error during updating order: %e", zap.Error(err))
		return err
	}
	return nil
}

func (c *IDBCursor) GetSession(token string, logger *zap.Logger) (*models.Session, error) {
	var row *sql.Row
	if row = c.DB.QueryRowContext(c.Context, GetSession, token); row.Err() != nil {
		logger.Error("error during getting user session from db: %e", zap.Error(row.Err()))
		return nil, row.Err()
	}
	foundSession := &models.Session{}

	err := row.Scan(&foundSession.Username, &foundSession.Token, &foundSession.ExpiresAt)
	if err != nil {
		logger.Error("error scanning session from db", zap.String("", err.Error()))
		return nil, err
	}
	return foundSession, nil
}

func (c *IDBCursor) GetAllOrders() ([]*models.Order, error) {
	rows, err := c.DB.QueryContext(c.Context, GetAllOrders)

	if err != nil {
		c.Logger.Error("error during getting all orders from db: %e", zap.String("", err.Error()))
		return nil, err
	}
	if rows.Err() != nil {
		c.Logger.Error("error during getting all orders from db: %e", zap.Error(rows.Err()))
		return nil, rows.Err()
	}
	foundOrders := []*models.Order{}
	for rows.Next() {
		var o models.Order
		if err = rows.Scan(&o.Username, &o.Number, &o.Status, &o.Accrual, &o.UploadedAt); err != nil {
			c.Logger.Error("error scanning order among orders from db: %e", zap.String("", err.Error()))
			return foundOrders, nil
		}
		foundOrders = append(foundOrders, &o)
	}
	return foundOrders, nil
}
