package api

import (
	"github.com/go-chi/chi/v5"

	"github.com/MlDenis/diploma-wannabe-v2/internal/db"
	"github.com/MlDenis/diploma-wannabe-v2/internal/jobmanager"
)

type UserRouter struct {
	*chi.Mux
	Cursor *db.Cursor
}

type OrderRouter struct {
	*chi.Mux
	Cursor  *db.Cursor
	Manager *jobmanager.Jobmanager
}

type BalanceRouter struct {
	*chi.Mux
	Cursor *db.Cursor
}

type Handler struct {
	*chi.Mux
	Cursor *db.Cursor
}

func NewHandler(cursor *db.Cursor, manager *jobmanager.Jobmanager) *Handler {
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

		OrdersRouter := NewOrdersRouter(cursor, manager)
		r.Mount("/orders", OrdersRouter)
	})

	return handler
}

func NewOrdersRouter(cursor *db.Cursor, manager *jobmanager.Jobmanager) *OrderRouter {
	r := &OrderRouter{
		Mux:     chi.NewMux(),
		Cursor:  cursor,
		Manager: manager,
	}
	r.Post("/", r.UploadOrder)
	r.Get("/", r.GetOrders)
	return r
}
