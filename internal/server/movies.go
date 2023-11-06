package server

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/yoramdelangen/iptv-web-app/internal/surreal"
	"github.com/yoramdelangen/iptv-web-app/internal/types"
	"github.com/yoramdelangen/iptv-web-app/query"
)

func MoviesIndex(c *fiber.Ctx) error {
	qs := c.Context().QueryArgs()

	sql := query.Movies
	parameters := map[string]interface{}{}
	if qs.Has("category") {
		sql = query.MoviesByCategory
		parameters["category_id"] = fmt.Sprintf("%d", qs.GetUintOrZero("category"))
	}

	movies := surreal.Query[[]types.Movie](sql, parameters)
	categories := surreal.Query[[]types.MovieCategory](query.MovieCategories, map[string]interface{}{})

	fmt.Printf("Category: %+v\n", fmt.Sprintf("%v", qs.GetUintOrZero("category")))

	return c.Render("content/movies/index", fiber.Map{
		"ActiveCategory": fmt.Sprintf("%d", qs.GetUintOrZero("category")),
		"Movies":         movies,
		"Categories":     categories,
	})
}

func MovieStream(c *fiber.Ctx) error {
	id := strings.Split(c.Params("id"), ".")[0]

	host := "http://thu.watchbiptv.co:80"
	creds := "4CNPVVkH7v/946265932979"
	url := fmt.Sprintf("%s/movie/%s/%s.mkv", host, creds, id)

	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Internal request failed: ", err)
		return err
	}

	c.Status(resp.StatusCode)
	err = c.SendStream(resp.Body, -1)
	if err != nil {
		if errors.Is(err, syscall.EPIPE) {
			return nil
		}

		return err
	}

	return nil
}
