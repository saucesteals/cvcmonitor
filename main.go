package main

import (
	"log"
	"net/url"
	"os"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/webhook"
	"github.com/joho/godotenv"
	"github.com/pasztorpisti/qs"
	"github.com/saucesteals/cvcmonitor/cvc"
)

func main() {
	godotenv.Load()

	c := cvc.NewClient()

	searchURL, filter, err := searchFilter()
	if err != nil {
		log.Panic(err)
	}

	hook, err := webhook.NewWithURL(os.Getenv("DISCORD_WEBHOOK_URL"))
	if err != nil {
		log.Panic(err)
	}

	last, err := c.Search(filter)
	if err != nil {
		log.Panic(err)
	}

	interval := time.Minute * 5

	for {
		log.Printf("Searching for courses...")
		courses, err := c.Search(filter)
		if err != nil {
			log.Println(err)
			continue
		}

		for _, course := range courses {
			if !hasCourse(last, course) {
				_, err := hook.CreateEmbeds([]discord.Embed{embed(course, searchURL)})
				if err != nil {
					log.Println(err)
					continue
				}
			}
		}

		last = courses

		log.Printf("Sleeping for %.0f minutes...", interval.Minutes())
		time.Sleep(interval)
	}
}

func embed(course cvc.Course, searchURL string) discord.Embed {
	inline := true
	return discord.Embed{
		Title: course.Name,
		URL:   searchURL,
		Author: &discord.EmbedAuthor{
			Name: course.College,
		},
		Fields: []discord.EmbedField{
			{
				Name:   "Units",
				Value:  course.Units,
				Inline: &inline,
			},
			{
				Name:   "Term",
				Value:  course.Term,
				Inline: &inline,
			},
			{
				Name:   "Cost",
				Value:  course.Cost,
				Inline: &inline,
			},
		},
		Footer: &discord.EmbedFooter{
			IconURL: "https://github.com/saucesteals/cvcmonitor/blob/main/assets/logo.png?raw=true",
			Text:    "CVC Exchange - https://github.com/saucesteals/cvcmonitor",
		},
		Color: 0xffc92a,
	}
}

func hasCourse(courses []cvc.Course, course cvc.Course) bool {
	for _, c := range courses {
		if c.College == course.College && c.Name == course.Name && c.Term == course.Term {
			return true
		}
	}

	return false
}

func searchFilter() (string, cvc.Filter, error) {
	searchURL := os.Getenv("SEARCH_URL")

	u, err := url.Parse(searchURL)
	if err != nil {
		return "", cvc.Filter{}, err
	}

	var filter cvc.Filter
	if err := qs.UnmarshalValues(&filter, u.Query()); err != nil {
		return "", cvc.Filter{}, err
	}

	return searchURL, filter, nil
}
