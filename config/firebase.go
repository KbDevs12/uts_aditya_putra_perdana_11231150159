package config

import (
	"context"

	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

var App *firebase.App

func InitFirebase() {
	opt := option.WithCredentialsFile("firebase.json")
	app, _ := firebase.NewApp(context.Background(), nil, opt)
	App = app
}
