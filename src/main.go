package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"

	_ "github.com/go-sql-driver/mysql"
)

type DBSecret struct {
	Username string
	Password string
}

var REGION *string
var SECRET_NAME *string
var DATABASE_HOST *string
var DATABASE_SCHEMA *string
var SAVED_DB *sql.DB

func main() {
	getEnv()

	http.HandleFunc("/", fn)
	http.ListenAndServe(":80", nil)
}


func getEnv() {
	REGION = flag.String("region", "ap-northeast-2", "Secrets manager region")
	SECRET_NAME = flag.String("secret", "", "Secret Name")
	DATABASE_HOST = flag.String("db_host", "", "Database Host URI")
	DATABASE_SCHEMA = flag.String("db_schema", "", "Database Schema")

	flag.Parse()
}

func fn(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
		case "GET":
			db := getDB()
			var version string

			if err := db.QueryRow("SELECT VERSION()",).Scan(&version); err != nil {
				fmt.Fprint(w, err.Error())
    	}
			
			fmt.Fprint(w, version)

			break
	}
}

func getDBSecret() DBSecret {
	config, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(*REGION))
	if err != nil {
		log.Fatal(err)
	}

	svc := secretsmanager.NewFromConfig(config)

	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(*SECRET_NAME),
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

		db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s",
			secret.Username,
			secret.Password,
			*DATABASE_HOST,
			*DATABASE_SCHEMA,
		))
		
		if err != nil {
			panic(err)
		}

		SAVED_DB = db
	}

	return SAVED_DB;
}
