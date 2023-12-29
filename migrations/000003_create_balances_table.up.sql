CREATE TABLE IF NOT EXISTS balances (
                                        username VARCHAR(50) UNIQUE,
                                        _current FLOAT NOT NULL DEFAULT 0.0,
                                        withdrawn FLOAT NOT NULL DEFAULT 0.0
);