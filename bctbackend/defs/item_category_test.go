package defs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNameOfCategory(t *testing.T) {
	pairs := []struct {
		id   Id
		name string
	}{
		{Clothing50_56, Clothing50_56Name},
		{Clothing56_62, Clothing56_62Name},
		{Clothing68_80, Clothing68_80Name},
		{Clothing86_92, Clothing86_92Name},
		{Clothing92_98, Clothing92_98Name},
		{Clothing104_116, Clothing104_116Name},
		{Clothing122_128, Clothing122_128Name},
		{Clothing128_140, Clothing128_140Name},
		{Clothing140_152, Clothing140_152Name},
		{Shoes, ShoesName},
		{Toys, ToysName},
		{BabyChildEquipment, BabyChildEquipmentName},
	}

	for _, pair := range pairs {
		id := pair.id
		expectedName := pair.name

		t.Run(expectedName, func(t *testing.T) {
			actualName, err := NameOfCategory(id)

			if assert.NoError(t, err) {
				assert.Equal(t, expectedName, actualName)
			}
		})
	}

	t.Run("invalid category", func(t *testing.T) {
		_, err := NameOfCategory(Id(999))

		assert.Error(t, err)
	})
}
