package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/nicklaw5/helix/v2"
	"github.com/zalando/go-keyring"
)

const twitchScopes = "moderator:read:followers channel:read:subscriptions"
const twitchClientID = "guf16ba4jxcwzh5zuxdh03uhmap8lr"

func auth(client *helix.Client) (helix.AccessCredentials, error) {
	device, err := client.GetDeviceCode(twitchScopes)
	if err != nil {
		slog.Error("Error getting device code", "error", err)
		os.Exit(1)
	}

	if device.StatusCode != 200 {
		slog.Error("Error getting device code", "error", device.ErrorMessage)
		os.Exit(1)
	}

	now := time.Now()
	expiresAt := now.Add(time.Duration(device.Data.ExpiresIn) * time.Second)

	fmt.Println("Code: ", device.Data.UserCode)
	fmt.Println("Authorization URL: ", device.Data.VerificationURI)
	fmt.Println("Expires in: ", time.Duration(device.Data.ExpiresIn)*time.Second, " seconds")

	ticker := time.NewTicker(time.Second)

	for {
		<-ticker.C
		if time.Now().After(expiresAt) {
			slog.Error("Device code expired")
			os.Exit(1)
		}

		resp, err := client.RequestUserAccessTokenWithDeviceCode(device.Data.DeviceCode, twitchScopes)
		if err != nil {
			slog.Error("Error requesting user access token", "error", err)
			os.Exit(1)
		}

		if resp.StatusCode == 400 && resp.ErrorMessage == "authorization_pending" {
			continue
		}

		if resp.StatusCode != 200 {
			slog.Error("Error requesting user access token", "error", resp.ErrorMessage)
			os.Exit(1)
		}

		return resp.Data, nil
	}
}

func main() {
	var f flag.FlagSet

	subscribeToNewestFollower := false
	f.BoolVar(&subscribeToNewestFollower, "newest-follower", false, "subscribe to newest follower")
	//f.BoolVar(&subscribeToNewestFollower, "f", false, "subscribe to newest follower")

	subscribeToNewestSubscriber := false
	f.BoolVar(&subscribeToNewestSubscriber, "newest-subscriber", false, "subscribe to newest subscriber")
	//f.BoolVar(&subscribeToNewestSubscriber, "s", false, "subscribe to newest subscriber")

	subscribeToBitsLeaderboard := false
	f.BoolVar(&subscribeToBitsLeaderboard, "bits-leaderboard", false, "subscribe to bits leaderboard")
	//f.BoolVar(&subscribeToBitsLeaderboard, "b", false, "subscribe to bits leaderboard")

	refreshInterval := 1 * time.Second
	f.DurationVar(&refreshInterval, "refresh-interval", 1*time.Second, "refresh interval")
	//f.DurationVar(&refreshInterval, "r", 1*time.Second, "refresh interval")

	logout := false
	f.BoolVar(&logout, "logout", false, "logout")
	//f.BoolVar(&logout, "l", false, "logout")

	outputPath := ""
	f.StringVar(&outputPath, "output", "", "output directory")
	//f.StringVar(&outputPath, "o", "", "output directory")

	help := false
	f.BoolVar(&help, "help", false, "show help")
	//f.BoolVar(&help, "h", false, "show help")

	if len(os.Args) < 2 {
		fmt.Println("Usage: streamlabels <channel_name> [flags]")
		f.PrintDefaults()
		os.Exit(0)
	}

	channelName := os.Args[1]

	f.Parse(os.Args[2:])

	if help {
		fmt.Println("Usage: streamlabels <channel_name> [flags]")
		f.PrintDefaults()
		os.Exit(0)
	}

	if logout {
		keyring.Delete("streamlabels", "access_token")
		keyring.Delete("streamlabels", "refresh_token")
		os.Exit(0)
	}

	if !subscribeToNewestFollower && !subscribeToNewestSubscriber && !subscribeToBitsLeaderboard {
		slog.Error("No subscription requested")
		flag.PrintDefaults()
		os.Exit(0)
	}

	client, err := helix.NewClient(&helix.Options{
		ClientID: twitchClientID,
	})
	if err != nil {
		slog.Error("Error creating helix client", "error", err)
		os.Exit(1)
	}

	accessToken, err := keyring.Get("streamlabels", "access_token")
	if err != nil && err != keyring.ErrNotFound {
		slog.Error("Error getting access token", "error", err)
		os.Exit(1)
	}

	refreshToken, err := keyring.Get("streamlabels", "refresh_token")
	if err != nil && err != keyring.ErrNotFound {
		slog.Error("Error getting refresh token", "error", err)
		os.Exit(1)
	}

	var expiresAt time.Time
	var accessCredentials helix.AccessCredentials
	if accessToken != "" && refreshToken != "" {
		client.SetUserAccessToken(accessToken)
		client.SetRefreshToken(refreshToken)

		resp, err := client.RefreshUserAccessToken(refreshToken)
		if err != nil {
			slog.Error("Error refreshing user access token", "error", err)
			os.Exit(1)
		}

		if resp.StatusCode != 200 {
			slog.Error("Error refreshing user access token", "error", resp.ErrorMessage)
			os.Exit(1)
		}

		accessCredentials = resp.Data
		client.SetUserAccessToken(accessCredentials.AccessToken)
		client.SetRefreshToken(accessCredentials.RefreshToken)
		keyring.Set("streamlabels", "access_token", accessCredentials.AccessToken)
		keyring.Set("streamlabels", "refresh_token", accessCredentials.RefreshToken)
		expiresAt = time.Now().Add(time.Duration(accessCredentials.ExpiresIn) * time.Second)
	} else {
		accessCredentials, err = auth(client)
		if err != nil {
			slog.Error("Error authenticating", "error", err)
			os.Exit(1)
		}

		expiresAt = time.Now().Add(time.Duration(accessCredentials.ExpiresIn) * time.Second)

		client.SetUserAccessToken(accessCredentials.AccessToken)
		client.SetRefreshToken(accessCredentials.RefreshToken)

		keyring.Set("streamlabels", "access_token", accessCredentials.AccessToken)
		keyring.Set("streamlabels", "refresh_token", accessCredentials.RefreshToken)
	}

	slog.Info("User access token and refresh token set")

	info, err := client.GetUsers(&helix.UsersParams{
		Logins: []string{channelName},
	})
	if err != nil {
		slog.Error("Error getting channel information", "error", err)
		os.Exit(1)
	}

	if info.StatusCode != 200 {
		slog.Error("Error getting channel information", "error", info.ErrorMessage)
		os.Exit(1)
	}

	broadcasterID = info.Data.Users[0].ID

	wg := &sync.WaitGroup{}

	currentValues := map[string]string{
		"newest_followers.txt":  "",
		"newest_subscriber.txt": "",
		"bits_leaderboard.txt":  "",
	}

	go func() {
		for {
			time.Sleep(time.Second)
			if time.Now().After(expiresAt.Add(time.Minute * -10)) {
				slog.Info("Access token expired, refreshing")

				resp, err := client.RefreshUserAccessToken(accessCredentials.RefreshToken)
				if err != nil {
					slog.Error("Error refreshing user access token", "error", err)
					os.Exit(1)
				}

				if resp.StatusCode != 200 {
					slog.Error("Error refreshing user access token", "error", resp.ErrorMessage)
					os.Exit(1)
				}

				accessCredentials = resp.Data
				client.SetUserAccessToken(accessCredentials.AccessToken)
				client.SetRefreshToken(accessCredentials.RefreshToken)
				keyring.Set("streamlabels", "access_token", accessCredentials.AccessToken)
				keyring.Set("streamlabels", "refresh_token", accessCredentials.RefreshToken)
				expiresAt = time.Now().Add(time.Duration(accessCredentials.ExpiresIn) * time.Second)
				slog.Info("Access token refreshed")
			}
		}
	}()

	if subscribeToNewestFollower {
		go basicRunner(wg, client, refreshInterval, newestFollower, filepath.Join(outputPath, "newest_followers.txt"), &currentValues)
	}
	if subscribeToNewestSubscriber {
		go basicRunner(wg, client, refreshInterval, newestSubscriber, filepath.Join(outputPath, "newest_subscriber.txt"), &currentValues)
	}
	if subscribeToBitsLeaderboard {
		go basicRunner(wg, client, refreshInterval, bitsLeaderboard, filepath.Join(outputPath, "bits_leaderboard.txt"), &currentValues)
	}

	time.Sleep(refreshInterval)

	wg.Wait()
	slog.Info("All runners finished")
}

