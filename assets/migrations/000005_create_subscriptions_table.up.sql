CREATE TABLE subscriptions(
    id TEXT NOT NULL PRIMARY KEY,
    paddle_subscription_id TEXT NOT NULL,
	paddle_plan_id TEXT NOT NULL,
	status TEXT NOT NULL,
	next_bill_date DATETIME NOT NULL,
    user_id int NOT NULL,
    created DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
