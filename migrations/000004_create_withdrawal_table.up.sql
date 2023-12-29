CREATE TABLE IF NOT EXISTS withdrawal (
                                          username VARCHAR(50) NOT NULL,
                                          _order VARCHAR(200) NOT NULL UNIQUE,
                                          _sum FLOAT NOT NULL DEFAULT 0.0,
                                          processed_at TIMESTAMP NOT NULL
);