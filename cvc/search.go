package cvc

import (
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/pasztorpisti/qs"
)

type Filter struct {
	DisplayHomeSchool      bool     `qs:"filter[display_home_school]" json:"display_home_school"`
	UniversityID           int      `qs:"filter[university_id]" json:"university_id"`
	SearchType             string   `qs:"filter[search_type]" json:"search_type"`
	Subject                string   `qs:"filter[subject]" json:"subject"`
	OeiPhase2Filter        []bool   `qs:"filter[oei_phase_2_filter]" json:"oei_phase_2_filter"`
	ShowOnlyAvailable      []bool   `qs:"filter[show_only_available]" json:"show_only_available"`
	DeliveryMethods        []string `qs:"filter[delivery_methods][]" json:"delivery_methods"`
	DeliveryMethodSubtypes []string `qs:"filter[delivery_method_subtypes][]" json:"delivery_method_subtypes"`
	Prerequisites          []string `qs:"filter[prerequisites][]" json:"prerequisites"`
	SessionNames           []string `qs:"filter[session_names][]" json:"session_names"`
	ZeroTextbookCostFilter bool     `qs:"filter[zero_textbook_cost_filter]" json:"zero_textbook_cost_filter"`
	StartDate              string   `qs:"filter[start_date]" json:"start_date"`
	EndDate                string   `qs:"filter[end_date]" json:"end_date"`
	TargetSchoolIDs        []string `qs:"filter[target_school_ids][]" json:"target_school_ids"`
	MinCreditsRange        int      `qs:"filter[min_credits_range]" json:"min_credits_range"`
	MaxCreditsRange        int      `qs:"filter[max_credits_range]" json:"max_credits_range"`
}

type searchRequest struct {
	Filter
	Page        int    `qs:"page"`
	RandomToken string `qs:"random_token"`
}

func (c *Client) Search(filter Filter) ([]Course, error) {
	req, err := http.NewRequest(http.MethodGet, "https://search.cvc.edu/search", nil)
	if err != nil {
		return nil, err
	}

	query, err := qs.Marshal(searchRequest{
		Filter: filter,
		Page:   1,
	})
	if err != nil {
		return nil, err
	}

	req.URL.RawQuery = query

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
		name := s.Find(`.course-details-link`).Text()
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
			Name:    strings.TrimSpace(name),
			Units:   strings.TrimSpace(units),
			Term:    strings.TrimSpace(term),
			Cost:    strings.TrimSpace(cost),
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
}
