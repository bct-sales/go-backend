package models

import (
	"testing"

	"bctbackend/defs"

	"github.com/stretchr/testify/assert"
)

func TestNameOfCategory(t *testing.T) {
	pairs := []struct {
		id   Id
		name string
	}{
		{defs.Clothing50_56, defs.Clothing50_56Name},
		{defs.Clothing56_62, defs.Clothing56_62Name},
		{defs.Clothing68_80, defs.Clothing68_80Name},
		{defs.Clothing86_92, defs.Clothing86_92Name},
		{defs.Clothing92_98, defs.Clothing92_98Name},
		{defs.Clothing104_116, defs.Clothing104_116Name},
		{defs.Clothing122_128, defs.Clothing122_128Name},
		{defs.Clothing128_140, defs.Clothing128_140Name},
		{defs.Clothing140_152, defs.Clothing140_152Name},
		{defs.Shoes, defs.ShoesName},
		{defs.Toys, defs.ToysName},
		{defs.BabyChildEquipment, defs.BabyChildEquipmentName},
	}

	for _, pair := range pairs {
		id := pair.id
		expectedName := pair.name

		t.Run(expectedName, func(t *testing.T) {
			actualName, err := defs.NameOfCategory(id)

			if assert.NoError(t, err) {
				assert.Equal(t, expectedName, actualName)
			}
		})
	}

	t.Run("invalid category", func(t *testing.T) {
		_, err := defs.NameOfCategory(Id(999))

		assert.Error(t, err)
	})
}
