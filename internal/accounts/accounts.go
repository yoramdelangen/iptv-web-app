package accounts

import (
	"errors"
	"fmt"

	"github.com/spf13/viper"
)

type Account struct {
  Name string `json:"name"`
  Username string  `json:"username"`
  Password string `json:"password"`
  Url string `json:"url"`
}

func Get(id string) (Account, error) {
	conf := viper.GetStringMap(fmt.Sprintf("accounts.%s", id))

  if len(conf) == 0 {
    return Account{}, errors.New(fmt.Sprintf("no account found with %s", id))
  }

  return Account{
    Name: conf["name"].(string),
    Username: conf["username"].(string),
    Password: fmt.Sprintf("%d", conf["password"]),
    Url: conf["url"].(string),
  }, nil
}
