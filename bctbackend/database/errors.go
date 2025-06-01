package database

import (
	"errors"
)

var ErrIdAlreadyInUse = errors.New("id already in use")
var ErrSaleMissingItems = errors.New("sale must have at least one item")
var ErrSaleRequiresCashier = errors.New("sale requires a cashier")
var ErrWrongPassword = errors.New("wrong password")
var ErrWrongRole = errors.New("user has a wrong role")
var ErrDuplicateItemInSale = errors.New("duplicate item in sale")
var ErrItemFrozen = errors.New("item is frozen")
var ErrItemHidden = errors.New("item is hidden")
var ErrHiddenFrozenItem = errors.New("items cannot be hidden and frozen at the same time")
var ErrDatabaseAlreadyExists = errors.New("database already exists")

var ErrNoSuchUser = errors.New("no such user")
var ErrNoSessionFound = errors.New("no session found")
var ErrNoSuchItem = errors.New("no such item")
var ErrNoSuchSale = errors.New("no such sale")
var ErrNoSuchSession = errors.New("no such session")
var ErrNoSuchCategory = errors.New("no such category")
var ErrNoSuchRole = errors.New("no such role")

var ErrInvalidPrice = errors.New("invalid price")
var ErrInvalidItemDescription = errors.New("invalid item description")
var ErrInvalidCategoryName = errors.New("category name is invalid")
