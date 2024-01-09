package db

const (
	SaveSession    string = `INSERT INTO _sessions VALUES ($1, $2, $3);`
	GetUserInfo           = `SELECT * FROM userinfo WHERE username=$1;`
	GetOrder              = `SELECT * FROM orders WHERE username=$1 AND _number=$2;`
	SaveOrder             = `INSERT INTO orders VALUES ($1, $2, $3, $4, $5);`
	GetOrders             = `SELECT * FROM orders WHERE username=$1;`
	GetSessionUser        = `SELECT username FROM _sessions WHERE token=$1;`
	GetBalance            = `SELECT * FROM balances WHERE username=$1;`
	UpdateBalance  string = `UPDATE balances SET _current=$1, withdrawn=$2 WHERE username=$3;`
	GetWithdrawals        = `SELECT * FROM withdrawal WHERE username=$1;`
	SaveWithdrawal        = `INSERT INTO withdrawal VALUES ($1, $2, $3, $4);`
	UpdateOrder           = `UPDATE orders SET _status=$1, accrual=$2 WHERE username=$3 AND _number=$4;`
	GetSession            = `SELECT * FROM _sessions WHERE token=$1;`
	GetAllOrders          = `SELECT * FROM orders;`
	SaveUserInfo          = `INSERT INTO userinfo VALUES ($1, $2);`
	SaveBalance           = `INSERT INTO balances VALUES ($1, $2, $3);`
)
