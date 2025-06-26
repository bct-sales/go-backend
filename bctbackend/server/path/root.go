package path

import "bctbackend/database/models"

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
	return Root().AddPathSegment("api").AddPathSegment("v1")
}

func Login() *URL {
	return RESTRoot().AddPathSegment("login")
}

func Logout() *URL {
	return RESTRoot().AddPathSegment("logout")
}

func Labels() *URL {
	return RESTRoot().AddPathSegment("labels")
}

func Users() *URL {
	return RESTRoot().AddPathSegment("users")
}

func UserStr(userId string) *URL {
	return Users().AddPathSegment(userId)
}

func User(id models.Id) *URL {
	return UserStr(id.String())
}

func Sales() *URL {
	return RESTRoot().AddPathSegment("sales")
}

func SaleStr(saleId string) *URL {
	return Sales().AddPathSegment(saleId)
}

func Sale(id models.Id) *URL {
	return SaleStr(id.String())
}

func Items() *URL {
	return RESTRoot().AddPathSegment("items")
}

func ItemStr(itemId string) *URL {
	return Items().AddPathSegment(itemId)
}

func Item(id models.Id) *URL {
	return ItemStr(id.String())
}