var broadcasterID string

func basicRunner(wg *sync.WaitGroup, client *helix.Client, duration time.Duration, runner func(*helix.Client) (string, error), fileName string, currentValues *map[string]string) {
	wg.Add(1)
	ticker := time.NewTicker(duration)

	for {
		<-ticker.C
		text, err := runner(client)
		if err != nil {
			slog.Error("Error getting text for file", "error", err, "fileName", fileName)
			continue
		}
		if _, ok := (*currentValues)[fileName]; ok {
			if (*currentValues)[fileName] != text {
				slog.Info("File has changed", "fileName", fileName, "value", text)
			}
		}
		(*currentValues)[fileName] = text
		err = os.WriteFile(fileName, []byte(text), 0644)
		if err != nil {
			slog.Error("Error writing file", "error", err, "fileName", fileName)
			continue
		}

	}
}

func newestFollower(client *helix.Client) (string, error) {
	resp, err := client.GetChannelFollows(&helix.GetChannelFollowsParams{
		BroadcasterID: broadcasterID,
		First:         1,
	})
	if err != nil {
		slog.Error("Error getting channel follows", "error", err)
		return "", err
	}

	for _, follow := range resp.Data.Channels {
		return follow.Username, nil
	}

	return "", nil
}

func newestSubscriber(client *helix.Client) (string, error) {
	resp, err := client.GetSubscriptions(&helix.SubscriptionsParams{
		BroadcasterID: broadcasterID,
		First:         1,
	})
	if err != nil {
		slog.Error("Error getting subscriptions", "error", err)
		return "", err
	}

	for _, sub := range resp.Data.Subscriptions {
		return sub.UserName, nil
	}

	return "", nil
}

func bitsLeaderboard(client *helix.Client) (string, error) {
	resp, err := client.GetBitsLeaderboard(&helix.BitsLeaderboardParams{
		UserID: broadcasterID,
		Period: "all",
		Count:  10,
	})
	if err != nil {
		slog.Error("Error getting bits leaderboard", "error", err)
		return "", err
	}

	var sb strings.Builder
	for _, user := range resp.Data.UserBitTotals {
		sb.WriteString(fmt.Sprintf("%s: %d\n", user.UserName, user.Score))
	}
	return sb.String(), nil
}
