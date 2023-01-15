package main

import (
	"context"
	"fmt"
	"os"

	"github.com/bitdotioinc/go-bitdotio/bitdotio"
)

func main() {
	// Setup
	token := os.Getenv("BITDOTIO_TOKEN")
	b := bitdotio.NewBitDotIO(token)

	ctx := context.Background()
	pool, err := b.CreatePool(ctx, "andrewdoss/rep-tracker")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Pool creation failed: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	var greeting string
	err = pool.QueryRow(context.Background(), "select 'Hello, world!'").Scan(&greeting)
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(greeting)

	// newDBName := "foo_db8"
	// // Demonstrate creating a database
	// newDatabase, err := b.CreateDatabase(
	// 	newDBName,
	// 	&bitdotio.DatabaseConfig{Name: newDBName, IsPrivate: true},
	// )
	// if err != nil {
	// 	fmt.Printf("main failed to create database: %v", err)
	// 	os.Exit(1)
	// }
	// fmt.Printf("Create Example: %v\n", newDatabase.Name)

	// // Demonstrate listing databases
	// databases, err := b.ListDatabases()
	// if err != nil {
	// 	fmt.Printf("main failed to list databases: %v", err)
	// 	os.Exit(1)
	// }
	// fmt.Printf("Found %d databases:\n", len(databases))
	// for _, db := range databases {
	// 	fmt.Printf("- %v\n", db.Name)
	// }

	// // Demonstrate getting a database
	// database, err := b.GetDatabase("andrewdoss", newDBName)
	// if err != nil {
	// 	fmt.Printf("failed to get database: %v", err)
	// 	os.Exit(1)
	// }
	// fmt.Printf("Get Example: %v\n", database.Name)
	// usageCurrent := database.UsageCurrent
	// fmt.Printf("Usage: %v %v %v\n", usageCurrent.RowsQueried, usageCurrent.PeriodStart, usageCurrent.PeriodEnd)

	// // Confirm update
	// updatedDBName := newDBName + "-updated"
	// database, err = b.UpdateDatabase(
	// 	"andrewdoss",
	// 	newDBName,
	// 	&bitdotio.DatabaseConfig{Name: updatedDBName, IsPrivate: true},
	// )
	// if err != nil {
	// 	fmt.Printf("failed to update database: %v", err)
	// 	os.Exit(1)
	// }
	// fmt.Printf("Update Example: %v\n", database.Name)

	// // Demonstrate deleting a database
	// err = b.DeleteDatabase("andrewdoss", updatedDBName)
	// if err != nil {
	// 	fmt.Printf("failed to delete database: %v", err)
	// 	os.Exit(1)
	// }
	// // Confirm deletion
	// databases, err = b.ListDatabases()
	// if err != nil {
	// 	fmt.Printf("failed to list databases: %v", err)
	// 	os.Exit(1)
	// }
	// fmt.Printf("Confirming delete... found %d databases\n", len(databases))

	// // Test creating a new API key
	// credentials, err := b.CreateKey()
	// if err != nil {
	// 	fmt.Printf("failed to create a personal key: %v", err)
	// 	os.Exit(1)
	// }
	// fmt.Printf("Username: %s, Key: %s\n", credentials.Username, credentials.APIKEY)

	// Test listing service accounts and getting a single service account
	// serviceAccounts, err := b.ListServiceAccounts()
	// if err != nil {
	// 	fmt.Printf("failed to list service accounts: %v", err)
	// 	os.Exit(1)
	// }
	// fmt.Printf("Found %d service accounts:\n", len(serviceAccounts))
	// var serviceAccountID string
	// for _, s := range serviceAccounts {
	// 	serviceAccountID = s.ID
	// 	fmt.Printf("- %s with role %s and created %v", s.Name, s.Role, s.DateCreated)
	// 	for _, db := range s.Databases {
	// 		fmt.Printf("    - %s\n", db.Name)
	// 	}
	// }

	// serviceAccount, err := b.GetServiceAccount(serviceAccountID)
	// if err != nil {
	// 	fmt.Printf("failed to get service account: %v", err)
	// 	os.Exit(1)
	// }
	// fmt.Printf("Service account name: %s\n", serviceAccount.Name)

	// credentials, err := b.CreateServiceAccountKey(serviceAccountID)
	// if err != nil {
	// 	fmt.Printf("failed to create service account key: %v", err)
	// 	os.Exit(1)
	// }
	// fmt.Printf("Username: %s, Key: %s\n", credentials.Username, credentials.APIKEY)

	// err = b.RevokeServiceAccountKeys(serviceAccountID)
	// if err != nil {
	// 	fmt.Printf("failed to create service account key: %v", err)
	// 	os.Exit(1)
	// }

	// // Test non-ok response handling
	// bad_auth_b := bitdotio.NewBitDotIO("fake-token")
	// _, err = bad_auth_b.ListServiceAccounts()
	// if err == nil {
	// 	fmt.Printf("Expected an error response")
	// 	os.Exit(1)
	// }
	// fmt.Printf("Got expected error: %v\n", err)

	// Test creating an import job
	// f, err := os.Open("iris.csv")
	// if err != nil {
	// 	fmt.Printf("failed to open file: %v\n", err)
	// 	os.Exit(1)
	// }
	// defer f.Close()
	// importJob, err := b.CreateImportJob("andrewdoss/import", "iris_test", &bitdotio.ImportJobConfig{File: f})
	// if err != nil {
	// 	fmt.Printf("failed to create new import job: %v\n", err)
	// 	os.Exit(1)
	// }
	// fmt.Printf("Import job ID %s and status url %s.\n", importJob.ID, importJob.StatusURL)

	// // Test retrieving status for an import job
	// importJob, err = b.GetImportJob(importJob.ID)
	// if err != nil {
	// 	fmt.Printf("failed to get import job status: %v", err)
	// 	os.Exit(1)
	// }
	// fmt.Printf("Import job ID %s and status url %s.\n", importJob.ID, importJob.StatusURL)

	// // Test creating an export job
	// exportJob, err := b.CreateExportJob("andrewdoss/import", &bitdotio.ExportJobConfig{TableName: "iris_test"})
	// if err != nil {
	// 	fmt.Printf("failed to create new export job: %v\n", err)
	// 	os.Exit(1)
	// }
	// fmt.Printf("Export job ID %s and status url %s.\n", exportJob.ID, exportJob.StatusURL)

	// // Test retrieving an export job status
	// exportJob, err = b.GetExportJob(exportJob.ID)
	// if err != nil {
	// 	fmt.Printf("failed to get export job status: %v\n", err)
	// 	os.Exit(1)
	// }
	// fmt.Printf("Export job ID %s and status url %s.\n", exportJob.ID, exportJob.StatusURL)
}
