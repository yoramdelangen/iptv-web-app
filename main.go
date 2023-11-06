package main

import (
	"crypto/md5"
	"embed"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/imroc/req/v3"
	"github.com/surrealdb/surrealdb.go"
	"github.com/yoramdelangen/iptv-web-app/query"
)

const (
	HOST = "http://thu.watchbiptv.co:80"
	PATH = "/player_api.php?username=%s&password=%s"

	USERNAME = "4CNPVVkH7v"
	PASSWORD = "946265932979"

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

func surQuery[T any](query string, payload map[string]interface{}) T {
	res, err := DB.Query(query, payload)
	out, err := surrealdb.SmartUnmarshal[T](res, err)

	if err != nil {
		log.Println("Failed Unmarshal", err)
	}

	return out
}

func SimpleSelect[T any](thing string) T {
	res, err := DB.Select(thing)
	out, err := surrealdb.SmartUnmarshal[T](res, err)

	if err != nil {
		log.Println("Failed Unmarshal:", err)
	}

	return out
}

func Server() {
	app := gin.Default()

	// engine.AddFunc("image_url", func(url, title string) string {
	// 	html := "<img lazy=\"loading\" src=\"%s\" alt=\"%s\" />"
	// 	// if strings.HasPrefix(url, "data:") {
	// 	// 	fmt.Println("image data url", url)
	// 	// 	return url
	// 	// }
	// 	return template. fmt.Sprintf(html, url, title)
	// 	// return fmt.Sprintf("/cache?image=%s", url)
	// })

	// app := fiber.New(fiber.Config{
	// 	Views:       engine,
	// 	ViewsLayout: "layouts/main",
	// })
	// app.Static("/", "./public")

	// app.GET("/", func(c *gin.Context) {
	// 	return c.Render("templates/index", fiber.Map{})
	// })
	//
	// app.GET("/movies", func(c *gin.Context) error {
	// 	qs := c.Context().QueryArgs()
	//
	// 	sql := query.Movies
	// 	parameters := map[string]interface{}{}
	// 	if qs.Has("category") {
	// 		sql = query.MoviesByCategory
	// 		parameters["category_id"] = fmt.Sprintf("%d", qs.GetUintOrZero("category"))
	// 	}
	//
	// 	movies := surQuery[[]Movie](sql, parameters)
	// 	categories := surQuery[[]MovieCategory](query.MovieCategories, map[string]interface{}{})
	//
	// 	fmt.Printf("Category: %+v\n", fmt.Sprintf("%v", qs.GetUintOrZero("category")))
	//
	// 	return c.Render("templates/movies/index", fiber.Map{
	// 		"ActiveCategory": c.Query("category"),
	// 		"Movies":         movies,
	// 		"Categories":     categories,
	// 	})
	// })
	//
	// app.GET("/movies/:id", func(c *gin.Context) {
	// 	thing := "movies:" + c.Params("id")
	// 	movie := SimpleSelect[Movie](thing)
	//
	// 	if _, ok := movie["info"]; !ok {
	// 		api := NewApi(DB)
	//
	// 		api.RunDetails(ACTION_MOVIE_INFO, c.Params("id"))
	// 		movie = SimpleSelect[Movie](thing)
	// 	}
	//
	// 	return c.Render("templates/movies/show", fiber.Map{
	// 		"Movie": movie,
	// 	})
	// })

	app.GET("/stream/:streamid/movies/:id", func(c *gin.Context) {
		const oneMB = 1024 * 1024
		const oneGB = 1024 * oneMB
		const responseSize = 2 * oneGB

		fmt.Println("")
		fmt.Println("")

		fmt.Printf("Params: %+v\n", c.Params)
		fmt.Printf("Headers: %+v\n", c.Request.Header)

		// streamId := c.ParamsInt("streamid")
		id, err := strconv.Atoi(strings.Split(c.Param("id"), ".")[0])
		if err != nil {
			fmt.Println("Error parsing params", err)
			return
		}

		fmt.Println("id", id)

		host := "http://thu.watchbiptv.co:80"
		creds := "4CNPVVkH7v/946265932979"
		url := fmt.Sprintf("%s/movie/%s/%d.mkv", host, creds, id)

		fmt.Println("URL to load", url)

		r, _ := http.NewRequest(http.MethodGet, url, nil)

		// add headers to request
		for h, v := range c.Request.Header {
			r.Header.Add(h, v[0])
		}

		client := http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				fmt.Println("Redirected to:", req.URL)
				fmt.Printf("Redirect headers: %+v\n", req.Header)
				// http.ErrUseLastResponse
				fmt.Println("Via:", via)
				fmt.Println("")
				return nil
			},
		}
		resp, err := client.Do(r)
		if err != nil {
			fmt.Printf("Request error: %s", err)
			return
		}

		allowHeaders := []string{
			"Server",
		}

		fmt.Println("")
		fmt.Println("SET HEADERS FROM REQUEST")
		// copy all headers back to the response
		for header := range resp.Header {
			if slices.Contains(allowHeaders, header) {
				continue
			}
			fmt.Println(header, resp.Header.Get(header))
			c.Writer.Header().Add(header, resp.Header.Get(header))
		}

		defer resp.Body.Close()

		fmt.Println("")

		// bytesRead := 0
		fmt.Println("ALREADY SET HEADERS", c.Writer.Header())

		_, err = io.Copy(c.Writer, resp.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to stream video"})
			return
		}

		// buf := make([]byte, oneMB*4)
		// c.Stream(func(w io.Writer) bool {
		// 	select {
		// 	case <-ctx.Done():
		// 		fmt.Println("Request canceled")
		// 		return false // Stop streaming
		// 	default:
		// 		n, err := resp.Body.Read(buf)
		// 		bytesRead += n
		// 		if err != nil {
		// 			if err == io.EOF {
		// 				fmt.Println("EOF")
		// 				return false
		// 			}
		//
		// 			fmt.Println("Error reading data from the external URL:", err)
		// 			return false
		// 		}
		//
		// 		// fmt.Println("Totally written: ", bytesRead)
		// 		_, writeErr := c.Writer.Write(buf[:n])
		// 		if writeErr != nil {
		// 			fmt.Println("Error writing data:", writeErr)
		// 			return false
		// 		}
		//
		// 		return true
		// 	}
		// })
	})

	// app.GET("/cache", func(c *gin.Context) error {
	// 	qs := c.Context().QueryArgs()
	//
	// 	cp, _ := os.Getwd()
	// 	storagePath := filepath.Join(cp, "storage")
	//
	// 	// create storage folder
	// 	if _, err := os.Stat(storagePath); errors.Is(err, os.ErrNotExist) {
	// 		os.Mkdir(storagePath, 0755)
	// 	}
	//
	// 	if qs.Has("image") {
	// 		img := strings.TrimSpace(c.Query("image"))
	//
	// 		// fmt.Println("ext", path.Ext(img))
	// 		imgHash := GetMD5Hash(img) + path.Ext(img)
	// 		fileStorage := filepath.Join(storagePath, imgHash)
	//
	// 		if _, err := os.Stat(fileStorage); errors.Is(err, os.ErrNotExist) {
	//
	// 			if strings.HasPrefix(img, "data:") {
	// 				fmt.Println("data image", img)
	// 				return c.SendString(img)
	// 			}
	//
	// 			// we should download it.
	// 			resp, err := http.Get(img)
	// 			if err != nil {
	// 				fmt.Println("Error, err")
	// 				return c.Status(404).SendString("Not found because of error: " + err.Error())
	// 			}
	//
	// 			defer resp.Body.Close()
	//
	// 			out, err := os.Create(fileStorage)
	// 			defer out.Close()
	//
	// 			_, err = io.Copy(out, resp.Body)
	// 			if err != nil {
	// 				fmt.Println("Error, err")
	// 				return c.Status(404).SendString("Storing failed: " + err.Error())
	// 			}
	//
	// 			return c.SendStream(out)
	// 		}
	//
	// 		return c.SendFile(fileStorage, true)
	// 	}
	//
	// 	return c.Status(404).SendString("Nothing to see")
	//
	// })

	log.Fatal(app.Run("localhost:3000"))
}

