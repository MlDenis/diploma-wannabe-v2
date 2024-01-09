CREATE TABLE IF NOT EXISTS _sessions (
                                         username VARCHAR (50) NOT NULL,
                                         token VARCHAR (100) UNIQUE NOT NULL,
                                         expires_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS userinfo (
                                        username VARCHAR (50) UNIQUE NOT NULL,
                                        _password VARCHAR (50)
);