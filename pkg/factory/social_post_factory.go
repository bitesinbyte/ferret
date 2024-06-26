package factory

import (
	"fmt"
	"strings"

	"github.com/bitesinbyte/ferret/pkg/external"
)

const (
	LINKEDIN = "linkedin"
	MASTODON = "mastodon"
	TWITTER  = "twitter"
	FACEBOOK = "facebook"
	THREAD   = "thread"
)

func CreateSocialPoster(socialSite string) external.Poster {
	var lowerCaseSocialSite = strings.ToLower(socialSite)

	if lowerCaseSocialSite == LINKEDIN {
		return external.Linkedin{}
	} else if lowerCaseSocialSite == MASTODON {
		return external.Mastodon{}
	} else if lowerCaseSocialSite == TWITTER {
		return external.Twitter{}
	} else if lowerCaseSocialSite == FACEBOOK {
		return external.Facebook{}
	} else if lowerCaseSocialSite == THREAD {
		return external.Thread{}
	} else {
		panic(fmt.Sprintf("%s is not supported", lowerCaseSocialSite))
	}
}
