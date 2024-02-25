package main

import (
	v1 "petrukhinandrew/rest-service/v1"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	router.GET("/product/:id", v1.GetProductByID)
	router.GET("/products", v1.GetProducts)
	router.POST("/product", v1.AddProduct)
	router.DELETE("/product/:id", v1.DeleteProductByID)
	router.PUT("/product/:id", v1.UpdateProductByID)

	router.POST("/product/:id/image", v1.AddProductImageById)
	router.GET("/product/:id/image", v1.GetProductImageById)
	router.Run("mkn.edu:80")
}
