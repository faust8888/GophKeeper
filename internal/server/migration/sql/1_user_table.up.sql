CREATE TABLE IF NOT EXISTS "user" (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    login TEXT NOT NULL UNIQUE CHECK (char_length(login) <= 255),
    password_hash TEXT NOT NULL CHECK (char_length(password_hash) <= 1024),
    salt TEXT NOT NULL CHECK (char_length(salt) <= 255)
);