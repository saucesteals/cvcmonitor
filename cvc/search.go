package cvc

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func (c *Client) Search(query url.Values) ([]Course, error) {
	req, err := http.NewRequest(http.MethodGet, "https://search.cvc.edu/search", nil)
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	req.URL.RawQuery = query.Encode()

	res, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	document, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	var courses []Course

	document.Find("#search-results .course").Each(func(i int, s *goquery.Selection) {
		college := s.Find(`.course-head .text-sm`).Text()
		nameSelection := s.Find(`.course-details-link`)
		units := s.Find(`.credit > p`).Text()
		term := s.Find(`.term > p`).Text()

		var cost string
		{
			nodes := s.Find(`.text-c_link`).Nodes
			if len(nodes) > 0 {
				cost = nodes[0].FirstChild.NextSibling.Data
			}
		}

		courses = append(courses, Course{
			College: strings.TrimSpace(college),
			Name:    strings.TrimSpace(nameSelection.Text()),
			Units:   strings.TrimSpace(units),
			Term:    strings.TrimSpace(term),
			Cost:    strings.TrimSpace(cost),
			Link:    fmt.Sprintf("https://search.cvc.edu%s", nameSelection.AttrOr("href", "")),
		})
	})

	return courses, nil
}

type Course struct {
	Name    string
	College string
	Units   string
	Term    string
	Cost    string
	Link    string
}
