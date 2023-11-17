package server

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/yoramdelangen/iptv-web-app/internal/accounts"
	"github.com/yoramdelangen/iptv-web-app/internal/types"
	"github.com/yoramdelangen/iptv-web-app/internal/xtream"
)

func XtreamPlayer(c *fiber.Ctx) error {
	queries := c.Queries()
	action := queries["action"]
	account, err := accounts.Get(c.Params("account"))
	if err != nil {
		return err
	}

	if c.Method() == fiber.MethodPost {
		values, err := url.ParseQuery(string(c.Body()))
		if err != nil {
			return err
		}

		queries = make(map[string]string)
		for k, v := range values {
			queries[k] = fmt.Sprintf("%v", v[0])
		}
		action = queries["action"]
	}

	// in case no action
	if len(action) == 0 {
		return root(c, account)
	}

	resp, err := forwardRequest[interface{}](account, queries)

	if err != nil {
		return err
	}

	c.JSON(resp)

	return nil
}

func XtreamXML(c *fiber.Ctx) error {
	account, err := accounts.Get(c.Params("account"))
	if err != nil {
		return err
	}

	r := xtream.NewRequest()
	payload := types.ResponseDetail{}

	url := fmt.Sprintf("%s/xmltv.php?username=%s&password=%d", account.Url, account.Username, account.Password)

	resp, err := r.Get(url)
	if err != nil {
		return err
	}

	resp.UnmarshalJson(&payload)

	c.SendStatus(resp.StatusCode)
	c.JSON(payload)

	return nil
}

func XtreamMedia(c *fiber.Ctx) error {
	id := c.Params("id")
	mt := c.Params("type")

	account, err := accounts.Get(c.Params("account"))
	if err != nil {
		return err
	}

	p := []string{account.Url, mt, account.Username, account.Password, id}
	url := fmt.Sprintf("%s.mkv", strings.Join(p, "/"))

	return xtream.CreateMediaStream(url, c)
}

func XtreamLive(c *fiber.Ctx) error {
	id := c.Params("id")

	fmt.Printf("Request: %+v", c.Request())

	account, err := accounts.Get(c.Params("account"))
	if err != nil {
		return err
	}

	p := []string{account.Url, account.Username, account.Password, id}
	url := strings.Join(p, "/")

	fmt.Println("url", url)

	return xtream.CreateMediaStream(url, c)
}

func forwardRequest[T interface{}](account accounts.Account, queries map[string]string) (T, error) {
	r := xtream.NewRequest()
	var payload T

	url := fmt.Sprintf("%s/player_api.php", account.Url)

	queries["username"] = account.Username
	queries["password"] = account.Password

	_, err := r.
		SetQueryParams(queries).
		SetSuccessResult(&payload).
		Get(url)

	if err != nil {
		return payload, err
	}

	return payload, nil
}

type RootResponse struct {
	UserInfo   UserInfo   `json:"user_info"`
	ServerInfo ServerInfo `json:"server_info"`
}

type UserInfo struct {
	Username            string   `json:"username"`
	Password            string   `json:"password"`
	Message             string   `json:"message"`
	Auth                int      `json:"auth"`
	Status              string   `json:"status"`
	ExpDate             string   `json:"exp_date"`
	IsTrial             string   `json:"is_trial"`
	ActiveCons          string   `json:"active_cons"`
	CreatedAt           string   `json:"created_at"`
	MaxConnections      string   `json:"max_connections"`
	AllowedOutputFormat []string `json:"allowed_output_formats"`
}

type ServerInfo struct {
	Url            string `json:"url"`
	Port           string `json:"port"`
	HttpsPort      string `json:"https_port"`
	ServerProtocol string `json:"server_protocol"`
	RtmpPort       string `json:"rtmp_port"`
	Timezone       string `json:"timezone"`
	TimezoneNow    int64  `json:"timestamp_now"`
	TimeNow        string `json:"time_now"`
}

// being used to check if the credentails match or not....
func root(c *fiber.Ctx, account accounts.Account) error {
	queries := make(map[string]string)

	resp, err := forwardRequest[RootResponse](account, queries)

	if err != nil {
		return err
	}

	resp.UserInfo.Username = c.Query("username")
	resp.UserInfo.Password = c.Query("password")
	resp.ServerInfo.Url = c.Hostname()

	return c.JSON(resp)
}
