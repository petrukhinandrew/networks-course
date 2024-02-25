package v1

import (
	"fmt"
	"strconv"
)

const ResourcesDir = "./resources/"

type IdErr int

const (
	Ok        IdErr = 0
	NotParsed IdErr = 1
	NotExist  IdErr = 2
)

func tryIdToIdx(idAsStr string) (int, IdErr) {
	id, parsed := strconv.Atoi(idAsStr)
	if parsed != nil {
		return -1, NotParsed
	}
	return getProductIdxByID(id)

}

func getProductIdxByID(id int) (int, IdErr) {
	for idx, p := range products {
		if p.Id == id {
			return idx, Ok
		}
	}
	return -1, NotExist
}

func removeProductByIdx(idx int) Product {
	prod := products[idx]
	for _, p := range products {
		fmt.Printf("%d, ", p.Id)
	}
	fmt.Println()
	products = append(products[:idx], products[idx+1:]...)
	for _, p := range products {
		fmt.Printf("%d, ", p.Id)
	}
	fmt.Println()
	fmt.Println(prod.Id)
	return prod
}
