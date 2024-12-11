package drive

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/clerkinc/clerk-sdk-go/clerk"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

func GetDriveConfig() (*oauth2.Config, error) {
	json, err := os.ReadFile("resources/drive/credentials.json")
	if err != nil {
		log.Fatalf(err.Error())
	}

	// needs config to get url
	return google.ConfigFromJSON(json, drive.DriveReadonlyScope) // other scopes possible
}

func GetDriveAuthURL(config *oauth2.Config) string {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline) // AccessTypeOffline
	return authURL
}

func GetDriveService(config *oauth2.Config, token *clerk.UserOAuthAccessToken) (*drive.Service, error) {

	oauthToken := &oauth2.Token{
		AccessToken: token.Token,
		TokenType:   "Bearer",
		Expiry:      time.Now().Add(time.Hour * 24),
	}

	ctx := context.Background()
	client := config.Client(ctx, oauthToken)

	return drive.NewService(ctx, option.WithHTTPClient(client))
}
