package queries

import (
	"context"
	dberr "bctbackend/database/errors"
	models "bctbackend/database/models"
	"bctbackend/database/queries/internal/isql"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

func timestampToNullInt64(timestamp *models.Timestamp) sql.NullInt64 {
	if timestamp == nil {
		return sql.NullInt64{}
	}

	return sql.NullInt64{Int64: int64(*timestamp), Valid: true}
}

func nullInt64ToTimestamp(value sql.NullInt64) *models.Timestamp {
	if !value.Valid {
		return nil
	}

	timestamp := models.Timestamp(value.Int64)
	return &timestamp
}

func sellerItemCountFlagToNullBool(flag GetSellerItemCountFlag) sql.NullBool {
	switch flag {
	case IncludeAll:
		return sql.NullBool{}
	case Exclude:
		return sql.NullBool{Bool: false, Valid: true}
	case IncludeOnly:
		return sql.NullBool{Bool: true, Valid: true}
	default:
		return sql.NullBool{}
	}
}

// AddUserWithID adds a user to the database with a specific user ID.
// An ErrUserIDAlreadyInUse is returned if the user ID is already in use.
// An ErrNoSuchRole is returned if the role ID is invalid.
func AddUserWithID(
	ctx context.Context,
	database DatabaseQuerier,
	userID models.ID,
	roleID models.RoleID,
	createdAt models.Timestamp,
	lastActivity *models.Timestamp,
	password string) (r_err error) {

	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	if !roleID.IsValid() {
		return fmt.Errorf("invalid role id %d: %w", roleID.ID, dberr.ErrNoSuchRole)
	}

	err := isql.New(database).AddUserWithID(ctx, isql.AddUserWithIDParams{
		UserID:       userID,
		RoleID:       roleID.Int64(),
		CreatedAt:    int64(createdAt),
		LastActivity: timestampToNullInt64(lastActivity),
		Password:     password,
	})

	if err != nil {
		userExists, err := UserWithIDExists(ctx, database, userID)
		if err != nil {
			return err
		}
		if userExists {
			return fmt.Errorf("trying to add user with id %d: %w", userID, dberr.ErrIDAlreadyInUse)
		}

		return fmt.Errorf("failed to add user with id %d: %w", userID, err)
	}

	return nil
}

func AddUser(
	ctx context.Context,
	db DatabaseQuerier,
	roleID models.RoleID,
	createdAt models.Timestamp,
	lastActivity *models.Timestamp,
	password string) (r_result models.ID, r_err error) {

	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	if !roleID.IsValid() {
		return 0, fmt.Errorf("invalid role id %d: %w", roleID.ID, dberr.ErrNoSuchRole)
	}

	userID, err := isql.New(db).AddUser(ctx, isql.AddUserParams{
		RoleID:       roleID.Int64(),
		CreatedAt:    int64(createdAt),
		LastActivity: timestampToNullInt64(lastActivity),
		Password:     password,
	})
	if err != nil {
		return 0, err
	}

	return userID, nil
}

type AddUsersCallback func(addUser func(userID models.ID, roleID models.RoleID, createdAt models.Timestamp, lastActivity *models.Timestamp, password string))

// NOTE: AddUsers is kept as raw SQL: sqlc's SQLite engine cannot capture a
// query parameter passed to the json_each table-valued function, which the
// batch insert would require. See docs/superpowers/specs/2026-11-01-sqlc-query-migration-design.md.
func AddUsers(ctx context.Context, database DatabaseQuerier, callback AddUsersCallback) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	valuesString := []string{}
	arguments := []any{}
	tupleString := "(?, ?, ?, ?, ?)"

	add := func(userID models.ID, roleID models.RoleID, createdAt models.Timestamp, lastActivity *models.Timestamp, password string) {
		valuesString = append(valuesString, tupleString)
		arguments = append(arguments, userID, roleID.Int64(), createdAt, lastActivity, password)
	}

	callback(add)

	if len(valuesString) == 0 {
		return nil
	}

	query := `INSERT INTO users (user_id, role_id, created_at, last_activity, password) VALUES ` + strings.Join(valuesString, ",")

	if _, err := database.ExecContext(ctx, query, arguments...); err != nil {
		return err
	}

	return nil
}

