package main

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"slices"
	"time"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Defines a "model" that we can use to communicate with the
// frontend or the database
// More on these "tags" like `bson:"_id,omitempty"`: https://go.dev/wiki/Well-known-struct-tags
type BookStore struct {
	MongoID     primitive.ObjectID `bson:"_id,omitempty"`
	ID          string             `json:"id"`
	BookName    string             `json:"title"`
	BookAuthor  string             `json:"author"`
	BookEdition string             `json:"edition,omitempty"`
	BookPages   string             `json:"pages,omitempty"`
	BookYear    string             `json:"year,omitempty"`
}

// Wraps the "Template" struct to associate a necessary method
// to determine the rendering procedure
type Template struct {
	tmpl *template.Template
}

// Preload the available templates for the view folder.
// This builds a local "database" of all available "blocks"
// to render upon request, i.e., replace the respective
// variable or expression.
// For more on templating, visit https://jinja.palletsprojects.com/en/3.0.x/templates/
// to get to know more about templating
// You can also read Golang's documentation on their templating
// https://pkg.go.dev/text/template
func loadTemplates() *Template {
	return &Template{
		tmpl: template.Must(template.ParseGlob("views/*.html")),
	}
}

// Method definition of the required "Render" to be passed for the Rendering
// engine.
// Contraire to method declaration, such syntax defines methods for a given
// struct. "Interfaces" and "structs" can have methods associated with it.
// The difference lies that interfaces declare methods whether struct only
// implement them, i.e., only define them. Such differentiation is important
// for a compiler to ensure types provide implementations of such methods.
func (t *Template) Render(w io.Writer, name string, data interface{}, ctx echo.Context) error {
	return t.tmpl.ExecuteTemplate(w, name, data)
}

// Here we make sure the connection to the database is correct and initial
// configurations exists. Otherwise, we create the proper database and collection
// we will store the data.
// To ensure correct management of the collection, we create a return a
// reference to the collection to always be used. Make sure if you create other
// files, that you pass the proper value to ensure communication with the
// database
// More on what bson means: https://www.mongodb.com/docs/drivers/go/current/fundamentals/bson/
func prepareDatabase(client *mongo.Client, dbName string, collecName string) (*mongo.Collection, error) {
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

// Here we prepare some fictional data and we insert it into the database
// the first time we connect to it. Otherwise, we check if it already exists.
func prepareData(client *mongo.Client, coll *mongo.Collection) {
	startData := []BookStore{
		{
			ID:          "example1",
			BookName:    "The Vortex",
			BookAuthor:  "José Eustasio Rivera",
			BookEdition: "958-30-0804-4",
			BookPages:   "292",
			BookYear:    "1924",
		},
		{
			ID:          "example2",
			BookName:    "Frankenstein",
			BookAuthor:  "Mary Shelley",
			BookEdition: "978-3-649-64609-9",
			BookPages:   "280",
			BookYear:    "1818",
		},
		{
			ID:          "example3",
			BookName:    "The Black Cat",
			BookAuthor:  "Edgar Allan Poe",
			BookEdition: "978-3-99168-238-7",
			BookPages:   "280",
			BookYear:    "1843",
		},
	}

	// This syntax helps us iterate over arrays. It behaves similar to Python
	// However, range always returns a tuple: (idx, elem). You can ignore the idx
	// by using _.
	// In the topic of function returns: sadly, there is no standard on return types from function. Most functions
	// return a tuple with (res, err), but this is not granted. Some functions
	// might return a ret value that includes res and the err, others might have
	// an out parameter.
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

		} else {
			for _, res := range results {
				cursor.Decode(&res)
				fmt.Printf("%+v\n", res)
			}
		}
	}
}

// Generic method to perform "SELECT * FROM BOOKS" (if this was SQL, which
// it is not :D ), and then we convert it into an array of map. In Golang, you
// define a map by writing map[<key type>]<value type>{<key>:<value>}.
// interface{} is a special type in Golang, basically a wildcard...
func findAllBooks(coll *mongo.Collection) []map[string]interface{} {
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

func findAllAuthors(coll *mongo.Collection) []map[string]interface{} {
	cursor, err := coll.Find(context.TODO(), bson.D{{}})
	var results []BookStore
	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}

	var ret []map[string]interface{}
	for _, res := range results {
		ret = append(ret, map[string]interface{}{
			"BookAuthor": res.BookAuthor,
		})
	}

	return ret
}

func findAllYears(coll *mongo.Collection) []map[string]interface{} {
	cursor, err := coll.Find(context.TODO(), bson.D{{}})
	var results []BookStore
	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}

	var ret []map[string]interface{}
	for _, res := range results {
		ret = append(ret, map[string]interface{}{
			"BookYear": res.BookYear,
		})
	}

	return ret
}

