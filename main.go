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
	"github.com/pasztorpisti/qs"
	"github.com/saucesteals/cvcmonitor/cvc"
)

var (
	hook     webhook.Client
	searches []Search
)

type Search struct {
	SearchURL string
	Filters   cvc.Filter
}

func init() {
	godotenv.Load()

	var err error
	hook, err = webhook.NewWithURL(os.Getenv("DISCORD_WEBHOOK_URL"))
	if err != nil {
		log.Panic(err)
	}

	searches, err = getSearches()
	if err != nil {
		log.Panic(err)
	}
}

func main() {
	godotenv.Load()

	for _, search := range searches {
		go watch(search)
	}

	select {}
}

func watch(search Search) {
	query := search.Filters.Subject
	if query == "" {
		query = search.Filters.Query
	}

	log := log.New(os.Stdout, fmt.Sprintf("- [%s] - ", query), log.Ltime|log.Lmsgprefix)

	c := cvc.NewClient()

	log.Println("Performing initial search...")
	last, err := c.Search(search.Filters)
	if err != nil {
		log.Panic(err)
	}
	log.Println("Initial search complete")

	interval := time.Minute * 5

	for {
		log.Printf("Searching for courses...")
		courses, err := c.Search(search.Filters)
		if err != nil {
			log.Println(err)
			continue
		}

		for _, course := range courses {
			if !hasCourse(last, course) {
				_, err := hook.CreateEmbeds([]discord.Embed{embed(course, search.SearchURL)})
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
	t := time.Now()
	return discord.Embed{
		Title: fmt.Sprintf("%s - %s", course.College, course.Name),
		URL:   course.Link,
		Author: &discord.EmbedAuthor{
			Name:    "CVC Exchange",
			URL:     searchURL,
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

func getSearches() ([]Search, error) {
	name := "SEARCH_URL"

	searchUrls := os.Getenv(name)

	if searchUrls == "" {
		return nil, fmt.Errorf("%s is empty", name)
	}

	var filters []Search
	for _, searchUrl := range strings.Split(searchUrls, ",") {
		u, err := url.Parse(searchUrl)
		if err != nil {
			return nil, err
		}

		var filter cvc.Filter
		if err := qs.UnmarshalValues(&filter, u.Query()); err != nil {
			return nil, err
		}

		filters = append(filters, Search{
			SearchURL: searchUrl,
			Filters:   filter,
		})
	}

	return filters, nil
}
