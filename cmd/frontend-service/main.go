package main

import (
	"log"
	"net/http"

	"github.com/CAPS-Cloud/exercises/internal"

	"github.com/labstack/echo/v4"
)

func main() {
	client, ctx, cancel, err := internal.ConnectDB()
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer cancel()
	defer client.Disconnect(ctx)

	coll, err := internal.PrepareDatabase(client, "exercise-1", "information")
	if err != nil {
		log.Fatalf("Error preparing database: %v", err)
	}
	internal.PrepareData(client, coll)

	e := echo.New()

	// Set the renderer for HTML templates
	e.Renderer = internal.LoadTemplates()

	// Serve static assets like CSS
	e.Static("/css", "css")

	// Routes serving HTML pages

	e.GET("/", func(c echo.Context) error {
		return c.Render(200, "index", nil)
	})

	e.GET("/books", func(c echo.Context) error {
		books := internal.FindAllBooks(coll)
		return c.Render(200, "book-table", books)
	})

	e.GET("/authors", func(c echo.Context) error {
		authors := internal.FindAllAuthors(coll)
		return c.Render(200, "authors", authors)
	})

	e.GET("/years", func(c echo.Context) error {
		years := internal.FindAllYears(coll)
		return c.Render(200, "years", years)
	})

	e.GET("/search", func(c echo.Context) error {
		return c.Render(200, "search-bar", nil)
	})

	e.GET("/create", func(c echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})

	// Start the frontend server on port 8080
	e.Logger.Fatal(e.Start(":8080"))
}
