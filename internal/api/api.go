package api

import (
	"github.com/MlDenis/diploma-wannabe-v2/internal/db"
	"github.com/MlDenis/diploma-wannabe-v2/internal/jobmanager"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type UserRouter struct {
	*chi.Mux
	Cursor *db.Cursor
}

type OrderRouter struct {
	*chi.Mux
	Cursor  *db.Cursor
	Manager *jobmanager.Jobmanager
	Logger  *zap.Logger
}

type BalanceRouter struct {
	*chi.Mux
	Cursor *db.Cursor
	Logger *zap.Logger
}

type Handler struct {
	*chi.Mux
	Cursor *db.Cursor
}

func NewHandler(cursor *db.Cursor, manager *jobmanager.Jobmanager, l *zap.Logger) *Handler {
	handler := &Handler{
		Mux:    chi.NewMux(),
		Cursor: cursor,
	}
	handler.Use(GzipHandle)
	handler.Use(handler.CookieHandle)

	userRouter := &UserRouter{
		Mux:    chi.NewMux(),
		Cursor: cursor,
	}

	balanceRouter := &BalanceRouter{
		Mux:    chi.NewMux(),
		Cursor: cursor,
	}

	handler.Route("/api/user", func(r chi.Router) {

		r.Post("/register", userRouter.RegisterUser)
		r.Post("/login", userRouter.Login)

		r.Get("/withdrawals", balanceRouter.GetWithdrawals)
		r.Get("/balance", balanceRouter.GetBalance)
		r.Post("/balance/withdraw", balanceRouter.WithdrawMoney)

		OrdersRouter := NewOrdersRouter(cursor, manager, l)
		r.Mount("/orders", OrdersRouter)
	})

	return handler
}

func NewOrdersRouter(cursor *db.Cursor, manager *jobmanager.Jobmanager, l *zap.Logger) *OrderRouter {
	r := &OrderRouter{
		Mux:     chi.NewMux(),
		Cursor:  cursor,
		Manager: manager,
	}
	r.Post("/", r.UploadOrder)
	r.Get("/", r.GetOrders)
	return r
}
