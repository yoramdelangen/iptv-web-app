package server

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/spf13/viper"
	"github.com/yoramdelangen/iptv-web-app/internal/accounts"
	"github.com/yoramdelangen/iptv-web-app/internal/types"
	"github.com/yoramdelangen/iptv-web-app/internal/xtream"
)

func XtreamPlayer(c *fiber.Ctx) error {
	account := c.Params("account")
	queries := c.Queries()
	action := queries["action"]

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
		return root(c)
	}

	a, err := xtream.ByAction(action)
	if err != nil {
		return err
	}

	resp, err := forwardRequest(account, queries, a)

	if err != nil {
		return err
	}

	c.JSON(resp)

	return nil
}

func XtreamXML(c *fiber.Ctx) error {
	account := c.Params("account")

	r := xtream.NewRequest()
	payload := types.ResponseDetail{}

	conf := viper.GetStringMap(fmt.Sprintf("accounts.%s", account))
	if len(conf) == 0 {
		return errors.New(fmt.Sprintf("Account '%s' not found", account))
	}

	url := fmt.Sprintf("%s/xmltv.php?username=%s&password=%d", conf["url"], conf["username"], conf["password"])

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

func forwardRequest(account string, queries map[string]string, a xtream.Action) (interface{}, error) {
	r := xtream.NewRequest()
	var payload interface{}

	conf := viper.GetStringMap(fmt.Sprintf("accounts.%s", account))
	if len(conf) == 0 {
		return payload, errors.New(fmt.Sprintf("Account '%s' not found", account))
	}

	url := fmt.Sprintf("%s/player_api.php", conf["url"])

	fmt.Printf("Conig: %+v\n", conf)
	fmt.Printf("Url: %+v\n", url)

	queries["username"] = fmt.Sprintf("%s", conf["username"])
	queries["password"] = fmt.Sprintf("%d", conf["password"])

	fmt.Printf("queries before sending: %+v\n", queries)

	_, err := r.
		SetQueryParams(queries).
		SetSuccessResult(&payload).
		Get(url)

	if err != nil {
		return payload, err
	}

	return payload, nil
}

// being used to check if the credentails match or not....
func root(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"user_info": fiber.Map{
			"username":        "4CNPVVkH7v",
			"password":        "946265932979",
			"message":         "Dtv & TeslaVision",
			"auth":            1,
			"status":          "Active",
			"exp_date":        "1702668321",
			"is_trial":        "0",
			"active_cons":     "1",
			"created_at":      "1694802321",
			"max_connections": "1",
			"allowed_output_formats": []string{
				"m3u8",
				"ts",
			},
		},
		"server_info": fiber.Map{
			"url":             "a.flixtv.org",
			"port":            "8080",
			"https_port":      "8443",
			"server_protocol": "http",
			"rtmp_port":       "25462",
			"timezone":        "Europe/Amsterdam",
			"timestamp_now":   1700061506,
			"time_now":        "2023-11-15 16:18:26",
		},
	})
}
