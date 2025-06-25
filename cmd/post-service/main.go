package main

import (
	"context"
	"net/http"

	"github.com/CAPS-Cloud/exercises/internal"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
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

	e.POST("/api/books", func(c echo.Context) error {
		var book internal.BookStore
		if err := c.Bind(&book); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		}

		if book.ID == "" || book.BookName == "" || book.BookAuthor == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing mandatory fields"})
		}

		filter := bson.M{"id": book.ID}
		count, err := coll.CountDocuments(context.TODO(), filter)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to check for existing book"})
		}

		if count > 0 {
			return c.JSON(http.StatusConflict, map[string]string{"error": "Book already exists"})
		}

		result, err := coll.InsertOne(context.TODO(), book)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to insert book"})
		}

		return c.JSON(http.StatusCreated, map[string]interface{}{
			"message": "Book created successfully",
			"id":      result.InsertedID,
		})
	})

	e.Logger.Fatal(e.Start(":8083"))
}
