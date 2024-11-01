package models

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
)

type ItemCategory struct {
	CategoryId Id
	Name       string
}

func NewCategory(
	id Id,
	name string) *ItemCategory {

	return &ItemCategory{
		CategoryId: id,
		Name:       name,
	}
}

func Categories() []Id {
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
