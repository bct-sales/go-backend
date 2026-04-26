package meta

var SaleItem = SalesMetadata{
	Table:  "sale_items",
	SaleID: "sale_id",
	ItemID: "item_id",
}

type SalesMetadata struct {
	Table  string
	SaleID string
	ItemID string
}
