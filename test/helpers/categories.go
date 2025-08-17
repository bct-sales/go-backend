package helpers

import (
	"bctbackend/database/models"
	"maps"
	"slices"
)

const (
	CategoryId_Clothing50_56      models.Id = 1
	CategoryId_Clothing56_62      models.Id = 2
	CategoryId_Clothing68_80      models.Id = 3
	CategoryId_Clothing86_92      models.Id = 4
	CategoryId_Clothing92_98      models.Id = 5
	CategoryId_Clothing104_116    models.Id = 6
	CategoryId_Clothing122_128    models.Id = 7
	CategoryId_Clothing128_140    models.Id = 8
	CategoryId_Clothing140_152    models.Id = 9
	CategoryId_Shoes              models.Id = 10
	CategoryId_Toys               models.Id = 11
	CategoryId_BabyChildEquipment models.Id = 12

	CategoryName_Clothing50_56      string = "Clothing 0-3 mos (50-56)"
	CategoryName_Clothing56_62      string = "Clothing 3-6 mos (56-62)"
	CategoryName_Clothing68_80      string = "Clothing 6-12 mos (68-80)"
	CategoryName_Clothing86_92      string = "Clothing 12-24 mos (86-92)"
	CategoryName_Clothing92_98      string = "Clothing 2-3 yrs (92-98)"
	CategoryName_Clothing104_116    string = "Clothing 4-6 yrs (104-116)"
	CategoryName_Clothing122_128    string = "Clothing 7-8 yrs (122-128)"
	CategoryName_Clothing128_140    string = "Clothing 9-10 yrs (128-140)"
	CategoryName_Clothing140_152    string = "Clothing 11-12 yrs (140-152)"
	CategoryName_Shoes              string = "Shoes (infant to 12 yrs)"
	CategoryName_Toys               string = "Toys"
	CategoryName_BabyChildEquipment string = "Baby/Child Equipment"
)

func DefaultCategoryNameTable() map[models.Id]string {
	return map[models.Id]string{
		CategoryId_Clothing50_56:      CategoryName_Clothing50_56,
		CategoryId_Clothing56_62:      CategoryName_Clothing56_62,
		CategoryId_Clothing68_80:      CategoryName_Clothing68_80,
		CategoryId_Clothing86_92:      CategoryName_Clothing86_92,
		CategoryId_Clothing92_98:      CategoryName_Clothing92_98,
		CategoryId_Clothing104_116:    CategoryName_Clothing104_116,
		CategoryId_Clothing122_128:    CategoryName_Clothing122_128,
		CategoryId_Clothing128_140:    CategoryName_Clothing128_140,
		CategoryId_Clothing140_152:    CategoryName_Clothing140_152,
		CategoryId_Shoes:              CategoryName_Shoes,
		CategoryId_Toys:               CategoryName_Toys,
		CategoryId_BabyChildEquipment: CategoryName_BabyChildEquipment,
	}
}

func DefaultCategoryIds() []models.Id {
	return slices.Collect(maps.Keys(DefaultCategoryNameTable()))
}
