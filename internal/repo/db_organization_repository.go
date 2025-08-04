package repo

import (
	"context"
	"data_processor/internal/common"
	"errors"
	"github.com/jackc/pgx/v5"
)

var _ IOrganizationRepository = (*PgxRepository)(nil)

func (r *PgxRepository) CreateOrganization(ctx context.Context, org *common.Organization) error {
	query := `INSERT INTO organizations (project_name, owner_id) VALUES ($1, $2) RETURNING id`
	return r.pool.QueryRow(ctx, query, org.ProjectName, org.OwnerID).Scan(&org.ID)
}

func (r *PgxRepository) GetOrganizationByID(ctx context.Context, id int) (*common.Organization, error) {
	query := `SELECT id, project_name, owner_id FROM organizations WHERE id = $1`
	org := &common.Organization{}
	err := r.pool.QueryRow(ctx, query, id).Scan(&org.ID, &org.ProjectName, &org.OwnerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return org, nil
}

func (r *PgxRepository) GetOrganizationByName(ctx context.Context, name string) (*common.Organization, error) {
	query := `SELECT id, project_name, owner_id FROM organizations WHERE project_name = $1`
	org := &common.Organization{}
	err := r.pool.QueryRow(ctx, query, name).Scan(&org.ID, &org.ProjectName, &org.OwnerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return org, nil
}

func (r *PgxRepository) UpdateOrganization(ctx context.Context, org *common.Organization) error {
	query := `UPDATE organizations SET project_name = $1, owner_id = $2 WHERE id = $3`
	_, err := r.pool.Exec(ctx, query, org.ProjectName, org.OwnerID, org.ID)
	return err
}

func (r *PgxRepository) DeleteOrganization(ctx context.Context, id int) error {
	query := `DELETE FROM organizations WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *PgxRepository) ListOrganizations(ctx context.Context) ([]*common.Organization, error) {
	query := `SELECT id, project_name, owner_id FROM organizations`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orgs, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (*common.Organization, error) {
		var org common.Organization
		err := row.Scan(&org.ID, &org.ProjectName, &org.OwnerID)
		return &org, err
	})
	if err != nil {
		return nil, err
	}
	return orgs, nil
}

func (r *PgxRepository) ListOrganizationsByOwner(ctx context.Context, ownerID common.UserID) ([]*common.Organization, error) {
	query := `SELECT id, project_name, owner_id FROM organizations WHERE owner_id = $1`
	rows, err := r.pool.Query(ctx, query, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orgs, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (*common.Organization, error) {
		var org common.Organization
		err := row.Scan(&org.ID, &org.ProjectName, &org.OwnerID)
		return &org, err
	})
	if err != nil {
		return nil, err
	}
	return orgs, nil
}
