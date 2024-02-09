package poster

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func PostToot(content string) error {

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
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
