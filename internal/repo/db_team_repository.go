package repo

import (
	"context"
	"data_processor/internal/common"
	"errors"
	"github.com/jackc/pgx/v5"
)

var _ ITeamRepository = (*PgxRepository)(nil)

func (r *PgxRepository) CreateTeam(ctx context.Context, team *common.Team) error {
	query := `INSERT INTO teams (team_name, owner_id, folder, organization_id) 
	          VALUES ($1, $2, $3, $4) RETURNING id`
	return r.pool.QueryRow(ctx, query, team.TeamName, team.OwnerID, team.Folder, team.OrganizationID).Scan(&team.ID)
}

func (r *PgxRepository) GetTeamByID(ctx context.Context, id int) (*common.Team, error) {
	query := `SELECT id, team_name, owner_id, folder, organization_id FROM teams WHERE id = $1`
	team := &common.Team{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&team.ID, &team.TeamName, &team.OwnerID, &team.Folder, &team.OrganizationID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return team, nil
}

func (r *PgxRepository) GetTeamByName(ctx context.Context, name string) (*common.Team, error) {
	query := `SELECT id, team_name, owner_id, folder, organization_id FROM teams WHERE team_name = $1`
	team := &common.Team{}
	err := r.pool.QueryRow(ctx, query, name).Scan(
		&team.ID, &team.TeamName, &team.OwnerID, &team.Folder, &team.OrganizationID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return team, nil
}

func (r *PgxRepository) UpdateTeam(ctx context.Context, team *common.Team) error {
	query := `UPDATE teams SET 
		team_name = $1, 
		owner_id = $2, 
		folder = $3, 
		organization_id = $4 
		WHERE id = $5`
	_, err := r.pool.Exec(ctx, query,
		team.TeamName, team.OwnerID, team.Folder, team.OrganizationID, team.ID)
	return err
}

func (r *PgxRepository) DeleteTeam(ctx context.Context, id int) error {
	query := `DELETE FROM teams WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *PgxRepository) ListTeams(ctx context.Context) ([]*common.Team, error) {
	query := `SELECT id, team_name, owner_id, folder, organization_id FROM teams`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	teams, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (*common.Team, error) {
		var team common.Team
		err := row.Scan(&team.ID, &team.TeamName, &team.OwnerID, &team.Folder, &team.OrganizationID)
		return &team, err
	})
	if err != nil {
		return nil, err
	}
	return teams, nil
}

func (r *PgxRepository) ListTeamsByOrganization(ctx context.Context, orgID int) ([]*common.Team, error) {
	query := `SELECT id, team_name, owner_id, folder, organization_id 
	          FROM teams WHERE organization_id = $1`
	rows, err := r.pool.Query(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	teams, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (*common.Team, error) {
		var team common.Team
		err := row.Scan(&team.ID, &team.TeamName, &team.OwnerID, &team.Folder, &team.OrganizationID)
		return &team, err
	})
	if err != nil {
		return nil, err
	}
	return teams, nil
}

func (r *PgxRepository) ListTeamsByOwner(ctx context.Context, ownerID int) ([]*common.Team, error) {
	query := `SELECT id, team_name, owner_id, folder, organization_id 
	          FROM teams WHERE owner_id = $1`
	rows, err := r.pool.Query(ctx, query, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	teams, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (*common.Team, error) {
		var team common.Team
		err := row.Scan(&team.ID, &team.TeamName, &team.OwnerID, &team.Folder, &team.OrganizationID)
		return &team, err
	})
	if err != nil {
		return nil, err
	}
	return teams, nil
}
