package main

import (
	"fmt"
	"os"

	"github.com/bitdotioinc/go-bitdotio/bitdotio"
)

func main() {
	// Setup
	token := os.Getenv("BITDOTIO_TOKEN")
	b := bitdotio.NewBitDotIO(token)

	// newDBName := "foo_db5"
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
	// databaseList, err := b.ListDatabases()
	// if err != nil {
	// 	fmt.Printf("main failed to list databases: %v", err)
	// 	os.Exit(1)
	// }
	// fmt.Printf("Found %d databases:\n", len(databaseList.Databases))
	// for _, db := range databaseList.Databases {
	// 	fmt.Printf("- %v\n", db.Name)
	// }

	// // Demonstrate getting a database
	// database, err := b.GetDatabase("andrewdoss", newDBName)
	// if err != nil {
	// 	fmt.Printf("main failed to get database: %v", err)
	// 	os.Exit(1)
	// }
	// fmt.Printf("Get Example: %v\n", database.Name)

	// // Confirm update
	// updatedDBName := newDBName + "-updated"
	// database, err = b.UpdateDatabase(
	// 	"andrewdoss",
	// 	newDBName,
	// 	&bitdotio.DatabaseConfig{Name: updatedDBName, IsPrivate: true},
	// )
	// if err != nil {
	// 	fmt.Printf("main failed to update database: %v", err)
	// 	os.Exit(1)
	// }
	// fmt.Printf("Update Example: %v\n", database.Name)

	// // Demonstrate deleting a database
	// err = b.DeleteDatabase("andrewdoss", updatedDBName)
	// if err != nil {
	// 	fmt.Printf("main failed to delete database: %v", err)
	// 	os.Exit(1)
	// }
	// // Confirm deletion
	// databaseList, err = b.ListDatabases()
	// if err != nil {
	// 	fmt.Printf("main failed to list databases: %v", err)
	// 	os.Exit(1)
	// }
	// fmt.Printf("Confirming delete... found %d databases\n", len(databaseList.Databases))

	// Test creating a new API key
	// credentials, err := b.CreateKey()
	// if err != nil {
	// 	fmt.Printf("main failed to create a personal key: %v", err)
	// 	os.Exit(1)
	// }
	// fmt.Printf("Username: %s, Key: %s\n", credentials.Username, credentials.APIKEY)

	// Test listing service accounts
	serviceAccounts, err := b.ListServiceAccounts()
	if err != nil {
		fmt.Printf("main failed to list service accounts: %v", err)
		os.Exit(1)
	}
	fmt.Printf("Found %d service accounts:\n", len(serviceAccounts.ServiceAccounts))
	for _, s := range serviceAccounts.ServiceAccounts {
		fmt.Printf("- %s with role %s\n", s.Name, s.Role)
		for _, db := range s.Databases {
			fmt.Printf("    - %s\n", db.Name)
		}
	}
}
