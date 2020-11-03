package database

import (
	"context"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

//Cloud Firestoreの初期化
func FirebaseInit(ctx context.Context) (*firestore.Client, error) {
	sa:=option.WithCredentialsFile("path/to/serviceAccount.json")
	app,err:=firebase.NewApp(ctx,nil,sa)
	if err != nil {
		return nil, err
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		return nil, err
	}

	return client, nil
}
