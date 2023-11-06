package query

import "fmt"

const Movies = `
	select *
	from movies
	order by added desc
	limit 100
`
const MoviesByCategory = `
	select *
	from movies
	where category_id = $category_id
	order by added desc
	limit 100
`
const Movie = `
	select *
	from $thing
`

const MovieCategories = `
	select *
	from movie_categories
	order by category_name
`

func UpdateCategoryStats(idField, targetId, targetTable string) string {
	targetField := fmt.Sprintf("%s_count", targetTable)
	return fmt.Sprintf(`
	update type::table($table)
	set %s = count(select %s from %s where %s = $parent.%s)
	`, targetField, targetId, targetTable, idField, idField)
}
