CREATE TYPE STATUS AS ENUM ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED');

CREATE TABLE IF NOT EXISTS orders (
                                      username VARCHAR(50) NOT NULL,
                                      _number VARCHAR(50) UNIQUE NOT NULL,
                                      _status STATUS NOT NULL,
                                      accrual FLOAT DEFAULT 0.0,
                                      uploaded_at TIMESTAMP
);