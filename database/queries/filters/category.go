package filters

import (
	"bctbackend/database/models"
)

type Category struct {
	categoryID         *models.ID
	descriptionPattern *string
}

func (filter *Category) WithCategory(category models.ID) {
	filter.categoryID = &category
}

// CategoryID returns the value set by WithCategory, or nil if it was never set.
func (filter *Category) CategoryID() *models.ID {
	return filter.categoryID
}

// WithDescriptionPattern stores a LIKE pattern for the item description.
// Note that the pattern is used as-is: callers add their own wildcards.
func (filter *Category) WithDescriptionPattern(pattern string) {
	filter.descriptionPattern = &pattern
}

// DescriptionPattern returns the value set by WithDescriptionPattern, or nil if it was never set.
func (filter *Category) DescriptionPattern() *string {
	return filter.descriptionPattern
}
