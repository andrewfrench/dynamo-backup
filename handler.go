package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"log"
	"time"
)

type Request struct{}

type Response struct {
	TablesComplete []string `json:"completed_tables"`
	TablesErrored  []string `json:"errored_tables"`
}

func main() {
	lambda.Start(HandleRequest)
}

func HandleRequest(ctx context.Context, name Request) (Response, error) {
	successTables := []string{}
	failTables := []string{}

	// Create dynamo connection
	sess := session.Must(session.NewSession())
	dynamo := dynamodb.New(sess)

	// Collect tables
	tables, err := getTableNames(dynamo)
	if err != nil {
		log.Fatalf("Unable to get tables: %s", err.Error())
	}

	// Maximum 50 backup requests per second
	for range time.Tick(40 * time.Millisecond) {
		if len(tables) == 0 {
			break
		}

		// Grab a table from the tables slice
		table := tables[0]

		// Create a new slice that doesn't include the current table
		tables = tables[1:]

		err := backupTable(dynamo, table)
		if err != nil {
			failTables = append(failTables, *table)
		} else {
			successTables = append(successTables, *table)
		}
	}

	// Create response struct
	resp := Response{
		TablesComplete: successTables,
		TablesErrored:  failTables,
	}

	return resp, nil
}

func getTableNames(dynamo *dynamodb.DynamoDB) ([]*string, error) {
	log.Printf("Collecting table names...")

	tableNames := []*string{}
	input := &dynamodb.ListTablesInput{}

	output, err := dynamo.ListTables(input)
	if err != nil {
		return []*string{}, err
	}

	tableNames = append(tableNames, output.TableNames...)

	// Output list is capped at 100 tables -- if a LastEvaluatedTableName is returned, there are more tables to backup
	for output.LastEvaluatedTableName != nil {
		input = &dynamodb.ListTablesInput{
			ExclusiveStartTableName: output.LastEvaluatedTableName,
		}

		output, err = dynamo.ListTables(input)
		if err != nil {
			return []*string{}, err
		}

		tableNames = append(tableNames, output.TableNames...)
	}

	log.Printf("Done.")

	return tableNames, nil
}

func backupTable(dynamo *dynamodb.DynamoDB, tableName *string) error {
	backupTime := getTimeStr()
	backupName := fmt.Sprintf("%s_%s", *backupTime, *tableName)

	log.Printf("Backing up %s as %s...", *tableName, backupName)

	input := &dynamodb.CreateBackupInput{
		BackupName: &backupName,
		TableName:  tableName,
	}

	_, err := dynamo.CreateBackup(input)
	if err != nil {
		return err
	}

	log.Printf("Done.")

	return nil
}

func getTimeStr() *string {
	timeStr := time.Now().Format("2006-01-02_15-04-05")

	return &timeStr
}
