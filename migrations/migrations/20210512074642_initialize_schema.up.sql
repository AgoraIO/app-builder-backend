CREATE TABLE IF NOT EXISTS users (
    id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    identifier TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    user_name TEXT,
    email TEXT NOT NULL,
    CONSTRAINT unique_email unique (email)
);CREATE TABLE IF NOT EXISTS channels (
    id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    title TEXT NOT NULL,
    channel_name TEXT NOT NULL,
    channel_secret TEXT,
    host_passphrase TEXT NOT NULL,
    viewer_passphrase TEXT,
    recording_uid INT,
    recording_sid TEXT,
    recording_rid TEXT,
    dtmf TEXT
);CREATE TABLE IF NOT EXISTS tokens (
    id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    token_id TEXT,
    user_id INT,
    CONSTRAINT tokens_fkey FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);CREATE TABLE IF NOT EXISTS credentials (
    id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    code TEXT NOT NULL,
    access_token TEXT NOT NULL,
    refresh_token TEXT NOT NULL,
    token_type TEXT NOT NULL,
    expiry TIMESTAMP
);