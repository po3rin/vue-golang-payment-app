package db

import (
	"fmt"
	"vue-golang-payment-app/backend-api/domain"
)

// SelectAllItems - select all
func SelectAllItems() (items domain.Items, err error) {
	stmt, err := Conn.Query("SELECT * FROM items")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer stmt.Close()
	for stmt.Next() {
		var id int64
		var name string
		var desctiption string
		var amount int64
		if err := stmt.Scan(&id, &name, &desctiption, &amount); err != nil {
			continue
		}
		item := domain.Item{
			ID:          id,
			Name:        name,
			Description: desctiption,
			Amount:      amount,
		}
		items = append(items, item)
	}
	return
}

// SelectItem - select post
func SelectItem(identifier int64) (item domain.Item, err error) {
	stmt, err := Conn.Prepare(fmt.Sprintf("SELECT * FROM items WHERE id = ? LIMIT 1"))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer stmt.Close()
	var id int64
	var name string
	var desctiption string
	var amount int64
	err = stmt.QueryRow(identifier).Scan(&id, &name, &desctiption, &amount)
	if err != nil {
		return
	}
	item.ID = id
	item.Name = name
	item.Description = desctiption
	item.Amount = amount
	return
}
