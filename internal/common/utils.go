package common

import "errors"

func (p Permission) Validate() error {
	if p.OrganizationID != nil && p.TeamID != nil {
		return errors.New("permission cannot belong to both organization and team")
	}
	return nil
}
