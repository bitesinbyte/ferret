package main

import (
	"fmt"
	"github.com/bitesinbyte/ferret/pkg/external"
	"github.com/bitesinbyte/ferret/pkg/factory"
	"log"
	"time"

	"github.com/bitesinbyte/ferret/pkg/config"
	"github.com/joho/godotenv"
	"github.com/mmcdole/gofeed"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(err)
	}

	// Load configuration from JSON file
	configData := config.LoadConfig("config.json")

	// Parse RSS feed
	feed, err := gofeed.NewParser().ParseURL(configData.BaseUrl + configData.FeedEndpoint)
	if err != nil {
		log.Fatalf("Error parsing RSS feed: %v", err)
	}

	// Check for new posts and post to Mastodon and Twitter
	for _, item := range feed.Items {
		if !item.PublishedParsed.After(configData.LastRunTime) {
			fmt.Printf("Processing %s", item.Title)

			// Create Hashtags
			hashTags := ""
			for _, category := range item.Categories {
				hashTags += fmt.Sprintf("%s%s ", "#", category)
			}

			for _, social := range configData.Socials {
				socialClient := factory.CreateSocialPoster(social)
				post := external.Post{
					Title:       item.Title,
					Link:        item.Link,
					HashTags:    hashTags,
					Description: item.Description,
				}
				err := socialClient.Post(configData, post)
				if err != nil {
					log.Fatalf("Error posting to %s: %v", social, err)
				}
			}
		}
	}

	// Update last run time
	configData.LastRunTime = time.Now()
	config.SaveConfig("config.json", configData)
	fmt.Printf("Done")
}
