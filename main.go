package main

import (
	"embed"
	"fmt"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/proxy"
	"github.com/gofiber/template/html/v2"
	"github.com/imroc/req/v3"
	"github.com/surrealdb/surrealdb.go"
	"github.com/yoramdelangen/iptv-web-app/query"
)

const (
	HOST = "http://thu.watchbiptv.co:80"
	PATH = "/player_api.php?username=%s&password=%s"

	STREAM      = "/%s/%s" // Username/Password
	STREAM_LIVE = "/live/%s/%s"
)

var DB *surrealdb.DB

func init() {
	DB = Db()
}

func main() {
	// SyncXtreamApi()
	Server()
}

//go:embed templates/* layouts/*
var templates embed.FS

type Movie map[string]interface{}
type MovieCategory map[string]interface{}

type AnyResponse[T any] interface {
	Movie | MovieCategory
}

type SurrealResponse[T any] struct {
	Result []T
	Status bool
	time   string
}

func surQuery[T any](query string, payload map[string]interface{}) T {
	res, err := DB.Query(query, payload)
	out, err := surrealdb.SmartUnmarshal[T](res, err)

	if err != nil {
		log.Println("Failed Unmarshal", err)
	}

	return out
}

func Server() {
	engine := html.NewFileSystem(http.FS(templates), ".html")
	engine.Reload(true)
	engine.Debug(true)

	app := fiber.New(fiber.Config{
		Views:       engine,
		ViewsLayout: "layouts/main",
	})
	app.Use(logger.New())
	app.Static("/", "./public")

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("templates/main", fiber.Map{})
	})

	app.Get("/movies", func(c *fiber.Ctx) error {
		qs := c.Context().QueryArgs()

		sql := query.Movies
		parameters := map[string]interface{}{}
		if qs.Has("category") {
			sql = query.MoviesByCategory
			parameters["category_id"] = fmt.Sprintf("%d", qs.GetUintOrZero("category"))
		}

		movies := surQuery[[]Movie](sql, parameters)
		categories := surQuery[[]MovieCategory](query.MovieCategories, map[string]interface{}{})

		fmt.Printf("Category: %+v\n", fmt.Sprintf("%v", qs.GetUintOrZero("category")))

		return c.Render("templates/movies/index", fiber.Map{
			"ActiveCategory": fmt.Sprintf("%d", qs.GetUintOrZero("category")),
			"Movies":         movies,
			"Categories":     categories,
		})
	})

	app.Get("/stream/:streamid/movies/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")

		host := "http://thu.watchbiptv.co:80"
		creds := "4CNPVVkH7v/946265932979"
		url := fmt.Sprintf("%s/movie/%s/%s", host, creds, id)

		fmt.Println("URL", url)

		if err := proxy.Do(c, url); err != nil {
			return err
		}
		// Remove Server header from response
		c.Response().Header.Del(fiber.HeaderServer)
		return nil
	})

	log.Fatal(app.Listen("localhost:3000"))
}

type Response = []map[string]interface{}

func SyncXtreamApi() {
	api := NewApi(DB)
	api.RunAll()
	api.CategoryStats()

	fmt.Printf("API XTREAM %+v\n", api)
}

type ActionIdType int

const (
	ACTION_ID_TYPE_FLOAT ActionIdType = iota
	ACTION_ID_TYPE_STRING
)

type ActionType int

const (
	ACTION_NONE ActionType = iota
	ACTION_MOVIE_CATEGORIES
	ACTION_MOVIES
	ACTION_TVSHOWS
	ACTION_TVSHOW_CATEGORIES
	ACTION_LIVESTREAMS
	ACTION_LIVESTREAM_CATEGORIES
)

type Action struct {
	Action  string
	Table   string
	IdField string
	IdType  ActionIdType // float64 or string
	For     ActionType
}

type XtreamApi struct {
	db      *surrealdb.DB
	client  *req.Request
	actions map[ActionType]Action
}

func (x XtreamApi) RunAll() {
	for action := range x.actions {
		x.RunSingle(action)
	}

	fmt.Println("Ran all actions")
}

