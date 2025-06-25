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

	e.Logger.Fatal(e.Start(":8084"))
}
