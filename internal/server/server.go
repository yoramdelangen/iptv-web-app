package server

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/html/v2"
	"github.com/yoramdelangen/iptv-web-app/internal/statics"
	"github.com/yoramdelangen/iptv-web-app/templates"
)

type UserInfo struct {
	Username            string
	Password            string
	Message             string
	Auth                int
	Status              string
	ExpDate             string `json:"exp_date"`
	IsTrial             string `json:"is_trial"`
	ActiveCons          string
	CreatedAt           string
	MaxConnections      string
	AllowedOutputFormat []string
}

type ServerInfo struct {
	Url            string
	Port           string
	HttpsPort      string
	ServerProtocol string
	RtmpPort       string
	Timezone       string
	TimezoneNow    int64
	TimeNow        string
}

func New() *fiber.App {
	engine := html.NewFileSystem(http.FS(templates.Files), ".html")
	engine.Reload(true)
	engine.Debug(true)

	app := fiber.New(fiber.Config{
		Views:       engine,
		ViewsLayout: "layouts/main",
	})
	app.Use(logger.New())

	app.Use("/assets", filesystem.New(filesystem.Config{
		Root:       http.FS(statics.Assets),
		PathPrefix: "assets",
		Browse:     true,
	}))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("index", fiber.Map{})
	})

	app.Get("/movies", MoviesIndex)
	app.Get("/movies/:id", MovieShow)
	app.Get("/stream/:streamid/movies/:id", MovieStream)


	app.All("/:account/player_api.php", XtreamPlayer)
	app.Get("/:account/xmltv.php", XtreamXML)
	app.Get("/:account/:type/*/*/:id.mkv", XtreamMedia)

	return app
}
