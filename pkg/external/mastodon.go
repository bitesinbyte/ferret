package external

import (
	"fmt"
	"github.com/bitesinbyte/ferret/pkg/config"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type Mastodon struct {
}

func (m Mastodon) Post(configData config.Config, post Post) error {
	content := fmt.Sprintf("Just posted a new blog \n%s \n%s\n%s", post.Title, post.Link, post.HashTags)
	var instanceURL = os.Getenv("MASTODON_INSTANCE_URL")
	var accessToken = os.Getenv("MASTODON_ACCESS_TOKEN")

	apiUrl := fmt.Sprintf("%s/api/v1/statuses?access_token=%s", instanceURL, accessToken)
	data := url.Values{}
	data.Set("status", content)

	req, err := http.NewRequest(http.MethodPost, apiUrl, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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
