package pages

import (
	"database/sql"
)

type PageBase struct {
	Database   *sql.DB
	ScreenSize *Size
}
