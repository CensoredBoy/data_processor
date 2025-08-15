package common

import "errors"

func (p Permission) Validate() error {
	if p.OrganizationID != nil && p.TeamID != nil {
		return errors.New("permission cannot belong to both organization and team")
	}
	return nil
}

func Int32PtrFromIntPtr(i *int) *int32 {
	if i == nil {
		return nil
	}
	i32 := int32(*i)
	return &i32
}
func Int32FromIntPtr(i *int) int32 {
	if i == nil {
		return 0
	}
	i32 := int32(*i)
	return i32
}
func Int32PtrToIntPtr(i32 *int32) *int {
	if i32 == nil {
		return nil
	}
	i := int(*i32)
	return &i
}