func (x XtreamApi) RunSingle(_type ActionType) {
	action := x.actions[_type]
	path := fmt.Sprintf(PATH, "4CNPVVkH7v", "946265932979")
	url := fmt.Sprintf("%s%s&action=%s", HOST, path, action.Action)

	fmt.Println("Starting", action.Action)

	resp, err := x.client.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	res := &Response{}
	resp.Unmarshal(res)

	// loop through the result set
	for _, item := range *res {
		// get the proper typed ID of the item and create a Surreal THING
		id := item[action.IdField]
		if action.IdType == ACTION_ID_TYPE_FLOAT {
			id = int(id.(float64))
		}
		thing := fmt.Sprintf("%s:%v", action.Table, id)

		// Query if the record exists and update if it does
		_, err := x.db.Select(thing)
		if err == nil {
			if _, err = x.db.Update(thing, item); err != nil {
				fmt.Printf("Failed update [%s] %s, error %s\n", action.Table, thing, err)
			}
			continue
		}

		// Create new record
		_, err = x.db.Create(thing, item)
		if err != nil {
			fmt.Printf("Failed [%s] %s, error %s\n", action.Table, thing, err)
		}
	}

	fmt.Println("Done syncing:", action.Table)
}

func (x XtreamApi) CategoryStats() {
	for _, action := range x.actions {
		// skipping
		if action.For == ACTION_NONE {
			continue
		}

		target := x.actions[action.For]

		q := query.UpdateCategoryStats(action.IdField, target.IdField, target.Table)

		x.db.Query(q, map[string]interface{}{
			"table": action.Table,
		})

		fmt.Println("Finished for table", action.Table)

		// ACTION_MOVIE_CATEGORIES: {
		// 	Action:  "get_vod_categories",
		// 	Table:   "movie_categories",
		// 	IdField: "category_id",
		// 	IdType:  ACTION_ID_TYPE_STRING,
		// 	For:     ACTION_MOVIES,
		// },
		// ACTION_MOVIES: {
		// 	Action:  "get_vod_streams",
		// 	Table:   "movies",
		// 	IdField: "stream_id",
		// 	IdType:  ACTION_ID_TYPE_FLOAT,
		// 	For:     ACTION_NONE,
		// },
	}

}

func NewApi(db *surrealdb.DB) *XtreamApi {
	client := req.C()
	r := client.R()

	actions := map[ActionType]Action{
		ACTION_MOVIE_CATEGORIES: {
			Action:  "get_vod_categories",
			Table:   "movie_categories",
			IdField: "category_id",
			IdType:  ACTION_ID_TYPE_STRING,
			For:     ACTION_MOVIES,
		},
		ACTION_MOVIES: {
			Action:  "get_vod_streams",
			Table:   "movies",
			IdField: "stream_id",
			IdType:  ACTION_ID_TYPE_FLOAT,
			For:     ACTION_NONE,
		},
		ACTION_TVSHOW_CATEGORIES: {
			Action:  "get_series_categories",
			Table:   "tvshow_categories",
			IdField: "category_id",
			IdType:  ACTION_ID_TYPE_STRING,
			For:     ACTION_TVSHOWS,
		},
		ACTION_TVSHOWS: {
			Action:  "get_series",
			Table:   "tvshows",
			IdField: "series_id",
			IdType:  ACTION_ID_TYPE_FLOAT,
			For:     ACTION_NONE,
		},
		ACTION_LIVESTREAM_CATEGORIES: {
			Action:  "get_live_categories",
			Table:   "live_stream_categories",
			IdField: "category_id",
			IdType:  ACTION_ID_TYPE_STRING,
			For:     ACTION_LIVESTREAMS,
		},
		ACTION_LIVESTREAMS: {
			Action:  "get_live_streams",
			Table:   "live_streams",
			IdField: "stream_id",
			IdType:  ACTION_ID_TYPE_FLOAT,
			For:     ACTION_NONE,
		},
	}

	return &XtreamApi{
		db:      db,
		client:  r,
		actions: actions,
	}
}

// Connect with SurrealDB
// TODO: add configuration .yaml/toml
func Db() *surrealdb.DB {
	db, err := surrealdb.New("ws://localhost:8000/rpc")
	if err != nil {
		panic(err)
	}

	if _, err = db.Signin(map[string]interface{}{
		"user": "root",
		"pass": "root",
	}); err != nil {
		panic(err)
	}

	db.Use("iptv_ns", "iptv_db")

	return db
}