// ====================================================

type Response = []map[string]interface{}
type ResponseDetail = map[string]interface{}

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
	ACTION_MOVIE_INFO
	ACTION_TVSHOW_CATEGORIES
	ACTION_TVSHOWS
	ACTION_TVSHOW_INFO
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

func (x XtreamApi) url(action string) string {
	path := fmt.Sprintf(PATH, USERNAME, PASSWORD)

	return fmt.Sprintf("%s%s&action=%s", HOST, path, action)
}

func (x XtreamApi) RunSingle(_type ActionType) {
	action := x.actions[_type]
	url := x.url(action.Action)

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
	}
}

func (x XtreamApi) RunDetails(_type ActionType, id string) {
	action := x.actions[_type]
	url := x.url(action.Action) + "&vod_id=" + id

	resp, err := x.client.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	res := &ResponseDetail{}
	resp.Unmarshal(res)

	thing := fmt.Sprintf("%s:%v", action.Table, id)
	query := "UPDATE $what MERGE $data"
	params := map[string]interface{}{
		"what": thing,
		"data": res,
	}
	if _, err = x.db.Query(query, params); err != nil {
		fmt.Printf("Failed Info update [%s] %s, error %s\n", action.Table, thing, err)
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
		ACTION_MOVIE_INFO: {
			Action:  "get_vod_info",
			Table:   "movies",
			IdField: "vod_id",
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
		ACTION_TVSHOW_INFO: {
			Action:  "get_serie_info",
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

func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}
