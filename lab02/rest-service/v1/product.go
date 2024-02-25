package v1

type Product struct {
	Id int `json:"id"`
	ProductData
}

var products = make([]Product, 0)
var productsNextIDX = 0

type ProductData struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description" binding:"required"`
	Icon        string `json:"icon"`
}

type ProductDataOpt struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Icon        *string `json:"icon"`
}

func (p *Product) HasIcon() bool {
	return p.Icon != ""
}
