package external

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/bitesinbyte/ferret/pkg/config"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Twitter struct {
}

func (m Twitter) Post(configData config.Config, post Post) error {
	content := fmt.Sprintf("Just posted a new blog \n%s \n%s\n%s", post.Title, post.Link, post.HashTags)
	var consumerKey = os.Getenv("TWITTER_CONSUMER_KEY")
	var consumerSecret = os.Getenv("TWITTER_CONSUMER_SECRET")
	var accessToken = os.Getenv("TWITTER_ACCESS_TOKEN")
	var accessTokenSecret = os.Getenv("TWITTER_ACCESS_TOKEN_SECRET")

	// Twitter API endpoint
	apiUrl := "https://api.twitter.com/2/tweets"

	// HTTP method
	method := http.MethodPost
	body, err := json.Marshal(map[string]string{"text": content})
	if err != nil {
		return err
	}

	// Construct and send the HTTP request with OAuth1 signature
	req, err := http.NewRequest(http.MethodPost, apiUrl, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	// Add OAuth1 Authorization header
	authHeader := buildOAuth1Header(apiUrl, method, consumerKey, consumerSecret, accessToken, accessTokenSecret)

	req.Header.Set("Authorization", authHeader)

	// Set Content-Type header
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending HTTP request: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(resp.Body)

	// Check the response status
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func buildOAuth1Header(path string, method string, consumerKey string, consumerSecret string, accessToken string, accessTokenSecret string) string {
	vals := url.Values{}
	vals.Add("oauth_nonce", generateNonce())
	vals.Add("oauth_consumer_key", consumerKey)
	vals.Add("oauth_signature_method", "HMAC-SHA1")
	vals.Add("oauth_timestamp", strconv.Itoa(int(time.Now().Unix())))
	vals.Add("oauth_token", accessToken)
	vals.Add("oauth_version", "1.0")

	// net/url package QueryEscape escapes " " into "+", this replaces it with the percentage encoding of " "
	parameterString := strings.Replace(vals.Encode(), "+", "%20", -1)

	// Calculating Signature Base String and Signing Key
	signatureBase := strings.ToUpper(method) + "&" + url.QueryEscape(strings.Split(path, "?")[0]) + "&" + url.QueryEscape(parameterString)
	signingKey := url.QueryEscape(consumerSecret) + "&" + url.QueryEscape(accessTokenSecret)
	signature := calculateSignature(signatureBase, signingKey)

	return "OAuth oauth_consumer_key=\"" + url.QueryEscape(vals.Get("oauth_consumer_key")) + "\", oauth_nonce=\"" + url.QueryEscape(vals.Get("oauth_nonce")) +
		"\", oauth_signature=\"" + url.QueryEscape(signature) + "\", oauth_signature_method=\"" + url.QueryEscape(vals.Get("oauth_signature_method")) +
		"\", oauth_timestamp=\"" + url.QueryEscape(vals.Get("oauth_timestamp")) + "\", oauth_token=\"" + url.QueryEscape(vals.Get("oauth_token")) +
		"\", oauth_version=\"" + url.QueryEscape(vals.Get("oauth_version")) + "\""
}

func calculateSignature(base, key string) string {
	hash := hmac.New(sha1.New, []byte(key))
	hash.Write([]byte(base))
	signature := hash.Sum(nil)
	return base64.StdEncoding.EncodeToString(signature)
}

func generateNonce() string {
	const allowed = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 48)
	for i := range b {
		b[i] = allowed[rand.Intn(len(allowed))]
	}
	return string(b)
}
