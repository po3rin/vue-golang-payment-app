package domain

// Item - set of item
type Item struct {
	ID          int64
	Name        string
	Discription string
	Amount      int64
}

// Items -set of item list
type Items []Item
