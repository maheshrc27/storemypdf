CREATE TABLE users (
    id TEXT PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    hashed_password TEXT NOT NULL,
    verified BOOLEAN NOT NULL DEFAULT 0,
    created DATETIME NOT NULL,
    updated DATETIME NOT NULL
);
CREATE TABLE IF NOT EXISTS "to_deletes" (
    id TEXT PRIMARY KEY,
    file_id TEXT NOT NULL,
    delete_time DATETIME NOT NULL,
    created DATETIME NOT NULL,
    updated DATETIME NOT NULL
);
CREATE TABLE subscriptions (
    id TEXT PRIMARY KEY,
    paddle_subscription_id TEXT NOT NULL,
    paddle_plan_id TEXT NOT NULL,
    status TEXT NOT NULL,
    next_bill_date DATETIME NOT NULL,
    user_id TEXT NOT NULL,
    created DATETIME NOT NULL,
    updated DATETIME NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
CREATE TABLE files (
    id TEXT PRIMARY KEY,
    filename TEXT NOT NULL,
    description TEXT,
    file_type TEXT NOT NULL,
    size INTEGER NOT NULL,
    user_id TEXT NOT NULL,
    created DATETIME NOT NULL,
    updated DATETIME NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
CREATE TABLE api_keys (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    api_key_hash TEXT NOT NULL,
    user_id TEXT NOT NULL,
    active BOOLEAN NOT NULL DEFAULT 1,
    created DATETIME NOT NULL,
    updated DATETIME NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
