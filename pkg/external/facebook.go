package external

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/bitesinbyte/ferret/pkg/config"
	"io"
	"net/http"
	"os"
)

const FacebookCreatePostUrl = "https://graph.facebook.com/v19.0/%s/feed"

type Facebook struct {
}

type facebookCreatePost struct {
	Message     string `json:"message"`
	Link        string `json:"link"`
	AccessToken string `json:"access_token"`
}

func (m Facebook) Post(configData config.Config, post Post) error {
	var accessToken = os.Getenv("FACEBOOK_ACCESS_TOKEN")
	var url = fmt.Sprintf(FacebookCreatePostUrl, os.Getenv("FACEBOOK_PAGE_ID"))
	content := facebookCreatePost{
		Message:     fmt.Sprintf("Just posted a new blog \n%s\n%s", post.Title, post.HashTags),
		AccessToken: accessToken,
		Link:        post.Link,
	}
	reqBody, err := json.Marshal(content)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
