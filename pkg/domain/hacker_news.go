package domain

import "fmt"

type NewsItem struct {
	Rank     int
	ID       string
	Title    string
	URL      string
	Site     string
	Score    int
	Author   string
	Age      string
	Comments int
}

func (n NewsItem) ToText() string {
	return fmt.Sprintf(
		"%d. %s (%s)\nPoints: %d | Author: %s | Age: %s | Comments: %d",
		n.Rank, n.Title, n.Site, n.Score, n.Author, n.Age, n.Comments,
	)
}
