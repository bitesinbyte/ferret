package factory

import (
	"fmt"
	"github.com/bitesinbyte/ferret/pkg/external"
	"strings"
)

const (
	LINKEDIN = "linkedin"
	MASTODON = "mastodon"
	TWITTER  = "twitter"
)

func CreateSocialPoster(socialSite string) external.Poster {
	var lowerCaseSocialSite = strings.ToLower(socialSite)

	if lowerCaseSocialSite == LINKEDIN {
		return external.Linkedin{}
	} else if lowerCaseSocialSite == MASTODON {
		return external.Mastodon{}
	} else if lowerCaseSocialSite == TWITTER {
		return external.Twitter{}
	} else {
		panic(fmt.Sprintf("%s is not supported", lowerCaseSocialSite))
	}
}
