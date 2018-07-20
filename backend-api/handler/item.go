package handler

import (
	"net/http"
	"vue-golang-payment-app/backend-api/domain"
)

// GetItem - get item by id
func GetItem(c Context) {
	resp := &domain.Item{
		Name:        "testItem",
		Discription: "this is a test item",
		Amount:      1200,
	}
	c.JSON(http.StatusOK, resp)
}

// GetLists - get all items
func GetLists(c Context) {
	resp1 := &domain.Item{
		Name:        "testItem",
		Discription: "this is a test item",
		Amount:      1200,
	}
	resp2 := &domain.Item{
		Name:        "testToy",
		Discription: "this is a test toy",
		Amount:      1500,
	}
	lists := domain.Items{
		resp1,
		resp2,
	}
	c.JSON(http.StatusOK, lists)
}
