package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/bitesinbyte/ferret/pkg/config"
	"github.com/bitesinbyte/ferret/pkg/poster"
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
	feed, err := gofeed.NewParser().ParseURL(os.Getenv("RSS_FEED_URL"))
	if err != nil {
		log.Fatalf("Error parsing RSS feed: %v", err)
	}

	// Check for new posts and post to Mastodon and Twitter
	for _, item := range feed.Items {
		if item.PublishedParsed.After(configData.LastRunTime) {
			fmt.Printf("Processing %s", item.Title)

			// New post found, post to Mastodon
			post := fmt.Sprintf("Just posted a new blog \n%s \n%s", item.Title, item.Link)
			err := poster.PostToot(post)
			if err != nil {
				log.Fatalf("Error posting to Mastodon: %v", err)
			}

			// New post found, post to Twitter
			err = poster.PostTweet(post)
			if err != nil {
				log.Fatalf("Error posting to Twitter: %v", err)
			}
		}
	}

	// Update last run time
	configData.LastRunTime = time.Now()
	config.SaveConfig("config.json", configData)
	fmt.Printf("Done")
}
