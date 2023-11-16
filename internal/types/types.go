package types

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

type Response = []map[string]interface{}
type ResponseDetail = map[string]interface{}
