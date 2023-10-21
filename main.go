package main

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/imroc/req/v3"
	"github.com/surrealdb/surrealdb.go"
)

const (
	HOST = "http://thu.watchbiptv.co:80"
	PATH = "/player_api.php?username=%s&password=%s"

	STREAM      = "/%s/%s" // Username/Password
	STREAM_LIVE = "/live/%s/%s"
)

func main() {
	SyncXtreamApi()
}

func Server() {
	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	app.Listen(":3000")
}

type Response = []map[string]interface{}

func SyncXtreamApi() {
	// Lets start
	db := Db()

	fmt.Println("kk")

	api := NewApi(db)

	api.RunAll()

	// db.Create("movie_categories")
}

type ActionIdType int
const (
	ACTION_ID_TYPE_FLOAT ActionIdType = iota
	ACTION_ID_TYPE_STRING
)

type ActionType int
const (
	ACTION_MOVIE_CATEGORIES ActionType = iota
	ACTION_MOVIES
	ACTION_TVSHOWS
	ACTION_TVSHOW_CATEGORIES
	ACTION_LIVESTREAMS
	ACTION_LIVESTREAM_CATEGORIES
)

type Action struct {
	Action string
	Table string
	IdField string
	IdType ActionIdType// float64 or string
}

type XtreamApi struct {
	db *surrealdb.DB
	client *req.Request
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

func NewApi(db *surrealdb.DB) *XtreamApi {
	client := req.C()
	r := client.R()

	actions := map[ActionType]Action{
		ACTION_MOVIE_CATEGORIES: {
			Action: "get_vod_categories",
			Table: "movie_categories",
			IdField: "category_id",
			IdType: ACTION_ID_TYPE_STRING,
		},
		ACTION_MOVIES: {
			Action:  "get_vod_streams",
			Table:   "movies",
			IdField: "stream_id",
			IdType:  ACTION_ID_TYPE_FLOAT,
		},
		ACTION_TVSHOW_CATEGORIES: {
			Action:  "get_series_categories",
			Table:   "tvshow_categories",
			IdField: "category_id",
			IdType:  ACTION_ID_TYPE_STRING,
		},
		ACTION_TVSHOWS: {
			Action:  "get_series",
			Table:   "tvshows",
			IdField: "series_id",
			IdType:  ACTION_ID_TYPE_FLOAT,
		},
		ACTION_LIVESTREAM_CATEGORIES: {
			Action:  "get_live_categories",
			Table:   "live_stream_categories",
			IdField: "category_id",
			IdType:  ACTION_ID_TYPE_STRING,
		},
		ACTION_LIVESTREAMS: {
			Action:  "get_live_streams",
			Table:   "live_streams",
			IdField: "stream_id",
			IdType:  ACTION_ID_TYPE_FLOAT,
		},
	}

	return &XtreamApi{
		db: db,
		client: r,
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
