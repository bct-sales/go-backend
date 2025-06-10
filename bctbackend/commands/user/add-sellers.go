package user

import (
	"bctbackend/algorithms"
	"bctbackend/commands/common"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"golang.org/x/exp/rand"

	"github.com/MakeNowJust/heredoc"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

type addSellersCommand struct {
	common.Command
	seed           uint64
	zonesString    string
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
	command.CobraCommand.Flags().StringVar(&command.zonesString, "zones", "", "Zones for which to add sellers")
	command.CobraCommand.Flags().IntVar(&command.sellersPerZone, "per-zone", 0, "Number of sellers to add per zone")
	command.CobraCommand.MarkFlagRequired("zones")
	command.CobraCommand.MarkFlagRequired("per-zone")

	return command.AsCobraCommand()
}

func (c *addSellersCommand) execute() error {
	return c.WithOpenedDatabase(func(db *sql.DB) error {
		zones, err := parseZones(c.zonesString)
		if err != nil {
			c.PrintErrorf("failed to parse zones")
			return err
		}

		sellersToBeCreated, err := c.determineSellersToBeCreated(db, zones, c.sellersPerZone)
		if err != nil {
			return nil
		}

		callback := func(add func(sellerId models.Id, roleId models.RoleId, createdAt models.Timestamp, lastActivity *models.Timestamp, password string)) {
			for _, seller := range sellersToBeCreated {
				add(
					seller.userId,
					models.NewSellerRoleId(),
					models.Now(),
					nil,
					seller.password,
				)
			}
		}
		if err := queries.AddUsers(db, callback); err != nil {
			c.PrintErrorf("failed to add sellers")
			return fmt.Errorf("failed to add sellers: %w", err)
		}

		c.showAddedSellers(sellersToBeCreated)

		return nil
	})
}

type sellerCreationData struct {
	userId   models.Id
	password string
}

func (c *addSellersCommand) collectExistingUserIds(db *sql.DB) (*algorithms.Set[models.Id], error) {
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

func (c *addSellersCommand) collectExistingPasswords(db *sql.DB) (*algorithms.Set[string], error) {
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

func (c *addSellersCommand) determineSellersToBeCreated(db *sql.DB, zones []int, sellersPerZone int) ([]*sellerCreationData, error) {
	existingSellers, err := c.collectExistingUserIds(db)
	if err != nil {
		return nil, err
	}

	usedPasswords, err := c.collectExistingPasswords(db)
	if err != nil {
		return nil, err
	}

	passwords := c.createUniquePasswordList(c.seed, *usedPasswords)
	passwordIndex := 0
	sellersToBeCreated := []*sellerCreationData{}
	for _, zone := range zones {
		for i := range sellersPerZone {
			sellerId := models.Id(zone*100 + i)

			if !existingSellers.Contains(sellerId) {
				if passwordIndex == len(passwords) {
					c.PrintErrorf("ran out of unique passwords")
					return nil, fmt.Errorf("ran out of unique passwords")
				}

				password := passwords[passwordIndex]
				passwordIndex++

				sellersToBeCreated = append(sellersToBeCreated, &sellerCreationData{
					userId:   sellerId,
					password: password,
				})
			}
		}
	}

	return sellersToBeCreated, nil
}

func (c *addSellersCommand) createUniquePasswordList(seed uint64, usedPasswords algorithms.Set[string]) []string {
	rng := rand.New(rand.NewSource(seed))
	passwords := algorithms.Filter(Passwords, func(password string) bool { return !usedPasswords.Contains(password) })
	rng.Shuffle(len(passwords), func(i, j int) {
		passwords[i], passwords[j] = passwords[j], passwords[i]
	})
	return passwords
}

var ErrInvalidZoneFormat = errors.New("invalid zone format")

func parseZones(str string) ([]int, error) {
	zoneStrings := strings.Split(str, ",")
	result := []int{}

	for _, zoneString := range zoneStrings {
		parts := strings.Split(zoneString, ",")

		for _, part := range parts {
			endpoints := strings.Split(part, "-")
			if len(endpoints) == 1 {
				zone, err := strconv.Atoi(strings.TrimSpace(endpoints[0]))
				if err != nil {
					return nil, fmt.Errorf("invalid zone format %s: %w, %w", part, err, ErrInvalidZoneFormat)
				}
				result = append(result, zone)
			} else if len(endpoints) == 2 {
				start, err := strconv.Atoi(strings.TrimSpace(endpoints[0]))
				if err != nil {
					return nil, fmt.Errorf("invalid zone format %s: %w, %w", part, err, ErrInvalidZoneFormat)
				}
				end, err := strconv.Atoi(strings.TrimSpace(endpoints[1]))
				if err != nil {
					return nil, fmt.Errorf("invalid zone format %s: %w, %w", part, err, ErrInvalidZoneFormat)
				}
				if start >= end {
					return nil, fmt.Errorf("invalid zone format %s: %w, %w", part, err, ErrInvalidZoneFormat)
				}
				for i := start; i <= end; i++ {
					result = append(result, i)
				}
			} else {
				return nil, fmt.Errorf("invalid zone format %s: %w", part, ErrInvalidZoneFormat)
			}
		}
	}

	slices.Sort(result)
	result = slices.Compact(result)

	return result, nil
}

func (c *addSellersCommand) showAddedSellers(sellers []*sellerCreationData) error {
	if len(sellers) == 0 {
		c.Printf("No sellers added, all specified zones already have enough sellers.\n")
		return nil
	}

	tableData := pterm.TableData{
		{"Seller ID", "Password"},
	}

	for _, seller := range sellers {
		tableData = append(tableData, []string{
			fmt.Sprintf("%d", seller.userId),
			seller.password,
		})
	}

	c.Printf("Added the following sellers:\n\n")
	if err := pterm.DefaultTable.WithData(tableData).Render(); err != nil {
		c.PrintErrorf("failed to render table: %v", err)
		return err
	}

	return nil
}
