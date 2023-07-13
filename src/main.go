package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type DBSecret struct {
	Username string
	Password string
}

var REGION string
var SECRET_NAME string
var DATABASE_HOST string
var DATABASE_PORT string
var DATABASE_SCHEMA string
var SAVED_DB *sql.DB

func main() {
	getEnv()

	http.HandleFunc("/", fn)
	http.ListenAndServe(":80", nil)
}


func getEnv() {
	REGION = os.Getenv("REGION")
	SECRET_NAME = os.Getenv("SECRET_NAME")
	DATABASE_HOST = os.Getenv("DATABASE_HOST")
	DATABASE_PORT = os.Getenv("DATABASE_PORT")
	DATABASE_SCHEMA = os.Getenv("DATABASE_SCHEMA")
}

func fn(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
		case "GET":
			db := getDB()
			fmt.Fprint(w, db.Stats())
			break
	}
}

func getDBSecret() DBSecret {
	config, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(REGION))
	if err != nil {
		log.Fatal(err)
	}

	svc := secretsmanager.NewFromConfig(config)

	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(SECRET_NAME),
		VersionStage: aws.String("AWSCURRENT"),
	}

	result, err := svc.GetSecretValue(context.TODO(), input)
	if err != nil {
		log.Fatal(err.Error())
	}

	var secret DBSecret;
	json.Unmarshal([]byte(*result.SecretString), &secret)

	return secret
}

func getDB() *sql.DB {
	if (SAVED_DB == nil || SAVED_DB.Ping() != nil) {
		secret := getDBSecret()

		db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
			secret.Username,
			secret.Password,
			DATABASE_HOST,
			DATABASE_PORT,
			DATABASE_SCHEMA,
		))
		
		if err != nil {
			panic(err)
		}

		SAVED_DB = db
	}

	return SAVED_DB;
}
