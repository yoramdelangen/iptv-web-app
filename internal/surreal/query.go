package surreal

import (
	"log"

	"github.com/surrealdb/surrealdb.go"
)

var DB *surrealdb.DB

func init() {
	DB = Db()
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

func Query[T any](query string, payload map[string]interface{}) T {
	res, err := DB.Query(query, payload)
	out, err := surrealdb.SmartUnmarshal[T](res, err)

	if err != nil {
		log.Println("Failed Unmarshal", err)
	}

	return out
}