func UserWithIDExists(ctx context.Context, db DatabaseQuerier, userID models.ID) (r_result bool, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	_, err := isql.New(db).UserWithIDExists(ctx, userID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

// GetUserWithID retrieves a user from the database by their user ID.
// An ErrNoSuchUser is returned if the user does not exist.
func GetUserWithID(ctx context.Context, db DatabaseQuerier, userID models.ID) (r_result *models.User, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	user, err := isql.New(db).GetUser(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("failed to get user with id %d: %w", userID, dberr.ErrNoSuchUser)
		}

		return nil, err
	}

	result := models.User{
		UserID:       user.UserID,
		RoleID:       models.RoleID{ID: models.ID(user.RoleID)},
		CreatedAt:    models.Timestamp(user.CreatedAt),
		LastActivity: nullInt64ToTimestamp(user.LastActivity),
		Password:     user.Password,
	}
	return &result, nil
}

// GetUsers retrieves all users from the database.
func GetUsers(ctx context.Context, db DatabaseQuerier, receiver func(*models.User) error) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	users, err := isql.New(db).GetUsers(ctx)
	if err != nil {
		return err
	}

	for i := range users {
		user := models.User{
			UserID:       users[i].UserID,
			RoleID:       models.RoleID{ID: models.ID(users[i].RoleID)},
			CreatedAt:    models.Timestamp(users[i].CreatedAt),
			LastActivity: nullInt64ToTimestamp(users[i].LastActivity),
			Password:     users[i].Password,
		}
		if err := receiver(&user); err != nil {
			return err
		}
	}

	return nil
}

type UserWithItemCount struct {
	models.User
	ItemCount int
}

func GetUsersWithItemCount(ctx context.Context, db DatabaseQuerier, itemSelection ItemSelection, receiver func(*UserWithItemCount) error) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	rows, err := isql.New(db).GetUsersWithItemCount(ctx, itemSelectionToNullBool(itemSelection))
	if err != nil {
		return err
	}

	for i := range rows {
		userWithItemCount := UserWithItemCount{
			User: models.User{
				UserID:       rows[i].UserID,
				RoleID:       models.RoleID{ID: models.ID(rows[i].RoleID)},
				CreatedAt:    models.Timestamp(rows[i].CreatedAt),
				LastActivity: nullInt64ToTimestamp(rows[i].LastActivity),
				Password:     rows[i].Password,
			},
			ItemCount: int(rows[i].ItemCount),
		}
		if err := receiver(&userWithItemCount); err != nil {
			return err
		}
	}

	return nil
}

// UpdateUserPassword updates the password of a user in the database by their user ID.
// An ErrNoSuchUser is returned if the user does not exist.
func UpdateUserPassword(ctx context.Context, database DatabaseQuerier, userID models.ID, password string) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	userExists, err := UserWithIDExists(ctx, database, userID)
	if err != nil {
		return err
	}
	if !userExists {
		return fmt.Errorf("failed to update password of user %d: %w", userID, dberr.ErrNoSuchUser)
	}

	err = isql.New(database).UpdateUserPassword(ctx, isql.UpdateUserPasswordParams{
		Password: password,
		UserID:   userID,
	})

	return err
}

// EnsureUserExists checks if a user exists in the database by their user ID.
// An ErrNoSuchUser is returned if the user does not exist.
func EnsureUserExists(ctx context.Context, db DatabaseQuerier, userID models.ID) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	userExists, err := UserWithIDExists(ctx, db, userID)
	if err != nil {
		return fmt.Errorf("failed to ensure user %d exists: %w", userID, err)
	}
	if !userExists {
		return fmt.Errorf("failed to ensure user %d exists: %w", userID, dberr.ErrNoSuchUser)
	}
	return nil
}

// EnsureUserExistsAndHasRole checks if a user has a specific role.
// An ErrNoSuchUser is returned if the user does not exist.
// An ErrWrongRole is returned if the user has a different role.
func EnsureUserExistsAndHasRole(ctx context.Context, db DatabaseQuerier, userID models.ID, expectedRoleID models.RoleID) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	user, err := GetUserWithID(ctx, db, userID)

	if err != nil {
		return err
	}

	if user.RoleID != expectedRoleID {
		return fmt.Errorf("user %d expected to have role %s but is %s instead: %w", userID, expectedRoleID.Name(), user.RoleID.Name(), dberr.ErrWrongRole)
	}

	return nil
}

