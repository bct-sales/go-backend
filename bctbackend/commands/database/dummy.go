package database

import (
	"bctbackend/commands/common"
	"bctbackend/database"
	"bctbackend/database/models"
	"bctbackend/database/queries"
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

var colors = [...]string{"red", "green", "blue", "yellow", "orange", "purple", "black", "white"}

var clothing = [...]string{"t-shirt", "hoodie", "jacket", "sweater", "jeans", "pants", "shorts", "skirt", "dress", "hat", "scarf", "socks", "gloves"}

var clothingCategories = [...]models.Id{
	CategoryId_Clothing50_56,
	CategoryId_Clothing56_62,
	CategoryId_Clothing68_80,
	CategoryId_Clothing86_92,
	CategoryId_Clothing92_98,
	CategoryId_Clothing104_116,
	CategoryId_Clothing122_128,
	CategoryId_Clothing128_140,
	CategoryId_Clothing140_152,
}

var shoeTypes = [...]string{
	"Nike Air Max 90 sneakers",
	"Nike Air Max 95 sneakers",
	"Nike Air Max 97 sneakers",
	"Nike TN sneakers",
	"Nike P-6000 sneakers",
	"Nike Vomero sneakers",
	"Nike Pegasus sneakers",
	"Nike Air Force sneakers",
	"Adidas Superstar sneakers",
	"Adidas Gazelle sneakers",
	"Adidas Stan Smith sneakers",
	"Adidas Samba sneakers",
	"Adidas Ultraboost sneakers",
	"Adidas NMD sneakers",
	"Adidas Yeezy sneakers",
	"Adidas Ozweego sneakers",
	"Puma Suede sneakers",
	"Puma Speedcat sneakers",
	"Fila Disruptor sneakers",
	"Fila Bubbles sneakers",
	"Reebok Instapump Fury sneakers",
	"Reebok Kamikaze sneakers",
	"Converse All-Star sneakers",
	"Vans slip-ons",
	"Vans Old Skool",
	"Vans Sk8-Hi",
	"New Balance 530 sneakers",
	"New Balance 9060 sneakers",
	"Asics Kayano 14 sneakers",
	"Hoka sneakers",
	"Saucony sneakers",
	"Brooks sneakers",
	"Mizuno sneakers",
	"On Cloudmonster sneakers",
	"On Cloudsurfer sneakers",
	"Jordan sneakers",
	"Under Armour sneakers",
	"Crocs Classic clogs",
	"Crocs Echo Clogs",
	"Crocs Crush Clogs",
}

var bootBrands = [...]string{
	"Sendra", "Timberland", "Dr. Martens", "Solovair", "New Rock", "Red Wing", "Frye", "Ariat", "Justin", "Lucchese", "Carmina", "Tony Mora",
}

var bootTypes = [...]string{
	"Combat boots", "Cowboy boots", "Harness boots", "Engineer boots", "Jodhpur boots", "Chelsea boots",
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
	seed uint64
	rng  *rand.Rand
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
			},
		},
	}

	command.CobraCommand.Flags().Uint64Var(&command.seed, "seed", 0, "Seed for random number generation")

	return command.AsCobraCommand()
}

func (c *dummyDatabaseCommand) execute() error {
	c.rng = rand.New(rand.NewPCG(0, c.seed))

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

		cashierIds, err := c.addCashiers(db)
		if err != nil {
			return err
		}

		sellerIds, err := c.addSellers(db)
		if err != nil {
			return err
		}

		itemIds, err := c.addItems(db, sellerIds)
		if err != nil {
			return err
		}

		if err := c.addSales(db, cashierIds, itemIds); err != nil {
			return err
		}

		return nil
	})
}

func (c *dummyDatabaseCommand) addCategories(db *sql.DB) error {
	c.Printf("Adding categories\n")

	addCategory := func(id models.Id, name string) error {
		return queries.AddCategoryWithId(db, id, name)
	}

	if err := GenerateDefaultCategories(addCategory); err != nil {
		return fmt.Errorf("failed to add categories: %w", err)
	}
	return nil
}

func (c *dummyDatabaseCommand) addAdmin(db *sql.DB) (models.Id, error) {
	c.Printf("Adding admin user\n")

	id := models.Id(1)
	roleId := models.NewAdminRoleId()
	createdAt := models.Now()
	var lastActivity *models.Timestamp = nil
	password := "abc"

	if err := queries.AddUserWithId(db, id, roleId, createdAt, lastActivity, password); err != nil {
		return 0, fmt.Errorf("failed to add admin: %w", err)
	}

	return id, nil
}

