package builders

import (
	"strconv"
	"strings"

	"github.com/vnafikov/common/types/nullable"
)

const (
	approxColumnLen      = 12
	approxPlaceholderLen = 5

	approxExpressionLen = 18

	approxUpsertExpressionLen = 32
)

func BuildInsertData(attrs map[string]any, lastUsedPlaceholderPosition int) (columns, placeholders string, values []any) {
	attrCount := len(attrs)
	var (
		columnsBuilder      strings.Builder
		placeholdersBuilder strings.Builder
	)
	columnsBuilder.Grow(attrCount * approxColumnLen)
	placeholdersBuilder.Grow(attrCount * approxPlaceholderLen)
	values = make([]any, lastUsedPlaceholderPosition, lastUsedPlaceholderPosition+attrCount)
	for column, value := range attrs {
		if nullable.IsNull(value) {
			continue
		}

		if columnsBuilder.Len() > 0 {
			columnsBuilder.WriteString(", ")
			placeholdersBuilder.WriteString(", ")
		}
		columnsBuilder.WriteString(column)
		lastUsedPlaceholderPosition++
		placeholdersBuilder.WriteString("$" + strconv.Itoa(lastUsedPlaceholderPosition))
		values = append(values, value)
	}
	return columnsBuilder.String(), placeholdersBuilder.String(), values
}

func BuildUpdateData(attrs map[string]any, lastUsedPlaceholderPosition int) (expressions string, values []any) {
	attrCount := len(attrs)
	var expressionsBuilder strings.Builder
	expressionsBuilder.Grow(attrCount * approxExpressionLen)
	values = make([]any, lastUsedPlaceholderPosition, lastUsedPlaceholderPosition+attrCount)
	for column, value := range attrs {
		isExplicitNull := nullable.IsExplicitNull(value)
		if nullable.IsNull(value) && !isExplicitNull {
			continue
		}

		if expressionsBuilder.Len() > 0 {
			expressionsBuilder.WriteString(", ")
		}
		if isExplicitNull {
			expressionsBuilder.WriteString(column + " = NULL")
			continue
		}

		lastUsedPlaceholderPosition++
		expressionsBuilder.WriteString(column + " = $" + strconv.Itoa(lastUsedPlaceholderPosition))
		values = append(values, value)
	}
	return expressionsBuilder.String(), values
}

func BuildUpsertData(attrs map[string]any, lastUsedPlaceholderPosition int) (columns, placeholders, expressions string, values []any) {
	attrCount := len(attrs)
	var (
		columnsBuilder      strings.Builder
		placeholdersBuilder strings.Builder
		expressionsBuilder  strings.Builder
	)
	columnsBuilder.Grow(attrCount * approxColumnLen)
	placeholdersBuilder.Grow(attrCount * approxPlaceholderLen)
	expressionsBuilder.Grow(attrCount * approxUpsertExpressionLen)
	values = make([]any, lastUsedPlaceholderPosition, lastUsedPlaceholderPosition+attrCount)
	for column, value := range attrs {
		isExplicitNull := nullable.IsExplicitNull(value)
		if nullable.IsNull(value) && !isExplicitNull {
			continue
		}

		if columnsBuilder.Len() > 0 {
			columnsBuilder.WriteString(", ")
			placeholdersBuilder.WriteString(", ")
			expressionsBuilder.WriteString(", ")
		}
		columnsBuilder.WriteString(column)
		expressionsBuilder.WriteString(column + " = excluded." + column)
		if isExplicitNull {
			placeholdersBuilder.WriteString("NULL")
			continue
		}

		lastUsedPlaceholderPosition++
		placeholdersBuilder.WriteString("$" + strconv.Itoa(lastUsedPlaceholderPosition))
		values = append(values, value)
	}
	return columnsBuilder.String(), placeholdersBuilder.String(), expressionsBuilder.String(), values
}
