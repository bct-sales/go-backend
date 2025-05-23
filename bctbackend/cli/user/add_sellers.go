package user

import (
	"bctbackend/algorithms"
	"bctbackend/database"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"database/sql"
	"errors"
	"fmt"

	"golang.org/x/exp/rand"
	_ "modernc.org/sqlite"
)

type sellerCreationData struct {
	userId   models.Id
	password string
}

func collectExistingUserIds(db *sql.DB) (*algorithms.Set[models.Id], error) {
	result := algorithms.NewSet[models.Id]()

	queries.GetUsers(db, func(user *models.User) error {
		result.Add(user.UserId)
		return nil
	})

	return result, nil
}

func collectExistingPasswords(db *sql.DB) (*algorithms.Set[string], error) {
	result := algorithms.NewSet[string]()

	queries.GetUsers(db, func(user *models.User) error {
		result.Add(user.Password)
		return nil
	})

	return result, nil
}

func determineSellersToBeCreated(zones []int, sellersPerZone int, receiver func(models.Id) error) error {
	for _, zone := range zones {
		for i := 0; i < sellersPerZone; i++ {
			sellerId := models.Id(zone*100 + i)
			if err := receiver(sellerId); err != nil {
				return fmt.Errorf("failed to process seller %d: %w", sellerId, err)
			}
		}
	}

	return nil
}

func createPasswordList(seed uint64, usedPasswords algorithms.Set[string]) []string {
	rng := rand.New(rand.NewSource(seed))
	passwords := algorithms.Filter(Passwords, func(password string) bool { return !usedPasswords.Contains(password) })
	rng.Shuffle(len(passwords), func(i, j int) {
		passwords[i], passwords[j] = passwords[j], passwords[i]
	})
	return passwords
}

func AddSellers(databasePath string, seed uint64, zones []int, sellersPerZone int) (r_err error) {
	db, err := database.OpenDatabase(databasePath)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer func() { r_err = errors.Join(r_err, db.Close()) }()

	existingSellers, err := collectExistingUserIds(db)
	if err != nil {
		return fmt.Errorf("failed to collect existing sellers: %w", err)
	}

	usedPasswords, err := collectExistingPasswords(db)
	if err != nil {
		return fmt.Errorf("failed to collect existing passwords: %w", err)
	}

	sellersToBeCreated := []sellerCreationData{}
	passwords := createPasswordList(seed, *usedPasswords)
	passwordIndex := 0
	err = determineSellersToBeCreated(zones, sellersPerZone, func(sellerId models.Id) error {
		if !existingSellers.Contains(sellerId) {
			if passwordIndex == len(passwords) {
				return fmt.Errorf("ran out of unique passwords for %d", sellerId)
			}

			password := passwords[passwordIndex]
			passwordIndex++

			sellersToBeCreated = append(sellersToBeCreated, sellerCreationData{
				userId:   sellerId,
				password: password,
			})
		}

		return nil
	})
	if err != nil {
		return err
	}

	callback := func(add func(sellerId models.Id, roleId models.Id, createdAt models.Timestamp, lastActivity *models.Timestamp, password string)) {
		for _, seller := range sellersToBeCreated {
			add(
				seller.userId,
				models.SellerRoleId,
				models.Now(),
				nil,
				seller.password,
			)
		}
	}
	if err := queries.AddUsers(db, callback); err != nil {
		return fmt.Errorf("failed to add sellers: %w", err)
	}

	return nil
}