func (c *dummyDatabaseCommand) addCashiers(db *sql.DB) ([]models.Id, error) {
	c.Printf("Adding cashier users\n")

	cashierCount := c.rng.IntN(10) + 1
	cashierIDs := make([]models.Id, 0, cashierCount)

	for range cashierCount {
		roleId := models.NewCashierRoleId()
		createdAt := models.Now()
		var lastActivity *models.Timestamp = nil
		password := "abc"

		cashierId, err := queries.AddUser(db, roleId, createdAt, lastActivity, password)

		if err != nil {
			return nil, fmt.Errorf("failed to add cashier: %w", err)
		}

		cashierIDs = append(cashierIDs, cashierId)
	}

	return cashierIDs, nil
}

func (c *dummyDatabaseCommand) addSellers(db *sql.DB) ([]models.Id, error) {
	c.Printf("Adding sellers\n")

	sellerIds := make([]models.Id, 0, zoneCount*10)

	addSellers := func(addUser func(userId models.Id, roleId models.RoleId, createdAt models.Timestamp, lastActivity *models.Timestamp, password string)) {
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

func (c *dummyDatabaseCommand) getSellerId(zone int, offset int) models.Id {
	return models.Id(zone*100 + offset)
}

func (c *dummyDatabaseCommand) addItems(db *sql.DB, sellerIds []models.Id) ([]models.Id, error) {
	c.Printf("Adding items\n")

	itemIds := make([]models.Id, 0, 1000)
	for _, sellerId := range sellerIds {
		itemCount := c.rng.IntN(20)

		for range itemCount {
			description, category := c.generateRandomItemDescriptionAndCategory()
			priceInCents := models.MoneyInCents(c.rng.IntN(100)*50 + 50)
			donation := c.rng.IntN(20) == 0
			charity := c.rng.IntN(20) == 0
			frozen := c.rng.IntN(20) == 0
			hidden := false
			addedAt := models.Now()

			itemId, err := queries.AddItem(db, addedAt, description, priceInCents, category, sellerId, donation, charity, frozen, hidden)
			if err != nil {
				return nil, fmt.Errorf("failed to add item: %w", err)
			}

			itemIds = append(itemIds, itemId)
		}
	}

	return itemIds, nil
}

func (c *dummyDatabaseCommand) generateRandomItemDescriptionAndCategory() (string, models.Id) {
	switch c.rng.IntN(3) {
	case 0:
		return c.generateRandomClothing()
	case 1:
		return c.generateRandomShoes()
	default:
		return c.generateRandomToys()
	}
}

func (c *dummyDatabaseCommand) generateRandomColor() string {
	return colors[c.rng.IntN(len(colors))]
}

func (c *dummyDatabaseCommand) generateRandomClothingCategoryId() models.Id {
	return pickRandom(c.rng, clothingCategories[:])
}

func (c *dummyDatabaseCommand) generateRandomClothing() (string, models.Id) {
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

func (c *dummyDatabaseCommand) generateRandomShoes() (string, models.Id) {
	switch c.rng.IntN(2) {
	case 0:
		color := c.generateRandomColor()
		shoeType := pickRandom(c.rng, shoeTypes[:])
		description := fmt.Sprintf("%s %s", color, shoeType)
		return description, CategoryId_Shoes
	default:
		brand := pickRandom(c.rng, bootBrands[:])
		bootType := pickRandom(c.rng, bootTypes[:])
		description := fmt.Sprintf("%s %s", brand, bootType)
		return description, CategoryId_Shoes
	}
}

func (c *dummyDatabaseCommand) generateRandomToys() (string, models.Id) {
	description := pickRandom(c.rng, toys[:])
	categoryId := CategoryId_Toys

	return description, categoryId
}

func (c *dummyDatabaseCommand) addSales(db *sql.DB, cashierIds []models.Id, itemIds []models.Id) error {
	c.Printf("Adding sales\n")

	// Make copy because we need to shuffle it repeatedly
	itemIds = slices.Clone(itemIds)

	saleCount := c.rng.IntN(100) + 10
	for range saleCount {
		cashierId := pickRandom(c.rng, cashierIds)
		itemCount := c.rng.IntN(20) + 1

		c.rng.Shuffle(len(itemIds), func(i, j int) {
			itemIds[i], itemIds[j] = itemIds[j], itemIds[i]
		})
		saleItems := itemIds[:itemCount]
		transactionTime := models.Now()
		_, err := queries.AddSale(db, cashierId, transactionTime, saleItems)

		if err != nil {
			return fmt.Errorf("failed to add sale: %w", err)
		}
	}

	return nil
}
