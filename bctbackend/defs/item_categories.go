package defs

import (
	"bctbackend/database/models"
	"fmt"
)

type Id = int64

const (
	Clothing50_56      Id = 1
	Clothing56_62      Id = 2
	Clothing68_80      Id = 3
	Clothing86_92      Id = 4
	Clothing92_98      Id = 5
	Clothing104_116    Id = 6
	Clothing122_128    Id = 7
	Clothing128_140    Id = 8
	Clothing140_152    Id = 9
	Shoes              Id = 10
	Toys               Id = 11
	BabyChildEquipment Id = 12

	Clothing50_56Name      string = "Clothing 0-3 mos (50-56)"
	Clothing56_62Name      string = "Clothing 3-6 mos (56-62)"
	Clothing68_80Name      string = "Clothing 6-12 mos (68-80)"
	Clothing86_92Name      string = "Clothing 12-24 mos (86-92)"
	Clothing92_98Name      string = "Clothing 2-3 yrs (92-98)"
	Clothing104_116Name    string = "Clothing 4-6 yrs (104-116)"
	Clothing122_128Name    string = "Clothing 7-8 yrs (122-128)"
	Clothing128_140Name    string = "Clothing 9-10 yrs (128-140)"
	Clothing140_152Name    string = "Clothing 11-12 yrs (140-152)"
	ShoesName              string = "Shoes (infant to 12 yrs)"
	ToysName               string = "Toys"
	BabyChildEquipmentName string = "Baby/Child Equipment"
)

func ListCategoryIds() []Id {
	return []Id{
		Clothing50_56,
		Clothing56_62,
		Clothing68_80,
		Clothing86_92,
		Clothing92_98,
		Clothing104_116,
		Clothing122_128,
		Clothing128_140,
		Clothing140_152,
		Shoes,
		Toys,
		BabyChildEquipment,
	}
}

func GenerateCategories(callback func(id models.Id, name string) error) error {
	for _, id := range ListCategoryIds() {
		name, err := NameOfCategory(id)
		if err != nil {
			panic(err)
		}

		if err := callback(id, name); err != nil {
			return err
		}
	}

	return nil
}

type UnknownCategoryError struct {
	CategoryId Id
}

func (e *UnknownCategoryError) Error() string {
	return fmt.Sprintf("unknown category id: %v", e.CategoryId)
}

// NameOfCategory returns the name of the category with the given id.
// If the category id is unknown, an UnknownCategoryError is returned.
func NameOfCategory(categoryId Id) (string, error) {
	switch categoryId {
	case Clothing50_56:
		return Clothing50_56Name, nil
	case Clothing56_62:
		return Clothing56_62Name, nil
	case Clothing68_80:
		return Clothing68_80Name, nil
	case Clothing86_92:
		return Clothing86_92Name, nil
	case Clothing92_98:
		return Clothing92_98Name, nil
	case Clothing104_116:
		return Clothing104_116Name, nil
	case Clothing122_128:
		return Clothing122_128Name, nil
	case Clothing128_140:
		return Clothing128_140Name, nil
	case Clothing140_152:
		return Clothing140_152Name, nil
	case Shoes:
		return ShoesName, nil
	case Toys:
		return ToysName, nil
	case BabyChildEquipment:
		return BabyChildEquipmentName, nil
	default:
		return "", &UnknownCategoryError{CategoryId: categoryId}
	}
}
