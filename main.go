package main

import (
	"os"

	"github.com/gempir/go-twitch-irc/v4"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

type config struct {
	twitchUsername string
	twitchOauth    string
	commandPrefix  string
}

type application struct {
	TwitchClient *twitch.Client
	Log          *zap.SugaredLogger
	Environment  string
	Config       config
	UserMsgStore map[string][]ollamaMessage
	MsgStore     []ollamaMessage
}

func main() {
	var cfg config

	logger := zap.NewExample()
	defer func() {
		if err := logger.Sync(); err != nil {
			logger.Sugar().Errorw("error syncing logger",
				"error", err,
			)
		}
	}()
	sugar := logger.Sugar()

	err := godotenv.Load()
	if err != nil {
		sugar.Fatal("Error loading .env")
	}

	// Twitch config variables
	cfg.twitchUsername = os.Getenv("TWITCH_USERNAME")
	cfg.twitchOauth = os.Getenv("TWITCH_OAUTH")
	tc := twitch.NewClient(cfg.twitchUsername, cfg.twitchOauth)

	userMsgStore := make(map[string][]ollamaMessage)
	app := &application{
		TwitchClient: tc,
		Log:          sugar,
		Config:       cfg,
		UserMsgStore: userMsgStore,
	}

	// Received a PrivateMessage (normal chat message).
	app.TwitchClient.OnPrivateMessage(func(message twitch.PrivateMessage) {
		// roomId is the Twitch UserID of the channel the message originated from.
		// If there is no roomId something went really wrong.
		roomId := message.Tags["room-id"]
		if roomId == "" {
			return
		}

		// Message was shorter than our prefix is therefore it's irrelevant for us.
		if len(message.Message) >= 2 {
			// Check if the first 2 characters of the mesage were our prefix.
			// if they were forward the message to the command handler.
			if message.Message[:2] == "()" {
				go app.handleCommand(message)
				return
			}
		}
	})

	app.TwitchClient.OnConnect(func() {
		app.TwitchClient.Say("nouryxd", "MrDestructoid")
		app.TwitchClient.Say("nourybot", "MrDestructoid")

		// Successfully connected to Twitch
		app.Log.Infow("Successfully connected to Twitch Servers",
			"Bot username", cfg.twitchUsername,
		)
	})

	app.TwitchClient.Join("nouryxd")
	app.TwitchClient.Join("nourybot")

	// Actually connect to chat.
	err = app.TwitchClient.Connect()
	if err != nil {
		panic(err)
	}
}