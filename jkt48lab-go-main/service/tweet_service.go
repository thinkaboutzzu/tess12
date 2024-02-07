package service

import (
	"context"
	"github.com/ChimeraCoder/anaconda"
	"github.com/dghubble/oauth1"
	"github.com/g8rswimmer/go-twitter/v2"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
)

type TweetService interface {
	GetClient(ctx context.Context) *twitter.Client
	GetMediaClient(ctx context.Context) *anaconda.TwitterApi
	UploadMedia(ctx context.Context, base64string string) string
	Post(ctx context.Context, content string, base64string string)
}

type TweetServiceImpl struct {
}

type authorization struct {
	Token string
}

func (a authorization) Add(req *http.Request) {
}

var CLIENT *twitter.Client
var MEDIACLIENT *anaconda.TwitterApi

func (t *TweetServiceImpl) GetClient(ctx context.Context) *twitter.Client {
	if CLIENT != nil {
		return CLIENT
	}
	godotenv.Load()
	config := oauth1.NewConfig(os.Getenv("TWITTER_CONSUMER_KEY"), os.Getenv("TWITTER_CONSUMER_KEY_SECRET"))
	httpClient := config.Client(oauth1.NoContext, &oauth1.Token{
		Token:       os.Getenv("TWITTER_ACCESS_TOKEN"),
		TokenSecret: os.Getenv("TWITTER_ACCESS_TOKEN_SECRET"),
	})
	client := &twitter.Client{
		Authorizer: &authorization{},
		Client:     httpClient,
		Host:       "https://api.twitter.com",
	}
	CLIENT = client
	return client
}

func (t *TweetServiceImpl) GetMediaClient(ctx context.Context) *anaconda.TwitterApi {
	if MEDIACLIENT != nil {
		return MEDIACLIENT
	}
	godotenv.Load()
	client := anaconda.NewTwitterApiWithCredentials(
		os.Getenv("TWITTER_ACCESS_TOKEN"),
		os.Getenv("TWITTER_ACCESS_TOKEN_SECRET"),
		os.Getenv("TWITTER_CONSUMER_KEY"),
		os.Getenv("TWITTER_CONSUMER_KEY_SECRET"),
	)
	MEDIACLIENT = client
	return client
}

func (t *TweetServiceImpl) UploadMedia(ctx context.Context, base64string string) string {
	mediaClient := t.GetMediaClient(ctx)
	defer mediaClient.Close()

	media, err := mediaClient.UploadMedia(base64string)
	if err != nil {
		log.Fatal(err)
	}
	return media.MediaIDString
}

func (t *TweetServiceImpl) Post(ctx context.Context, content string, base64string string) {
	client := t.GetClient(ctx)

	mediaId := t.UploadMedia(ctx, base64string)
	req := twitter.CreateTweetRequest{
		Text: content,
		Media: &twitter.CreateTweetMedia{
			IDs: []string{mediaId},
		},
	}
	tweet, err := client.CreateTweet(ctx, req)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(tweet.Tweet.Text)
}

func NewTweetService() TweetService {
	return &TweetServiceImpl{}
}
