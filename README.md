## go-bitdotio

Very early work-in-progress on a Go SDK for bit.io. Many TODO's are noted 
throughout the code. At a high-level, major TODO's include:
- Handling HTTP request/wrapper methods with optional body keys
- API Import/Export
- Query Endpoint
- Postgres connections/pooling
- Tests
- Real logging
- Organizing as a proper Go module

Testing is not set up yet, `main.go` demonstrates the initial progress.

Demo:

```go
func main() {
	// Setup
	token := os.Getenv("BITDOTIO_TOKEN")
	b := bitdotio.NewBitDotIO(token)
	newDBName := "foo_db3"

	// Demonstrate creating a database
	newDatabase, err := b.CreateDatabase(newDBName, true)
	if err != nil {
		fmt.Printf("main failed to create database: %v", err)
		os.Exit(1)
	}
	fmt.Printf("Create Example: %v\n", newDatabase.Name)

	// Demonstrate listing databases
	databaseList, err := b.ListDatabases()
	if err != nil {
		fmt.Printf("main failed to list databases: %v", err)
		os.Exit(1)
	}
	fmt.Printf("Found %d databases:\n", len(databaseList.Databases))
	for _, db := range databaseList.Databases {
		fmt.Printf("- %v\n", db.Name)
	}

	// Demonstrate getting a database
	database, err := b.GetDatabase("andrewdoss", newDBName)
	if err != nil {
		fmt.Printf("main failed to get database: %v", err)
		os.Exit(1)
	}
	fmt.Printf("Get Example: %v\n", database.Name)

	// Confirm update
	updatedDBName := newDBName + "-updated"
	database, err = b.UpdateDatabase("andrewdoss", newDBName, updatedDBName, true)
	if err != nil {
		fmt.Printf("main failed to update database: %v", err)
		os.Exit(1)
	}
	fmt.Printf("Update Example: %v\n", database.Name)

	// Demonstrate deleting a database
	err = b.DeleteDatabase("andrewdoss", updatedDBName)
	if err != nil {
		fmt.Printf("main failed to delete database: %v", err)
		os.Exit(1)
	}
	// Confirm deletion
	databaseList, err = b.ListDatabases()
	if err != nil {
		fmt.Printf("main failed to list databases: %v", err)
		os.Exit(1)
	}
	fmt.Printf("Confirming delete... found %d databases\n", len(databaseList.Databases))
}
```