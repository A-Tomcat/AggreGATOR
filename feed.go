package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/atomcat/AggreGATOR/internal/database"
	"github.com/google/uuid"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	var feed *RSSFeed
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return feed, err
	}
	req.Header.Set("User-Agent", "gator")
	res, err := client.Do(req)
	if err != nil {
		return feed, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return feed, err
	}
	if err = xml.Unmarshal(body, &feed); err != nil {
		return feed, err
	}
	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)

	for i := range feed.Channel.Item {
		feed.Channel.Item[i].Title = html.UnescapeString(feed.Channel.Item[i].Title)
		feed.Channel.Item[i].Description = html.UnescapeString(feed.Channel.Item[i].Description)
	}
	return feed, nil
}

func scrapeFeeds(s *state, user database.User) error {

	nextfeed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return err
	}
	err = s.db.MarkFeedFetched(context.Background(), nextfeed.ID)
	if err != nil {
		return err
	}
	feed, err := fetchFeed(context.Background(), nextfeed.Url)
	if err != nil {
		return err
	}
	fmt.Printf("Scanning for Feed...\n")
	for _, item := range feed.Channel.Item {
		pubTime := sql.NullTime{}
		if t, err := time.Parse(time.RFC1123Z, item.PubDate); err == nil {
			pubTime = sql.NullTime{
				Time:  t,
				Valid: true,
			}
		}
		params := database.CreatePostParams{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Title:     item.Title,
			Url:       item.Link,
			Description: sql.NullString{
				String: item.Description,
				Valid:  true},
			PublishedAt: pubTime,
			FeedID:      nextfeed.ID,
		}
		_, err := s.db.CreatePost(context.Background(), params)
		if err != nil {
			if strings.Contains(err.Error(), "unique constraint") {
				continue
			} else {
				log.Printf("Error: %v", err)
				continue
			}
		}
	}
	return nil
}