// RemoveUserWithID removes a user from the database by their user ID.
// An ErrNoSuchUser is returned if the user does not exist.
// An error is returned if the user cannot be removed, e.g., because items or sales are
// associated with the user.
func RemoveUserWithID(ctx context.Context, db DatabaseQuerier, userID models.ID) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	{
		userExist, err := UserWithIDExists(ctx, db, userID)
		if err != nil {
			return err
		}
		if !userExist {
			return fmt.Errorf("failed to remove user with id %d: %w", userID, dberr.ErrNoSuchUser)
		}
	}

	err := isql.New(db).RemoveUserWithID(ctx, userID)

	return err
}

// UpdateLastActivity updates the last activity timestamp of a user in the database by their user ID.
// An ErrNoSuchUser is returned if the user does not exist.
func UpdateLastActivity(ctx context.Context, db DatabaseQuerier, userID models.ID, lastActivity models.Timestamp) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	{
		userExist, err := UserWithIDExists(ctx, db, userID)
		if err != nil {
			return err
		}
		if !userExist {
			return fmt.Errorf("failed to update last activity of user %d: %w", userID, dberr.ErrNoSuchUser)
		}
	}

	err := isql.New(db).UpdateLastActivity(ctx, isql.UpdateLastActivityParams{
		LastActivity: sql.NullInt64{Int64: int64(lastActivity), Valid: true},
		UserID:       userID,
	})

	return err
}

type GetSellerItemCountFlag int

const (
	IncludeAll GetSellerItemCountFlag = iota
	Exclude
	IncludeOnly
)

// CountSellerItems returns the number of sold items owned by a given seller.
// An ErrNoSuchUser is returned if the user does not exist.
// An ErrWrongRole is returned if the user with the given id is not a seller.
// The frozen and hidden parameters allow to specify whether to include or exclude frozen and hidden items.
// For example, if frozen equals IncludeAll, both frozen and unfrozen items are included.
// If frozen equals Exclude, only unfrozen items are included.
// If frozen equals IncludeOnly, only frozen items are included.
func CountSellerItems(ctx context.Context, db DatabaseQuerier, sellerID models.ID, frozen GetSellerItemCountFlag, hidden GetSellerItemCountFlag) (r_result int, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	if err := EnsureUserExistsAndHasRole(ctx, db, sellerID, models.NewSellerRoleID()); err != nil {
		return 0, fmt.Errorf("failed to get hidden item count of user %d: %w", sellerID, err)
	}

	itemCount, err := isql.New(db).CountSellerItems(ctx, isql.CountSellerItemsParams{
		SellerID: sellerID.Int64(),
		Frozen:   sellerItemCountFlagToNullBool(frozen),
		Hidden:   sellerItemCountFlagToNullBool(hidden),
	})
	if err != nil {
		return 0, fmt.Errorf("failed to get seller's %d item count: %w", sellerID, err)
	}

	return int(itemCount), nil
}

// GetSellerTotalValueOfAllItems returns the total value of all items owned by a given seller.
// Note that whether an item has been sold does not matter.
// An ErrNoSuchUser is returned if the user does not exist.
// An ErrWrongRole is returned if the user with the given id is not a seller.
func GetSellerTotalValueOfAllItems(ctx context.Context, db DatabaseQuerier, sellerID models.ID, itemSelection ItemSelection) (r_result models.MoneyInCents, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	// Ensure the user exists and is a seller
	if err := EnsureUserExistsAndHasRole(ctx, db, sellerID, models.NewSellerRoleID()); err != nil {
		return 0, fmt.Errorf("failed to get total price of all items of user %d: %w", sellerID, err)
	}

	totalValue, err := isql.New(db).GetSellerTotalValueOfAllItems(ctx, isql.GetSellerTotalValueOfAllItemsParams{
		SellerID: sellerID.Int64(),
		Hidden:   itemSelectionToNullBool(itemSelection),
	})
	if err != nil {
		return 0, err
	}

	return models.MoneyInCents(totalValue), nil
}

func CollectPasswords(ctx context.Context, db DatabaseQuerier) (r_result []string, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	passwords, err := isql.New(db).CollectPasswords(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch passwords from database: %w", err)
	}

	return passwords, nil
}
