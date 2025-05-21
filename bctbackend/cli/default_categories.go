package cli

import (
	"bctbackend/database/models"
)

type Id = int64

const (
	CategoryId_Clothing50_56      Id = 1
	CategoryId_Clothing56_62      Id = 2
	CategoryId_Clothing68_80      Id = 3
	CategoryId_Clothing86_92      Id = 4
	CategoryId_Clothing92_98      Id = 5
	CategoryId_Clothing104_116    Id = 6
	CategoryId_Clothing122_128    Id = 7
	CategoryId_Clothing128_140    Id = 8
	CategoryId_Clothing140_152    Id = 9
	CategoryId_Shoes              Id = 10
	CategoryId_Toys               Id = 11
	CategoryId_BabyChildEquipment Id = 12

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

func ListCategoryIds() []Id {
	return []Id{
		CategoryId_Clothing50_56,
		CategoryId_Clothing56_62,
		CategoryId_Clothing68_80,
		CategoryId_Clothing86_92,
		CategoryId_Clothing92_98,
		CategoryId_Clothing104_116,
		CategoryId_Clothing122_128,
		CategoryId_Clothing128_140,
		CategoryId_Clothing140_152,
		CategoryId_Shoes,
		CategoryId_Toys,
		CategoryId_BabyChildEquipment,
	}
}

func GenerateDefaultCategories(callback func(id models.Id, name string) error) error {
	if err := callback(CategoryId_Clothing50_56, CategoryName_Clothing50_56); err != nil {
		return err
	}
	if err := callback(CategoryId_Clothing56_62, CategoryName_Clothing56_62); err != nil {
		return err
	}
	if err := callback(CategoryId_Clothing68_80, CategoryName_Clothing68_80); err != nil {
		return err
	}
	if err := callback(CategoryId_Clothing86_92, CategoryName_Clothing86_92); err != nil {
		return err
	}
	if err := callback(CategoryId_Clothing92_98, CategoryName_Clothing92_98); err != nil {
		return err
	}
	if err := callback(CategoryId_Clothing104_116, CategoryName_Clothing104_116); err != nil {
		return err
	}
	if err := callback(CategoryId_Clothing122_128, CategoryName_Clothing122_128); err != nil {
		return err
	}
	if err := callback(CategoryId_Clothing128_140, CategoryName_Clothing128_140); err != nil {
		return err
	}
	if err := callback(CategoryId_Clothing140_152, CategoryName_Clothing140_152); err != nil {
		return err
	}
	if err := callback(CategoryId_Shoes, CategoryName_Shoes); err != nil {
		return err
	}
	if err := callback(CategoryId_Toys, CategoryName_Toys); err != nil {
		return err
	}
	if err := callback(CategoryId_BabyChildEquipment, CategoryName_BabyChildEquipment); err != nil {
		return err
	}

	return nil
}
