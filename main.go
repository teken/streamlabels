package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/nicklaw5/helix/v2"
	"github.com/spf13/viper"
)

func main() {
	subscribeToNewestFollower := false
	flag.BoolVar(&subscribeToNewestFollower, "newest-follower", false, "subscribe to newest follower")
	//flag.BoolVar(&subscribeToNewestFollower, "f", false, "subscribe to newest follower")

	subscribeToNewestSubscriber := false
	flag.BoolVar(&subscribeToNewestSubscriber, "newest-subscriber", false, "subscribe to newest subscriber")
	//flag.BoolVar(&subscribeToNewestSubscriber, "s", false, "subscribe to newest subscriber")

	subscribeToBitsLeaderboard := false
	flag.BoolVar(&subscribeToBitsLeaderboard, "bits-leaderboard", false, "subscribe to bits leaderboard")
	//flag.BoolVar(&subscribeToBitsLeaderboard, "b", false, "subscribe to bits leaderboard")

	refreshInterval := 1 * time.Second
	flag.DurationVar(&refreshInterval, "refresh-interval", 1*time.Second, "refresh interval")
	//flag.DurationVar(&refreshInterval, "r", 1*time.Second, "refresh interval")

	outputPath := ""
	flag.StringVar(&outputPath, "output", "", "output directory")
	//flag.StringVar(&outputPath, "o", "", "output directory")

	help := false
	//flag.BoolVar(&help, "h", false, "show help")
	flag.BoolVar(&help, "help", false, "show help")

	flag.Parse()

	if help {
		flag.PrintDefaults()
		os.Exit(0)
	}

	if !subscribeToNewestFollower && !subscribeToNewestSubscriber && !subscribeToBitsLeaderboard {
		slog.Error("No subscription requested")
		flag.PrintDefaults()
		os.Exit(0)
	}

	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath("/etc/streamlabels")
	viper.AddConfigPath("$HOME/.config/streamlabels")
	viper.AddConfigPath(".")

	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		slog.Error("Fatal error config file", "error", err)
		os.Exit(1)
	}

	slog.Info("Config file was read successfully.")

	client, err := helix.NewClient(&helix.Options{
		ClientID:     viper.GetString("client_id"),
		ClientSecret: viper.GetString("client_secret"),
		RedirectURI:  "http://localhost:6789/callback",
	})
	if err != nil {
		slog.Error("Error creating helix client", "error", err)
		os.Exit(1)
	}

	authURL := client.GetAuthorizationURL(&helix.AuthorizationURLParams{
		ResponseType: "code",
		Scopes:       []string{"moderator:read:followers", "channel:read:subscriptions"},
	})

	tokenChan := make(chan string, 1)

	server := http.Server{
		Addr: ":6789",
	}

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		tokenChan <- r.URL.Query().Get("code")
		w.WriteHeader(http.StatusOK)
	})
	go server.ListenAndServe()

	fmt.Println("Authorization URL: ", authURL)

	code := <-tokenChan

	if code == "" {
		slog.Error("No code received")
		os.Exit(1)
	}

	server.Shutdown(context.Background())

	resp, err := client.RequestUserAccessToken(code)
	if err != nil {
		slog.Error("Error requesting user access token", "error", err)
		os.Exit(1)
	}

	if resp.StatusCode != 200 {
		slog.Error("Error requesting user access token", "error", resp.ErrorMessage)
		os.Exit(1)
	}

	client.SetUserAccessToken(resp.Data.AccessToken)

	info, err := client.GetUsers(&helix.UsersParams{
		Logins: []string{viper.GetString("login")},
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

	if subscribeToNewestFollower {
		go basicRunner(wg, client, refreshInterval, newestFollower, filepath.Join(outputPath, "newest_followers.txt"))
	}
	if subscribeToNewestSubscriber {
		go basicRunner(wg, client, refreshInterval, newestSubscriber, filepath.Join(outputPath, "newest_subscriber.txt"))
	}
	if subscribeToBitsLeaderboard {
		go basicRunner(wg, client, refreshInterval, bitsLeaderboard, filepath.Join(outputPath, "bits_leaderboard.txt"))
	}

	time.Sleep(refreshInterval)

	wg.Wait()
	slog.Info("All runners finished")
}

var broadcasterID string

func basicRunner(wg *sync.WaitGroup, client *helix.Client, duration time.Duration, runner func(*helix.Client) (string, error), fileName string) {
	wg.Add(1)
	ticker := time.NewTicker(duration)

	for {
		<-ticker.C
		text, err := runner(client)
		if err != nil {
			slog.Error("Error getting text for file", "error", err, "fileName", fileName)
			continue
		}
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

	return "", errors.New("no new follower found")
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

	return "", errors.New("no new subscriber found")
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
