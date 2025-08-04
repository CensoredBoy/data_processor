package repo

import (
	"context"
	"data_processor/internal/common"
	"errors"
	"github.com/jackc/pgx/v5"
)

var _ IApplicationRepository = (*PgxRepository)(nil)

func (r *PgxRepository) CreateApplication(ctx context.Context, app *common.Application) error {
	query := `INSERT INTO applications (name, description, team_id) 
	          VALUES ($1, $2, $3) RETURNING id`
	return r.pool.QueryRow(ctx, query, app.Name, app.Description, app.TeamID).Scan(&app.ID)
}

func (r *PgxRepository) GetApplicationByID(ctx context.Context, id int) (*common.Application, error) {
	query := `SELECT id, name, description, team_id FROM applications WHERE id = $1`
	app := &common.Application{}
	err := r.pool.QueryRow(ctx, query, id).Scan(&app.ID, &app.Name, &app.Description, &app.TeamID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return app, nil
}

func (r *PgxRepository) GetApplicationByName(ctx context.Context, name string) (*common.Application, error) {
	query := `SELECT id, name, description, team_id FROM applications WHERE name = $1`
	app := &common.Application{}
	err := r.pool.QueryRow(ctx, query, name).Scan(&app.ID, &app.Name, &app.Description, &app.TeamID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return app, nil
}

func (r *PgxRepository) UpdateApplication(ctx context.Context, app *common.Application) error {
	query := `UPDATE applications SET 
		name = $1, 
		description = $2, 
		team_id = $3 
		WHERE id = $4`
	_, err := r.pool.Exec(ctx, query, app.Name, app.Description, app.TeamID, app.ID)
	return err
}

func (r *PgxRepository) DeleteApplication(ctx context.Context, id int) error {
	query := `DELETE FROM applications WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *PgxRepository) ListApplications(ctx context.Context) ([]*common.Application, error) {
	query := `SELECT id, name, description, team_id FROM applications`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	apps, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (*common.Application, error) {
		var app common.Application
		err := row.Scan(&app.ID, &app.Name, &app.Description, &app.TeamID)
		return &app, err
	})
	if err != nil {
		return nil, err
	}
	return apps, nil
}

func (r *PgxRepository) ListApplicationsByTeam(ctx context.Context, teamID int) ([]*common.Application, error) {
	query := `SELECT id, name, description, team_id FROM applications WHERE team_id = $1`
	rows, err := r.pool.Query(ctx, query, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	apps, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (*common.Application, error) {
		var app common.Application
		err := row.Scan(&app.ID, &app.Name, &app.Description, &app.TeamID)
		return &app, err
	})
	if err != nil {
		return nil, err
	}
	return apps, nil
}
