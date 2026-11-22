package meta

/*
Lookup_ содержит метаданные модели для поиска.

_ в Lookup_ уменьшает вероятность совпадения названия с полем модели при встраивании.
Содержит срез и карты с [Field] по значению для неизменяемости.
При копировании [Field] строки копируют только заголовки.
*/
type Lookup_ struct {
	fields     []Field
	byField    map[string]Field
	byColumn   map[string]Field
	byProperty map[string]Field
}

func (lookup *Lookup_) Fields() []Field {
	return lookup.fields
}

func (lookup *Lookup_) ByField(field string) (Field, bool) {
	fieldMeta, ok := lookup.byField[field]
	return fieldMeta, ok
}

func (lookup *Lookup_) ByColumn(column string) (Field, bool) {
	fieldMeta, ok := lookup.byColumn[column]
	return fieldMeta, ok
}

func (lookup *Lookup_) ByProperty(property string) (Field, bool) {
	fieldMeta, ok := lookup.byProperty[property]
	return fieldMeta, ok
}

func NewLookup(fields []Field) *Lookup_ {
	fieldCount := len(fields)
	lookup := &Lookup_{
		fields:     fields,
		byField:    make(map[string]Field, fieldCount),
		byColumn:   make(map[string]Field, fieldCount),
		byProperty: make(map[string]Field, fieldCount),
	}
	for _, fieldMeta := range lookup.fields {
		if fieldMeta.Field != "" {
			lookup.byField[fieldMeta.Field] = fieldMeta
		}
		if fieldMeta.Column != "" {
			lookup.byColumn[fieldMeta.Column] = fieldMeta
		}
		if fieldMeta.Property != "" {
			lookup.byProperty[fieldMeta.Property] = fieldMeta
		}
	}
	return lookup
}
