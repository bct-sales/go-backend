package database

import (
	"bctbackend/algorithms"
	"bctbackend/commands/common"
	"bctbackend/database"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"context"
	"database/sql"
	"fmt"
	"math/rand/v2"
	"slices"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

const (
	zoneCount = 12
)

var colors = [...]string{
	"red",
	"green",
	"blue",
	"yellow",
	"orange",
	"purple",
	"black",
	"white",
}

var clothing = [...]string{
	"T-shirt",
	"hoodie",
	"jacket",
	"sweater",
	"jeans",
	"pants",
	"shorts",
	"skirt",
	"dress",
	"hat",
	"scarf",
	"socks",
	"gloves",
}

var clothingCategories = [...]models.ID{
	common.CategoryId_Clothing50_56,
	common.CategoryId_Clothing56_62,
	common.CategoryId_Clothing68_80,
	common.CategoryId_Clothing86_92,
	common.CategoryId_Clothing92_98,
	common.CategoryId_Clothing104_116,
	common.CategoryId_Clothing122_128,
	common.CategoryId_Clothing128_140,
	common.CategoryId_Clothing140_152,
}

var books = [...]string{
	"War and Peace",
	"Price and Prejudice",
	"Crime and Punishment",
	"Little Women",
	"It",
	"House of Leaves",
	"Ulysses",
	"Finnegan's Wake",
	"Brave New World",
	"Animal Farm",
	"1984",
	"Space, Time and Nathaniel",
	"The Road",
	"No Country for Old Men",
	"The Catcher in the Rye",
	"A Column of Fire",
	"Pillars of the Earth",
	"Frankenstein",
	"Dracula",
}

var toys = [...]string{
	"LEGO set",
	"Barbie doll",
	"Action figure",
	"Puzzle",
	"Board game",
	"Stuffed animal",
	"Nintendo Switch",
	"Nintendo Switch 2",
	"Nintendo DS",
	"Nintendo 3DS",
	"Nintendo Wii",
	"Nintendo Wii U",
	"PlayStation 1",
	"PlayStation 2",
	"PlayStation 3",
	"PlayStation 4",
	"PlayStation 5",
	"Xbox Series X",
	"Zelda: Breath of the Wild",
	"Zelda: Tears of the Kingdom",
	"Zelda: Ocarina of Time",
	"Zelda: Majora's Mask",
	"Zelda: Wind Waker",
	"Zelda: Twilight Princess",
	"Zelda: Skyward Sword",
	"Zelda: A Link to the Past",
	"Zelda: A Link Between Worlds",
	"Zelda: Oracle of Seasons",
	"Zelda: Oracle of Ages",
	"Zelda: Four Swords Adventures",
	"Zelda: Spirit Tracks",
	"Zelda: Phantom Hourglass",
	"Zelda: Link's Awakening",
	"Zelda: A Link Between Worlds",
	"Zelda: Tri Force Heroes",
	"Zelda: Hyrule Warriors",
	"Super Mario Odyssey",
	"Super Mario Galaxy",
	"Super Mario 3D World",
	"Super Mario 64",
	"Super Mario Sunshine",
	"Super Mario Maker",
	"Super Mario Party",
	"Super Paper Mario",
	"Smash Bros. Ultimate",
	"Call of Duty",
	"Disco Elysium",
	"The Witcher 3",
	"The Last of Us",
	"Uncharted 4",
	"God of War",
	"Chess set",
	"Checkers set",
	"Poker set",
}

type dummyDatabaseCommand struct {
	common.Command
	seed  uint64     `exhaustruct:"optional"`
	force bool       `exhaustruct:"optional"`
	rng   *rand.Rand `exhaustruct:"optional"`
}

func NewDatabaseDummyCommand() *cobra.Command {
	var command *dummyDatabaseCommand

	command = &dummyDatabaseCommand{
		Command: common.Command{
			CobraCommand: &cobra.Command{
				Use:   "dummy",
				Short: "Add dummy data to the database",
				Long: heredoc.Doc(`
					This command adds dummy data to the database for testing purposes.
					WARNING: This will reset the database and remove all existing data.
				`),
				RunE: func(cmd *cobra.Command, args []string) error {
					return command.execute()
				},
				Args: cobra.NoArgs,
			},
		},
	}

	command.CobraCommand.Flags().Uint64Var(&command.seed, "seed", 0, "Seed for random number generation")
	command.CobraCommand.Flags().BoolVar(&command.force, "overwrite", false, "Necessary to indicate that you're aware all data will be overwritten")

	return command.AsCobraCommand()
}

func (c *dummyDatabaseCommand) execute() error {
	c.rng = rand.New(rand.NewPCG(0, c.seed))

	if !c.force {
		c.PrintErrorf("Missing --overwrite; database left untouched\n")
		return nil
	}

	return c.WithOpenedDatabase(func(db *sql.DB) error {
		c.Printf("Resetting database\n")
		if err := database.ResetDatabase(db); err != nil {
			return fmt.Errorf("failed to reset database: %w", err)
		}

		if err := c.addCategories(db); err != nil {
			return err
		}

		_, err := c.addAdmin(db)
		if err != nil {
			return err
		}

		cashierIDs, err := c.addCashiers(db)
		if err != nil {
			return err
		}

		sellerIDs, err := c.addSellers(db)
		if err != nil {
			return err
		}

		itemIDs, err := c.addItems(db, sellerIDs)
		if err != nil {
			return err
		}

		if err := c.addSales(db, cashierIDs, itemIDs); err != nil {
			return err
		}

		return nil
	})
}

func (c *dummyDatabaseCommand) addCategories(db *sql.DB) error {
	c.Printf("Adding categories\n")

	addCategory := func(id models.ID, name string) error {
		return queries.AddCategoryWithID(db, id, name)
	}

	if err := common.GenerateDefaultCategories(addCategory); err != nil {
		return fmt.Errorf("failed to add categories: %w", err)
	}
	return nil
}

func (c *dummyDatabaseCommand) addAdmin(db *sql.DB) (models.ID, error) {
	c.Printf("Adding admin user\n")

	id := models.ID(1)
	roleId := models.NewAdminRoleID()
	createdAt := models.Now()
	var lastActivity *models.Timestamp = nil
	password := "abc"

	if err := queries.AddUserWithId(db, id, roleId, createdAt, lastActivity, password); err != nil {
		return 0, fmt.Errorf("failed to add admin: %w", err)
	}

	return id, nil
}

func (c *dummyDatabaseCommand) addCashiers(db *sql.DB) ([]models.ID, error) {
	c.Printf("Adding cashier users\n")

	cashierCount := c.rng.IntN(10) + 1
	cashierIDs := make([]models.ID, 0, cashierCount)

	for range cashierCount {
		roleId := models.NewCashierRoleID()
		createdAt := models.Now()
		var lastActivity *models.Timestamp = nil
		password := "abc"

		cashierID, err := queries.AddUser(db, roleId, createdAt, lastActivity, password)

		if err != nil {
			return nil, fmt.Errorf("failed to add cashier: %w", err)
		}

		cashierIDs = append(cashierIDs, cashierID)
	}

	return cashierIDs, nil
}

func (c *dummyDatabaseCommand) addSellers(db *sql.DB) ([]models.ID, error) {
	c.Printf("Adding sellers\n")

	sellerIds := make([]models.ID, 0, zoneCount*10)

	addSellers := func(addUser func(userId models.ID, roleId models.RoleId, createdAt models.Timestamp, lastActivity *models.Timestamp, password string)) {
		for area := 1; area <= zoneCount; area++ {
			sellerCount := c.rng.IntN(10) + 1

			for offset := 0; offset != sellerCount; offset++ {
				userId := c.getSellerId(area, offset)
				roleId := models.NewSellerRoleId()
				createdAt := models.Now()
				var lastActivity *models.Timestamp = nil
				password := fmt.Sprintf("%d", userId)

				addUser(userId, roleId, createdAt, lastActivity, password)

				sellerIds = append(sellerIds, userId)
			}
		}
	}
	if err := queries.AddUsers(db, addSellers); err != nil {
		return nil, fmt.Errorf("failed to add sellers: %w", err)
	}

	return sellerIds, nil
}

func (c *dummyDatabaseCommand) getSellerId(zone int, offset int) models.ID {
	return models.ID(zone*100 + offset)
}

func (c *dummyDatabaseCommand) addItems(db *sql.DB, sellerIds []models.ID) ([]models.ID, error) {
	c.Printf("Adding items\n")

	err := queries.AddItems(db, func(addItem queries.AddItemFunction) {
		for _, sellerId := range sellerIds {
			itemCount := c.rng.IntN(50) + 5

			times := c.generateChronologicalTimes(itemCount, 60*60*24, 60*60*24*30)

			for _, addedAt := range times {
				description, category := c.generateRandomItemDescriptionAndCategory()
				priceInCents := models.MoneyInCents(c.rng.IntN(100)*50 + 50)
				donation := c.rng.IntN(20) == 0
				charity := c.rng.IntN(20) == 0
				frozen := c.rng.IntN(20) == 0
				hidden := false

				addItem(
					addedAt,
					description,
					priceInCents,
					category,
					sellerId,
					donation,
					charity,
					frozen,
					hidden,
				)
			}
		}
	})

	if err != nil {
		return nil, fmt.Errorf("failed to add items: %w", err)
	}

	itemIds, err := queries.GetItemIds(db)
	if err != nil {
		return nil, fmt.Errorf("failed to get item IDs: %w", err)
	}

	return itemIds, nil
}

func (c *dummyDatabaseCommand) generateRandomTime(minDelta int64, maxDelta int64) models.Timestamp {
	now := models.Now().Int64()
	delta := c.rng.Int64N(maxDelta-minDelta) + minDelta
	return models.Timestamp(now - delta)
}

func (c *dummyDatabaseCommand) generateChronologicalTimes(count int, minDelta int64, maxDelta int64) []models.Timestamp {
	times := algorithms.RepeatCollect(count, func() models.Timestamp { return c.generateRandomTime(minDelta, maxDelta) })
	slices.Sort(times)
	return times
}

func (c *dummyDatabaseCommand) generateRandomItemDescriptionAndCategory() (string, models.ID) {
	switch c.rng.IntN(3) {
	case 0:
		return c.generateRandomClothing()
	case 1:
		return c.generateRandomBooks()
	default:
		return c.generateRandomToys()
	}
}

func (c *dummyDatabaseCommand) generateRandomColor() string {
	return colors[c.rng.IntN(len(colors))]
}

func (c *dummyDatabaseCommand) generateRandomClothingCategoryId() models.ID {
	return pickRandom(c.rng, clothingCategories[:])
}

func (c *dummyDatabaseCommand) generateRandomClothing() (string, models.ID) {
	color := c.generateRandomColor()
	categoryId := c.generateRandomClothingCategoryId()
	clothingType := pickRandom(c.rng, clothing[:])
	description := fmt.Sprintf("%s %s", color, clothingType)

	return description, categoryId
}

func pickRandom[T any](rng *rand.Rand, items []T) T {
	if len(items) == 0 {
		panic("cannot pick from empty slice")
	}
	return items[rng.IntN(len(items))]
}

func (c *dummyDatabaseCommand) generateRandomBooks() (string, models.ID) {
	description := pickRandom(c.rng, books[:])
	categoryId := common.CategoryId_Books

	return description, categoryId
}

func (c *dummyDatabaseCommand) generateRandomToys() (string, models.ID) {
	description := pickRandom(c.rng, toys[:])
	categoryId := common.CategoryId_Toys

	return description, categoryId
}

func (c *dummyDatabaseCommand) addSales(db *sql.DB, cashierIds []models.ID, itemIds []models.ID) (r_err error) {
	c.Printf("Adding sales\n")

	transaction, err := queries.NewTransactionalDatabaseQuerier(context.Background(), db)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer transaction.RollbackIfNotCommitted()

	// Make copy because we need to shuffle it repeatedly
	itemIds = slices.Clone(itemIds)

	saleCount := c.rng.IntN(100) + 10
	times := c.generateChronologicalTimes(saleCount, 0, 60*60*24)
	for _, transactionTime := range times {
		cashierId := pickRandom(c.rng, cashierIds)
		itemCount := c.rng.IntN(20) + 1

		c.rng.Shuffle(len(itemIds), func(i, j int) {
			itemIds[i], itemIds[j] = itemIds[j], itemIds[i]
		})
		saleItems := itemIds[:itemCount]
		_, err := queries.AddSale(transaction, cashierId, transactionTime, saleItems)

		if err != nil {
			return fmt.Errorf("failed to add sale: %w", err)
		}
	}

	if err := transaction.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
