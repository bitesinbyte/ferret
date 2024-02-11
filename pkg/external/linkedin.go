package external

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/bitesinbyte/ferret/pkg/config"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
)

const (
	LinkedinCreatePostUrl = "https://api.linkedin.com/v2/posts"
	LinkedinProfileUrl    = "https://api.linkedin.com/v2/userinfo"
	LinkedinImageUrl      = "https://api.linkedin.com/rest/images?action=initializeUpload"
)

type Linkedin struct {
}

func (m Linkedin) Post(configData config.Config, post Post) error {
	var accessToken = fmt.Sprintf("Bearer %s", os.Getenv("LINKEDIN_ACCESS_TOKEN"))
	err, userInfo := fetchProfile(accessToken)
	if err != nil {
		return err
	}

	content := fmt.Sprintf("Just posted a new blog\n\n%s", post.HashTags)
	err = createPost(configData, post, content, userInfo.Sub, accessToken)
	return err
}
func createPost(configData config.Config, post Post, content string, authorId string, accessToken string) error {
	err, thumbnail := getThumbnail(configData, post.Link, authorId, accessToken)
	if err != nil {
		return err
	}

	body, err := json.Marshal(linkedinPost{
		Author:             fmt.Sprintf("urn:li:person:%s", authorId),
		Commentary:         content,
		ContentLandingPage: post.Link,
		Visibility:         "PUBLIC",
		Distribution: struct {
			FeedDistribution string `json:"feedDistribution"`
		}{
			FeedDistribution: "MAIN_FEED",
		},
		Content: struct {
			Article postArticleContent `json:"article"`
		}{
			Article: postArticleContent{
				Source:      post.Link,
				Thumbnail:   thumbnail,
				Title:       post.Title,
				Description: post.Description,
			},
		},
		LifecycleState:           "PUBLISHED",
		ContentCallToActionLabel: "SEE_MORE",
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, LinkedinCreatePostUrl, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", accessToken)
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
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

type postArticleContent struct {
	Source      string `json:"source"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Thumbnail   string `json:"thumbnail"`
}
type linkedinPost struct {
	Author             string `json:"author"`
	Commentary         string `json:"commentary"`
	Visibility         string `json:"visibility"`
	ContentLandingPage string `json:"contentLandingPage"`
	Distribution       struct {
		FeedDistribution string `json:"feedDistribution"`
	} `json:"distribution"`
	Content struct {
		Article postArticleContent `json:"article"`
	} `json:"content"`
	LifecycleState            string `json:"lifecycleState"`
	IsReshareDisabledByAuthor bool   `json:"isReshareDisabledByAuthor"`
	ContentCallToActionLabel  string `json:"contentCallToActionLabel"`
}

// thumbnail
type initializeUploadResponse struct {
	Value struct {
		UploadUrlExpiresAt int64  `json:"uploadUrlExpiresAt"`
		UploadUrl          string `json:"uploadUrl"`
		Image              string `json:"image"`
	} `json:"value"`
}
type initializeUploadRequest struct {
	InitializeUploadRequest struct {
		Owner string `json:"owner"`
	} `json:"initializeUploadRequest"`
}

func getThumbnail(configData config.Config, articleUrl string, userId string, accessToken string) (error, string) {
	err, initializeUpload := initializeUpload(accessToken, userId)
	if err != nil {
		return err, ""
	}
	err = uploadImage(articleUrl, configData, initializeUpload)
	if err != nil {
		return err, ""
	}
	return nil, initializeUpload.Value.Image
}

func uploadImage(articleUrl string, configData config.Config, initializeUpload *initializeUploadResponse) error {
	imageUrl, err := getOGImageURL(articleUrl)
	if err != nil {
		return err
	}

	if configData.DoesMetaOgHasRelativePath {
		imageUrl = configData.BaseUrl + imageUrl
	}

	resp, err := http.Get(imageUrl)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(resp.Body)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", imageUrl)
	if err != nil {
		return err
	}
	_, err = io.Copy(part, resp.Body)
	if err != nil {
		return err
	}
	err = writer.Close()
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", initializeUpload.Value.UploadUrl, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("HTTP request failed with status code %d", resp.StatusCode)
	}
	return nil
}

// GetOGImageURL retrieves the og:image URL from the HTTP header of a given URL
func getOGImageURL(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP request failed with status code %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
	}

	var ogImageURL string
	doc.Find("meta").Each(func(i int, s *goquery.Selection) {
		property, _ := s.Attr("property")
		content, _ := s.Attr("content")
		if strings.ToLower(property) == "og:image" {
			ogImageURL = content
		}
	})

	if ogImageURL == "" {
		return "", fmt.Errorf("og:image tag not found in HTML head section")
	}

	return ogImageURL, nil
}
func initializeUpload(accessToken string, userId string) (error, *initializeUploadResponse) {
	request := initializeUploadRequest{
		InitializeUploadRequest: struct {
			Owner string `json:"owner"`
		}{
			Owner: fmt.Sprintf("urn:li:person:%s", userId),
		},
	}
	body, err := json.Marshal(request)
	if err != nil {
		return err, nil
	}
	req, err := http.NewRequest(http.MethodPost, LinkedinImageUrl, bytes.NewBuffer(body))
	if err != nil {
		return err, nil
	}
	req.Header.Set("Authorization", accessToken)
	req.Header.Set("LinkedIn-Version", "202401")
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err, nil
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode), nil
	}
	var uploadResponse initializeUploadResponse
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return err, nil
	}
	if err := json.Unmarshal(body, &uploadResponse); err != nil {
		panic(err)
	}
	return nil, &uploadResponse
}

type userInfo struct {
	Sub string `json:"sub"`
}

func fetchProfile(accessToken string) (error, *userInfo) {
	req, err := http.NewRequest(http.MethodGet, LinkedinProfileUrl, nil)
	if err != nil {
		return err, nil
	}
	req.Header.Set("Authorization", accessToken)

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err, nil
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode), nil
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err, nil
	}
	var data userInfo
	err = json.Unmarshal(body, &data)
	if err != nil {
		return err, nil
	}
	return nil, &data
}
