package v1

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func AddProduct(c *gin.Context) {
	var newData ProductData

	if err := c.ShouldBindJSON(&newData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	newProduct := Product{productsNextIDX, newData}
	productsNextIDX++

	products = append(products, newProduct)
	c.JSON(http.StatusOK, &newProduct)
}

func GetProducts(c *gin.Context) {
	c.JSON(http.StatusOK, products)
}

func handleIdErr(err IdErr, c *gin.Context, callback func(*gin.Context)) {
	switch err {
	case NotParsed:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad product ID format"})
		return
	case NotExist:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product with given ID not found"})
		return
	case Ok:
		callback(c)
	}
}

func GetProductByID(c *gin.Context) {
	reqIdAsStr, found := c.Params.Get("id")

	if !found {
		panic("id is not matched")
	}

	idx, err := tryIdToIdx(reqIdAsStr)

	handleIdErr(err, c, func(ctx *gin.Context) {
		c.JSON(http.StatusOK, products[idx])
	})

}

func DeleteProductByID(c *gin.Context) {
	reqIdAsStr, found := c.Params.Get("id")

	if !found {
		panic("id is not matched")
	}

	idx, err := tryIdToIdx(reqIdAsStr)
	handleIdErr(err, c, func(ctx *gin.Context) {
		prod := removeProductByIdx(idx)
		if prod.HasIcon() {
			os.Remove(ResourcesDir + prod.Icon)
		}
		c.JSON(http.StatusOK, &prod)
	})
}

func UpdateProductByID(c *gin.Context) {
	reqIdAsStr, found := c.Params.Get("id")

	if !found {
		panic("id is not matched")
	}

	idx, err := tryIdToIdx(reqIdAsStr)
	handleIdErr(err, c, func(ctx *gin.Context) {

		newData := ProductDataOpt{}
		prod := &products[idx]

		icon, iconErr := ctx.FormFile("icon")
		if iconErr == nil {
			(*prod).Icon = fmt.Sprintf("icon-%s%s", reqIdAsStr, filepath.Ext(icon.Filename))
			ctx.SaveUploadedFile(icon, ResourcesDir+products[idx].Icon)
		} else if err := c.BindJSON(&newData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if newData.Name != nil {
			(*prod).Name = *newData.Name
		}

		if newData.Description != nil {
			(*prod).Description = *newData.Description
		}

		c.JSON(http.StatusAccepted, prod)
	})
}

func AddProductImageById(c *gin.Context) {
	reqIdAsStr, found := c.Params.Get("id")

	if !found {
		panic("id is not matched")
	}

	idx, err := tryIdToIdx(reqIdAsStr)
	handleIdErr(err, c, func(ctx *gin.Context) {

		file, err := ctx.FormFile("icon")

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		products[idx].Icon = fmt.Sprintf("icon-%s%s", reqIdAsStr, filepath.Ext(file.Filename))

		ctx.SaveUploadedFile(file, ResourcesDir+products[idx].Icon)
		ctx.String(http.StatusOK, "Icon uploaded")
	})
}
func GetProductImageById(c *gin.Context) {
	reqIdAsStr, found := c.Params.Get("id")

	if !found {
		panic("id is not matched")
	}

	idx, err := tryIdToIdx(reqIdAsStr)

	handleIdErr(err, c, func(ctx *gin.Context) {
		prod := products[idx]
		if !prod.HasIcon() {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("No icon provided for ID %s", reqIdAsStr)})
			return
		}
		path := ResourcesDir + prod.Icon
		c.File(path)
	})
}
