package external

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/bitesinbyte/ferret/pkg/config"
)

const ThreadCreatePostUrl = "https://graph.threads.net/v1.0/%s/threads?media_type=Text&text=%s&access_token=%s"
const ThreadPublishPostUrl = "https://graph.threads.net/v1.0/%s/threads_publish?creation_id=%s&access_token=%s"

type Thread struct {
}

type threadPostResponse struct {
	Id string `json:"id"`
}

func (t Thread) Post(configData config.Config, post Post) error {

	postData, err := createThreadPost(post)
	if err != nil {
		return err
	}
	return publishPost(postData)
}

func createThreadPost(post Post) (*threadPostResponse, error) {
	var postUrl = fmt.Sprintf(ThreadCreatePostUrl, os.Getenv("THREAD_USER_ID"),
		url.QueryEscape(fmt.Sprintf("Just posted a new blog \n%s \n%s\n%s", post.Title, post.Link, post.HashTags)),
		os.Getenv("THREAD_ACCESS_TOKEN"))

	req, err := http.NewRequest(http.MethodPost, postUrl, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	jsonData, err := io.ReadAll(resp.Body)

	if err != nil {
		panic(err)
	}

	var data threadPostResponse
	err = json.Unmarshal([]byte(jsonData), &data)
	if err != nil {
		panic(err)
	}

	return &data, nil
}

func publishPost(post *threadPostResponse) error {
	var postUrl = fmt.Sprintf(ThreadPublishPostUrl, os.Getenv("THREAD_USER_ID"),
		post.Id,
		os.Getenv("THREAD_ACCESS_TOKEN"))

	req, err := http.NewRequest(http.MethodPost, postUrl, nil)
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
