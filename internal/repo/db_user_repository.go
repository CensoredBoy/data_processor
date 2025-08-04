package repo

import (
	"context"
	"data_processor/internal/common"
	"errors"
	"github.com/jackc/pgx/v5"
)

var _ IUserRepository = (*PgxRepository)(nil)

func (r *PgxRepository) GetUserID(ctx context.Context, user *common.User) (*common.UserID, error) {
	var userId common.UserID
	query := `SELECT id FROM users WHERE name = $1`
	err := r.pool.QueryRow(ctx, query, user.Name).Scan(&userId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &userId, nil
}
func (r *PgxRepository) CreateUser(ctx context.Context, user *common.User) error {
	query := `INSERT INTO users (name, password) VALUES ($1, $2) RETURNING id`
	return r.pool.QueryRow(ctx, query, user.Name, user.Password).Scan(&user.ID)
}

func (r *PgxRepository) GetUserByID(ctx context.Context, id common.UserID) (*common.User, error) {
	user := &common.User{}
	query := `SELECT id, name, password FROM users WHERE id = $1`
	err := r.pool.QueryRow(ctx, query, id).Scan(&user.ID, &user.Name, &user.Password)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

func (r *PgxRepository) GetUserByName(ctx context.Context, name string) (*common.User, error) {
	user := &common.User{}
	query := `SELECT id, name, password FROM users WHERE name = $1`
	err := r.pool.QueryRow(ctx, query, name).Scan(&user.ID, &user.Name, &user.Password)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

func (r *PgxRepository) UpdateUser(ctx context.Context, user *common.User) error {
	query := `UPDATE users SET name = $1, password = $2 WHERE id = $3`
	_, err := r.pool.Exec(ctx, query, user.Name, user.Password, user.ID)
	return err
}

func (r *PgxRepository) DeleteUser(ctx context.Context, id common.UserID) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *PgxRepository) ListUsers(ctx context.Context) ([]*common.User, error) {
	query := `SELECT id, name, password FROM users`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (*common.User, error) {
		var user common.User
		err := row.Scan(&user.ID, &user.Name, &user.Password)
		return &user, err
	})
	if err != nil {
		return nil, err
	}
	return users, nil
}
