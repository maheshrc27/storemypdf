package database

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type Subscription struct {
	ID                   int       `db:"id"`
	PaddleSubscriptionID string    `db:"paddle_subscription_id"`
	PaddlePlanID         string    `db:"paddle_plan_id"`
	Status               string    `db:"status"`
	NextBillDate         time.Time `db:"next_bill_date"`
	UserID               int       `db:"user_id"`
	Created              time.Time `db:"created"`
	Updated              time.Time `db:"updated"`
}

func (db *DB) InsertSubscription(paddleSubscriptionID, paddlePlanID, status string, nextBillDate time.Time, userID int) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `
		INSERT INTO subscriptions (paddle_subscription_id, paddle_plan_id, status, next_bill_date, user_id)
		VALUES ($1, $2, $3, $4, $5)`

	result, err := db.ExecContext(ctx, query, paddleSubscriptionID, paddlePlanID, status, nextBillDate, userID)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), err
}

func (db *DB) GetSubscriptionByID(paddleSubscriptionID string) (*User, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var user User

	query := `SELECT * FROM subscriptions WHERE paddle_subscription_id = $1`

	err := db.GetContext(ctx, &user, query, paddleSubscriptionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}

	return &user, true, err
}

func (db *DB) GetSubscriptionByUserID(userID int) (*User, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var user User

	query := `SELECT * FROM subscriptions WHERE user_id = $1`

	err := db.GetContext(ctx, &user, query, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return &user, true, err
}

func (db *DB) UpdateSubscriptionStatus(status, paddleSubscriptionID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `UPDATE subscriptions SET status = $1, updated = $2 WHERE paddle_subscription_id = $3`

	_, err := db.ExecContext(ctx, query, status, time.Now(), paddleSubscriptionID)
	return err
}

func (db *DB) UpdateSubscriptionStatusAndNextBill(status string, nextBillDate time.Time, paddleSubscriptionID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `UPDATE subscriptions SET status = $1, next_bill_date = $2, updated = $3 WHERE paddle_subscription_id = $4`

	_, err := db.ExecContext(ctx, query, status, nextBillDate, time.Now(), paddleSubscriptionID)
	return err
}
