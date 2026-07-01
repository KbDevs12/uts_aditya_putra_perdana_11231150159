package config

import (
	"context"
	"log"
	"os"

	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

var App *firebase.App

func InitFirebase() {
	credentialPath := os.Getenv("FIREBASE_CREDENTIALS")
	if credentialPath == "" {
		credentialPath = "firebase.json"
	}

	app, err := firebase.NewApp(context.Background(), nil, option.WithCredentialsFile(credentialPath))
	if err != nil {
		log.Printf("Firebase disabled: failed to load credentials from %s: %v", credentialPath, err)
		App = nil
		return
	}

	App = app
	log.Printf("Firebase initialized using %s", credentialPath)
}
