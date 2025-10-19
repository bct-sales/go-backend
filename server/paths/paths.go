package paths

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
)

type PathNode struct {
	parent      *PathNode
	pathSegment string
}

type QueryNode struct {
	key   string
	value string
	next  *QueryNode
}

type QueriedPath struct {
	parent *PathNode
	query  *QueryNode
}

func (p *PathNode) String() string {
	if p.parent == nil {
		return p.pathSegment
	}
	return p.parent.String() + "/" + p.pathSegment
}

func (p *PathNode) Descend(segment string) *PathNode {
	return &PathNode{
		parent:      p,
		pathSegment: segment,
	}
}

func (p *PathNode) Query(key, value string) *QueriedPath {
	return &QueriedPath{
		parent: p,
		query: &QueryNode{
			key:   key,
			value: value,
			next:  nil,
		},
	}
}

func (p *QueriedPath) Query(key, value string) *QueriedPath {
	return &QueriedPath{
		parent: p.parent,
		query: &QueryNode{
			key:   key,
			value: value,
			next:  p.query,
		},
	}
}

func (p *QueriedPath) String() string {
	result := p.parent.String()
	separator := "?"

	for q := p.query; q != nil; q = q.next {
		result += separator + q.key + "=" + q.value
		separator = "&"
	}

	return result
}

func NewRootPath(segment string) *PathNode {
	return &PathNode{
		parent:      nil,
		pathSegment: segment,
	}
}

type PPath struct {
	raw *PathNode
}

func (p PPath) String() string {
	return p.raw.String()
}

type RootPath struct {
	PPath
}

func Root() *URL {
	return NewURL()
}

func RESTRoot() *URL {
	return Root().WithExtraPathSegment("api").WithExtraPathSegment("v1")
}

func Login() *URL {
	return RESTRoot().WithExtraPathSegment("login")
}

func Logout() *URL {
	return RESTRoot().WithExtraPathSegment("logout")
}

func Labels() *URL {
	return RESTRoot().WithExtraPathSegment("labels")
}

func Users() *URL {
	return RESTRoot().WithExtraPathSegment("users")
}

func UserStr(userID string) *URL {
	return Users().WithExtraPathSegment(userID)
}

func User(id models.ID) *URL {
	return UserStr(id.String())
}

func Sales() *URL {
	return RESTRoot().WithExtraPathSegment("sales")
}

func SaleStr(saleID string) *URL {
	return Sales().WithExtraPathSegment(saleID)
}

func Sale(id models.ID) *URL {
	return SaleStr(id.String())
}

func SoldItems() *URL {
	return Sales().WithExtraPathSegment("items")
}

func Items() *URL {
	return RESTRoot().WithExtraPathSegment("items")
}

func ItemStr(itemID string) *URL {
	return Items().WithExtraPathSegment(itemID)
}

func Item(id models.ID) *URL {
	return ItemStr(id.String())
}

func Categories() *URL {
	return RESTRoot().WithExtraPathSegment("categories")
}

func CategoriesWithCounts(itemSelection queries.ItemSelection) *URL {
	switch itemSelection {
	case queries.AllItems:
		return Categories().AddQueryParameter("counts", "all")
	case queries.OnlyHiddenItems:
		return Categories().AddQueryParameter("counts", "hidden")
	case queries.OnlyVisibleItems:
		return Categories().AddQueryParameter("counts", "visible")
	default:
		panic("bug: unknown item selection")
	}
}

func CategoriesWithSoldItemCounts() *URL {
	return Categories().AddQueryParameter("counts", "sold")
}

func SellerItemsStr(sellerID string) *URL {
	return RESTRoot().WithExtraPathSegment("sellers").WithExtraPathSegment(sellerID).WithExtraPathSegment("items")
}

func SellerItems(sellerID models.ID) *URL {
	return SellerItemsStr(sellerID.String())
}

func CashierSalesStr(cashierID string) *URL {
	return RESTRoot().WithExtraPathSegment("cashiers").WithExtraPathSegment(cashierID).WithExtraPathSegment("sales")
}

func CashierSales(cashierID models.ID) *URL {
	return CashierSalesStr(cashierID.String())
}

func Websocket() *URL {
	return RESTRoot().WithExtraPathSegment("websocket")
}
