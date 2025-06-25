package main

import (
	"context"
	"net/http"

	"github.com/CAPS-Cloud/exercises/internal"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func main() {
	client, ctx, cancel, err := internal.ConnectDB()
	if err != nil {
		panic(err)
	}
	defer cancel()
	defer client.Disconnect(ctx)

	coll, err := internal.PrepareDatabase(client, "exercise-1", "information")
	if err != nil {
		panic(err)
	}
	internal.PrepareData(client, coll)

	e := echo.New()

	e.GET("/api/books", func(c echo.Context) error {
		books := internal.FindAllBooks(coll)
		return c.JSON(http.StatusOK, books)
	})

	e.GET("/api/authors", func(c echo.Context) error {
		authors := internal.FindAllAuthors(coll)
		return c.JSON(http.StatusOK, authors)
	})

	e.GET("/api/books/:id", func(c echo.Context) error {
		id := c.Param("id")

		// Query MongoDB for a book with the matching ID
		var book internal.BookStore
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
		years := internal.FindAllYears(coll)
		return c.JSON(http.StatusOK, years)
	})

	e.Logger.Fatal(e.Start(":8081"))
}
