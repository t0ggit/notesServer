package list

type node struct {
	id       int64 // Уникальный идентификатор узла, может не совпадать с порядковым номером (индексом)
	value    any
	nextNode *node
}
