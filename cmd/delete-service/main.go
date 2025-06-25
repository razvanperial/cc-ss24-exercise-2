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

	e.DELETE("/api/books/:id", func(c echo.Context) error {
		id := c.Param("id")

		// Delete the book from the database
		filter := bson.M{"id": id}
		result, err := coll.DeleteOne(context.TODO(), filter)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete book"})
		}
		if result.DeletedCount == 0 {
			return c.JSON(http.StatusOK, map[string]string{"message": "Book not found"})

		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Book deleted successfully"})
	})

	e.Logger.Fatal(e.Start(":8082"))
}
