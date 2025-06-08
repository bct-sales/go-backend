package user

import (
	"bctbackend/algorithms"
	"bctbackend/commands/common"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"database/sql"
	"fmt"

	"golang.org/x/exp/rand"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
	"github.com/urfave/cli/v2"
)

type addSellersCommand struct {
	common.Command
	seed           uint64
	zones          []int
	sellersPerZone int
}

func NewUserAddSellersCommand() *cobra.Command {
	var command *addSellersCommand

	command = &addSellersCommand{
		Command: common.Command{
			CobraCommand: &cobra.Command{
				Use:   "add-sellers",
				Short: "Add multiple sellers",
				Long: heredoc.Doc(`
					This command allows you to add many sellers in one go.
					It requires a list of zones and the number of sellers to add per zone.
					Note that the command will ensure that each specified zone
					reaches the required number of sellers.
					If a zone already has sellers, only the missing sellers will be added.

					For example, say zone 1 has 2 sellers and zone 2 has 3 sellers.
					If you run the command with --zones 1,2 --per-zone 5,
					then 3 sellers will be added to zone 1 and 2 sellers will be added to zone 2.
				`),
				RunE: func(cmd *cobra.Command, args []string) error {
					return command.execute()
				},
			},
		},
	}

	command.CobraCommand.Flags().Uint64Var(&command.seed, "seed", 0, "Seed for random password assignment")
	command.CobraCommand.Flags().IntSliceVar(&command.zones, "zones", nil, "Zones for which to add sellers")
	command.CobraCommand.Flags().IntVar(&command.sellersPerZone, "per-zone", 0, "Number of sellers to add per zone")
	command.CobraCommand.MarkFlagRequired("zones")
	command.CobraCommand.MarkFlagRequired("per-zone")

	return command.AsCobraCommand()
}

func (c *addSellersCommand) execute() error {
	return c.WithOpenedDatabase(func(db *sql.DB) error {
		existingSellers, err := collectExistingUserIds(db)
		if err != nil {
			return fmt.Errorf("failed to collect existing sellers: %w", err)
		}

		usedPasswords, err := collectExistingPasswords(db)
		if err != nil {
			return fmt.Errorf("failed to collect existing passwords: %w", err)
		}

		sellersToBeCreated := []sellerCreationData{}
		passwords := createPasswordList(c.seed, *usedPasswords)
		passwordIndex := 0
		err = determineSellersToBeCreated(c.zones, c.sellersPerZone, func(sellerId models.Id) error {
			if !existingSellers.Contains(sellerId) {
				if passwordIndex == len(passwords) {
					return cli.Exit("ran out of unique passwords", 1)
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
	})
}

type sellerCreationData struct {
	userId   models.Id
	password string
}

func collectExistingUserIds(db *sql.DB) (*algorithms.Set[models.Id], error) {
	result := algorithms.NewSet[models.Id]()

	err := queries.GetUsers(db, func(user *models.User) error {
		result.Add(user.UserId)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to collect existing user ids: %w", err)
	}

	return &result, nil
}

func collectExistingPasswords(db *sql.DB) (*algorithms.Set[string], error) {
	result := algorithms.NewSet[string]()

	err := queries.GetUsers(db, func(user *models.User) error {
		result.Add(user.Password)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to collect existing passwords: %w", err)
	}

	return &result, nil
}

func determineSellersToBeCreated(zones []int, sellersPerZone int, receiver func(models.Id) error) error {
	for _, zone := range zones {
		for i := range sellersPerZone {
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
