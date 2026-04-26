package meta

var Role = RoleMetadata{
	Table:  "roles",
	RoleID: "role_id",
	Name:   "name",
}

type RoleMetadata struct {
	Table  string
	RoleID string
	Name   string
}
