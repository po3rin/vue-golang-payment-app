package domain

// Item - set of item
type Item struct {
	Name        string
	Discription string
	Amount      int
}

// Items -set of item list
type Items []*Item
