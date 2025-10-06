package common

import (
	"bctbackend/database/models"
)

const (
	CategoryID_Clothing50_56      models.ID = 1
	CategoryID_Clothing56_62      models.ID = 2
	CategoryID_Clothing68_80      models.ID = 3
	CategoryID_Clothing86_92      models.ID = 4
	CategoryID_Clothing92_98      models.ID = 5
	CategoryID_Clothing104_116    models.ID = 6
	CategoryID_Clothing122_128    models.ID = 7
	CategoryID_Clothing128_140    models.ID = 8
	CategoryID_Clothing140_152    models.ID = 9
	CategoryId_Books              models.ID = 10
	CategoryId_Toys               models.ID = 11
	CategoryId_BabyChildEquipment models.ID = 12
	CategoryId_Maternity          models.ID = 13
	CategoryId_Large              models.ID = 14

	CategoryName_Clothing50_56      string = "Clothing 0-3 mos (50-56)"
	CategoryName_Clothing56_62      string = "Clothing 3-6 mos (56-62)"
	CategoryName_Clothing68_80      string = "Clothing 6-12 mos (68-80)"
	CategoryName_Clothing86_92      string = "Clothing 12-24 mos (86-92)"
	CategoryName_Clothing92_98      string = "Clothing 2-3 yrs (92-98)"
	CategoryName_Clothing104_116    string = "Clothing 4-6 yrs (104-116)"
	CategoryName_Clothing122_128    string = "Clothing 7-8 yrs (122-128)"
	CategoryName_Clothing128_140    string = "Clothing 9-10 yrs (128-140)"
	CategoryName_Clothing140_152    string = "Clothing 11-12 yrs (140-152)"
	CategoryName_Books              string = "Books"
	CategoryName_Toys               string = "Toys"
	CategoryName_BabyChildEquipment string = "Baby/Child Equipment"
	CategoryName_Maternity          string = "Maternity"
	CategoryName_Large              string = "Large Item"
)

func ListCategoryIDs() []models.ID {
	return []models.ID{
		CategoryID_Clothing50_56,
		CategoryID_Clothing56_62,
		CategoryID_Clothing68_80,
		CategoryID_Clothing86_92,
		CategoryID_Clothing92_98,
		CategoryID_Clothing104_116,
		CategoryID_Clothing122_128,
		CategoryID_Clothing128_140,
		CategoryID_Clothing140_152,
		CategoryId_Books,
		CategoryId_Toys,
		CategoryId_BabyChildEquipment,
		CategoryId_Maternity,
		CategoryId_Large,
	}
}

func GenerateDefaultCategories(callback func(id models.ID, name string) error) error {
	if err := callback(CategoryID_Clothing50_56, CategoryName_Clothing50_56); err != nil {
		return err
	}
	if err := callback(CategoryID_Clothing56_62, CategoryName_Clothing56_62); err != nil {
		return err
	}
	if err := callback(CategoryID_Clothing68_80, CategoryName_Clothing68_80); err != nil {
		return err
	}
	if err := callback(CategoryID_Clothing86_92, CategoryName_Clothing86_92); err != nil {
		return err
	}
	if err := callback(CategoryID_Clothing92_98, CategoryName_Clothing92_98); err != nil {
		return err
	}
	if err := callback(CategoryID_Clothing104_116, CategoryName_Clothing104_116); err != nil {
		return err
	}
	if err := callback(CategoryID_Clothing122_128, CategoryName_Clothing122_128); err != nil {
		return err
	}
	if err := callback(CategoryID_Clothing128_140, CategoryName_Clothing128_140); err != nil {
		return err
	}
	if err := callback(CategoryID_Clothing140_152, CategoryName_Clothing140_152); err != nil {
		return err
	}
	if err := callback(CategoryId_Books, CategoryName_Books); err != nil {
		return err
	}
	if err := callback(CategoryId_Toys, CategoryName_Toys); err != nil {
		return err
	}
	if err := callback(CategoryId_BabyChildEquipment, CategoryName_BabyChildEquipment); err != nil {
		return err
	}
	if err := callback(CategoryId_Maternity, CategoryName_Maternity); err != nil {
		return err
	}
	if err := callback(CategoryId_Large, CategoryName_Large); err != nil {
		return err
	}

	return nil
}
