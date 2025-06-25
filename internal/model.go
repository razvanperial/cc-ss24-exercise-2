package internal

import (
	"context"
	"fmt"
	"log"
	"os"
	"slices"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// BookStore model (same as before)
type BookStore struct {
	MongoID     primitive.ObjectID `bson:"_id,omitempty"`
	ID          string             `json:"id"`
	BookName    string             `json:"title"`
	BookAuthor  string             `json:"author"`
	BookEdition string             `json:"edition,omitempty"`
	BookPages   string             `json:"pages,omitempty"`
	BookYear    string             `json:"year,omitempty"`
}

func PrepareDatabase(client *mongo.Client, dbName, collecName string) (*mongo.Collection, error) {
	db := client.Database(dbName)

	names, err := db.ListCollectionNames(context.TODO(), bson.D{{}})
	if err != nil {
		return nil, err
	}
	if !slices.Contains(names, collecName) {
		cmd := bson.D{{"create", collecName}}
		var result bson.M
		if err = db.RunCommand(context.TODO(), cmd).Decode(&result); err != nil {
			log.Fatal(err)
			return nil, err
		}
	}

	coll := db.Collection(collecName)
	return coll, nil
}

func PrepareData(client *mongo.Client, coll *mongo.Collection) {
	startData := []BookStore{
		{ID: "example1", BookName: "The Vortex", BookAuthor: "José Eustasio Rivera", BookEdition: "958-30-0804-4", BookPages: "292", BookYear: "1924"},
		{ID: "example2", BookName: "Frankenstein", BookAuthor: "Mary Shelley", BookEdition: "978-3-649-64609-9", BookPages: "280", BookYear: "1818"},
		{ID: "example3", BookName: "The Black Cat", BookAuthor: "Edgar Allan Poe", BookEdition: "978-3-99168-238-7", BookPages: "280", BookYear: "1843"},
	}

	for _, book := range startData {
		cursor, err := coll.Find(context.TODO(), book)
		var results []BookStore
		if err = cursor.All(context.TODO(), &results); err != nil {
			panic(err)
		}
		if len(results) > 1 {
			log.Fatal("more records were found")
		} else if len(results) == 0 {
			result, err := coll.InsertOne(context.TODO(), book)
			if err != nil {
				panic(err)
			} else {
				fmt.Printf("%+v\n", result)
			}
		}
	}
}

// Helper to connect to MongoDB
func ConnectDB() (*mongo.Client, context.Context, context.CancelFunc, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	uri := os.Getenv("DATABASE_URI")
	if uri == "" {
		cancel()
		return nil, nil, nil, fmt.Errorf("DATABASE_URI not set")
	}
	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		cancel()
		return nil, nil, nil, err
	}
	return client, ctx, cancel, nil
}

func FindAllBooks(coll *mongo.Collection) []map[string]interface{} {
	cursor, err := coll.Find(context.TODO(), bson.D{{}})
	var results []BookStore
	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}

	var ret []map[string]interface{}

	for _, res := range results {
		ret = append(ret, map[string]interface{}{
			"id":      res.ID,          // Changed "ID" to "id" and using res.ID
			"title":   res.BookName,    // Changed "BookName" to "title"
			"author":  res.BookAuthor,  // Changed "BookAuthor" to "author"
			"pages":   res.BookPages,   // Changed "BookPages" to "pages"
			"edition": res.BookEdition, // Changed "BookEdition" to "edition"
			"year":    res.BookYear,    // Added "year"
		})
	}

	return ret
}

func FindAllAuthors(coll *mongo.Collection) []map[string]interface{} {
	cursor, err := coll.Find(context.TODO(), bson.D{{}})
	var results []BookStore
	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}

	var ret []map[string]interface{}
	for _, res := range results {
		ret = append(ret, map[string]interface{}{
			"BookName":   res.BookName,
			"BookAuthor": res.BookAuthor,
		})
	}

	return ret
}

func FindAllYears(coll *mongo.Collection) []map[string]interface{} {
	cursor, err := coll.Find(context.TODO(), bson.D{{}})
	var results []BookStore
	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}

	var ret []map[string]interface{}
	for _, res := range results {
		ret = append(ret, map[string]interface{}{
			"BookName": res.BookName,
			"BookYear": res.BookYear,
		})
	}

	return ret
}
