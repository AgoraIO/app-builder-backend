CREATE TABLE IF NOT EXISTS credentials (
    id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    code TEXT NOT NULL,
    access_token TEXT NOT NULL,
    refresh_token TEXT NOT NULL,
    token_type TEXT NOT NULL,
    expiry TIMESTAMP
);