CREATE TABLE IF NOT EXISTS users (
    id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    identifier TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE,
    user_name TEXT,
    email TEXT NOT NULL,
    CONSTRAINT unique_email unique (email)
);CREATE TABLE IF NOT EXISTS channels (
    id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE,
    title TEXT,
    channel_name TEXT,
    channel_secret TEXT,
    host_passphrase TEXT,
    viewer_passphrase TEXT,
    recording_uid INT,
    recording_sid TEXT,
    recording_rid TEXT,
    dtmf TEXT
);CREATE TABLE IF NOT EXISTS tokens (
    id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE,
    token_id TEXT,
    user_id INT,
    CONSTRAINT tokens_fkey FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);