func main() {
	// Connect to the database. Such defer keywords are used once the local
	// context returns; for this case, the local context is the main function
	// By user defer function, we make sure we don't leave connections
	// dangling despite the program crashing. Isn't this nice? :D
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	uri := os.Getenv("DATABASE_URI")
	if uri == "" {
		log.Fatal("DATABASE_URI environment variable not set")
	}
	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		panic(err)
	}

	// // TODO: make sure to pass the proper username, password, and port
	// client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb+srv://mongodb:ex01@cloudcomputingtum.z7rle34.mongodb.net/?retryWrites=true&w=majority&appName=CloudComputingTUM"))

	// This is another way to specify the call of a function. You can define inline
	// functions (or anonymous functions, similar to the behavior in Python)
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	// You can use such name for the database and collection, or come up with
	// one by yourself!
	coll, err := prepareDatabase(client, "exercise-1", "information")
	if err != nil {
		log.Fatalf("Error preparing database: %v", err)
	}

	prepareData(client, coll)

	// Here we prepare the server
	e := echo.New()

	// e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
	// 	return func(c echo.Context) error {
	// 		fmt.Println(c.Request().Method, c.Request().URL.Path)
	// 		return next(c)
	// 	}
	// })

	// Define our custom renderer
	e.Renderer = loadTemplates()

	// Log the requests. Please have a look at echo's documentation on more
	// middleware
	// e.Use(middleware.Logger())

	e.Static("/css", "css")

	// Endpoint definition. Here, we divided into two groups: top-level routes
	// starting with /, which usually serve webpages. For our RESTful endpoints,
	// we prefix the route with /api to indicate more information or resources
	// are available under such route.
	e.GET("/", func(c echo.Context) error {
		return c.Render(200, "index", nil)
	})

	e.GET("/books", func(c echo.Context) error {
		books := findAllBooks(coll)
		return c.Render(200, "book-table", books)
	})

	e.GET("/authors", func(c echo.Context) error {
		authors := findAllAuthors(coll)
		return c.Render(200, "authors", authors)
	})

	e.GET("/years", func(c echo.Context) error {
		years := findAllYears(coll)
		return c.Render(200, "years", years)
	})

	e.GET("/search", func(c echo.Context) error {
		return c.Render(200, "search-bar", nil)
	})

	e.GET("/create", func(c echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})

	// You will have to expand on the allowed methods for the path
	// `/api/route`, following the common standard.
	// A very good documentation is found here:
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Methods
	// It specifies the expected returned codes for each type of request
	// method.
	e.GET("/api/books", func(c echo.Context) error {
		books := findAllBooks(coll)
		return c.JSON(http.StatusOK, books)
	})

	e.GET("/api/authors", func(c echo.Context) error {
		authors := findAllAuthors(coll)
		return c.JSON(http.StatusOK, authors)
	})

	e.GET("/api/books/:id", func(c echo.Context) error {
		id := c.Param("id")

		// Query MongoDB for a book with the matching ID
		var book BookStore
		err := coll.FindOne(context.TODO(), bson.M{"id": id}).Decode(&book)

		if err != nil {
			if err == mongo.ErrNoDocuments {
				return c.JSON(http.StatusNotFound, map[string]string{"error": "Book not found"})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve book"})
		}

		return c.JSON(http.StatusOK, book)
	})

	e.GET("/api/years", func(c echo.Context) error {
		years := findAllYears(coll)
		return c.JSON(http.StatusOK, years)
	})

	e.POST("/api/books", func(c echo.Context) error {
		var book BookStore

		// Bind the request body to the BookStore struct
		if err := c.Bind(&book); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		}

		// Validate mandatory fields
		if book.ID == "" || book.BookName == "" || book.BookAuthor == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Missing mandatory fields: 'id', 'title', or 'author'",
			})
		}

		// Check if the book already exists
		filter := bson.M{
			"id":         book.ID,         // Matches the 'id' field in MongoDB
			"bookname":   book.BookName,   // Matches 'bookname'
			"bookauthor": book.BookAuthor, // Matches 'bookauthor'
			"bookyear":   book.BookYear,   // Matches 'bookyear'
			"bookpages":  book.BookPages,  // Matches 'bookpages'
		}

		log.Printf("POST filter: %+v\n", filter)

		count, err := coll.CountDocuments(context.TODO(), filter)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to check for existing book"})
		}

		if count > 0 {
			return c.JSON(http.StatusConflict, map[string]string{"error": "Book already exists"})
		}

		// Insert the book into the database
		result, err := coll.InsertOne(context.TODO(), book)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to insert book"})
		}

		// Return success response with the inserted ID
		return c.JSON(http.StatusCreated, map[string]interface{}{
			"message": "Book created successfully",
			"id":      result.InsertedID,
		})
	})

	e.PUT("/api/books/:id", func(c echo.Context) error {
		id := c.Param("id")
		var updatesFromRequest map[string]interface{}
		var updates bson.M = make(bson.M)

		// Bind the request body to a map
		if err := c.Bind(&updatesFromRequest); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		}

		// Map the keys to the lowercase versions used in MongoDB
		if title, ok := updatesFromRequest["title"]; ok {
			updates["bookname"] = title
		}
		if pages, ok := updatesFromRequest["pages"]; ok {
			updates["bookpages"] = pages
		}
		if author, ok := updatesFromRequest["author"]; ok {
			updates["bookauthor"] = author
		}
		if edition, ok := updatesFromRequest["edition"]; ok {
			updates["bookedition"] = edition
		}
		if year, ok := updatesFromRequest["year"]; ok {
			updates["bookyear"] = year
		}

		// Remove ID and MongoID from updates (if they were accidentally sent)
		delete(updates, "id")
		delete(updates, "_id")

		// Update the book in the database
		filter := bson.M{"id": id}
		update := bson.M{"$set": updates}
		result, err := coll.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update book"})
		}
		if result.MatchedCount == 0 {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Book not found"})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Book updated successfully"})
	})

	// DELETE: Delete a book by ID
	e.DELETE("/api/books/:id", func(c echo.Context) error {
		id := c.Param("id")

		// Delete the book from the database
		filter := bson.M{"id": id}
		result, err := coll.DeleteOne(context.TODO(), filter)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete book"})
		}
		if result.DeletedCount == 0 {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Book not found"})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Book deleted successfully"})
	})

	// We start the server and bind it to port 3030. For future references, this
	// is the application's port and not the external one. For this first exercise,
	// they could be the same if you use a Cloud Provider. If you use ngrok or similar,
	// they might differ.
	// In the submission website for this exercise, you will have to provide the internet-reachable
	// endpoint: http://<host>:<external-port>
	e.Logger.Fatal(e.Start(":8080"))
}
