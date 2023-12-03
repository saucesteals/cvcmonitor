package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/webhook"
	"github.com/joho/godotenv"
	"github.com/saucesteals/cvcmonitor/cvc"
)

var (
	hook    webhook.Client
	queries []url.Values
)

func init() {
	godotenv.Load()

	var err error
	hook, err = webhook.NewWithURL(os.Getenv("DISCORD_WEBHOOK_URL"))
	if err != nil {
		log.Panic(err)
	}

	queries, err = getSearches()
	if err != nil {
		log.Panic(err)
	}
}

func main() {
	godotenv.Load()

	for _, query := range queries {
		go watch(query)
	}

	select {}
}

func watch(query url.Values) {
	name := query.Get("filter[query]")
	if name == "" {
		name = query.Get("filter[subject]")
	}

	log := log.New(os.Stdout, fmt.Sprintf("- [%s] - ", name), log.Ltime|log.Lmsgprefix)

	c := cvc.NewClient()

	var last []cvc.Course
	interval := time.Minute * 5

	for {
		log.Printf("Searching for courses...")
		courses, err := c.Search(query)
		if err != nil {
			log.Println(err)
			continue
		}

		for _, course := range courses {
			if !hasCourse(last, course) {
				_, err := hook.CreateEmbeds([]discord.Embed{embed(course, query)})
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

func embed(course cvc.Course, query url.Values) discord.Embed {
	inline := true
	t := time.Now()
	return discord.Embed{
		Title: fmt.Sprintf("%s - %s", course.College, course.Name),
		URL:   course.Link,
		Author: &discord.EmbedAuthor{
			Name:    "CVC Exchange",
			URL:     fmt.Sprintf("https://search.cvc.edu/search?%s", query.Encode()),
			IconURL: "https://github.com/saucesteals/cvcmonitor/blob/main/assets/logo.png?raw=true",
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
		Timestamp: &t,
		Color:     0xffc92a,
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

func getSearches() ([]url.Values, error) {
	name := "SEARCH_URL"

	searchUrls := os.Getenv(name)

	if searchUrls == "" {
		return nil, fmt.Errorf("%s is empty", name)
	}

	var queries []url.Values
	for _, searchUrl := range strings.Split(searchUrls, ",") {
		u, err := url.Parse(searchUrl)
		if err != nil {
			return nil, err
		}

		queries = append(queries, u.Query())
	}

	return queries, nil
}
