//go:build test

package helpers

import (
	models "bctbackend/database/models"
)

func ItemsTotalWorth(items []*models.Item) models.MoneyInCents {
	total := models.MoneyInCents(0)

	for _, item := range items {
		total += item.PriceInCents
	}

	return total
}
