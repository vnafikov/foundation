package builders

import (
	"strconv"
	"strings"

	"github.com/vnafikov/foundation/core/parameters/page/filtering"
)

const (
	approxExpressionLen  = 18
	approxPlaceholderLen = 5
)

/*
BuildWhereCondition формирует условие для WHERE.

IMPORTANT: Требует наличие pg_trgm, COLLATION und_ai_ci, FUNCTION escape_like_text.
*/
func BuildWhereCondition(conditions []filtering.Condition, lastUsedPlaceholderPosition int) (whereCondition, and string, values []any) {
	conditionCount := len(conditions)
	var conditionBuilder strings.Builder
	conditionBuilder.Grow(conditionCount * approxExpressionLen)
	values = make([]any, lastUsedPlaceholderPosition, lastUsedPlaceholderPosition+conditionCount)
	for _, condition := range conditions {
		valueCount := len(condition.Values)
		if valueCount == 0 {
			continue
		}

		if conditionBuilder.Len() > 0 {
			conditionBuilder.WriteString(" AND ")
		}
		value := condition.Values[0]
		column := condition.Field.Column
		switch condition.Operator {
		case filtering.Equal:
			if value == nil {
				conditionBuilder.WriteString(column + " IS NULL")
				break
			}

			lastUsedPlaceholderPosition++
			conditionBuilder.WriteString(column + " = $" + strconv.Itoa(lastUsedPlaceholderPosition))
			values = append(values, value)
		case filtering.NotEqual:
			if value == nil {
				conditionBuilder.WriteString(column + " IS NOT NULL")
				break
			}

			lastUsedPlaceholderPosition++
			conditionBuilder.WriteString(column + " != $" + strconv.Itoa(lastUsedPlaceholderPosition))
			values = append(values, value)
		case filtering.In, filtering.NotIn:
			var (
				operator               string
				connective             string
				preparedNullExpression string
			)
			switch condition.Operator {
			case filtering.In:
				operator = " IN "
				connective = " OR "
				preparedNullExpression = column + " IS NULL"
			case filtering.NotIn:
				operator = " NOT IN "
				connective = " AND "
				preparedNullExpression = column + " IS NOT NULL"
			}
			nullExpression := ""
			var placeholdersBuilder strings.Builder
			placeholdersBuilder.Grow(valueCount * approxPlaceholderLen)
			for _, value := range condition.Values {
				if value == nil {
					nullExpression = preparedNullExpression
					continue
				}

				if placeholdersBuilder.Len() > 0 {
					placeholdersBuilder.WriteString(", ")
				}
				lastUsedPlaceholderPosition++
				placeholdersBuilder.WriteString("$" + strconv.Itoa(lastUsedPlaceholderPosition))
				values = append(values, value)
			}
			inExpression := ""
			if placeholdersBuilder.Len() > 0 {
				inExpression = column + operator + "(" + placeholdersBuilder.String() + ")"
			}
			conditionBuilder.WriteString("(" + inExpression)
			if inExpression != "" && nullExpression != "" {
				conditionBuilder.WriteString(connective)
			}
			conditionBuilder.WriteString(nullExpression + ")")
		case filtering.Substring:
			lastUsedPlaceholderPosition++
			position := strconv.Itoa(lastUsedPlaceholderPosition)
			conditionBuilder.WriteString(column + " COLLATE und_ai_ci LIKE '%' || escape_like_text($" + position + ") || '%'")
			values = append(values, value)
		default:
			var operator string
			switch condition.Operator {
			case filtering.GreaterThan:
				operator = " > "
			case filtering.GreaterThanOrEqual:
				operator = " >= "
			case filtering.LessThan:
				operator = " < "
			case filtering.LessThanOrEqual:
				operator = " <= "
			case filtering.Fuzzy:
				operator = " %> "
			}
			lastUsedPlaceholderPosition++
			conditionBuilder.WriteString(column + operator + "$" + strconv.Itoa(lastUsedPlaceholderPosition))
			values = append(values, value)
		}
	}
	if conditionBuilder.Len() > 0 {
		and = " AND "
	}
	return conditionBuilder.String(), and, values
}

func BuildWhereClause(conditions []filtering.Condition, lastUsedPlaceholderPosition int) (clause string, values []any) {
	whereCondition, _, values := BuildWhereCondition(conditions, lastUsedPlaceholderPosition)
	if whereCondition == "" {
		return "", values
	}

	clause = `
WHERE ` + whereCondition
	return clause, values
}
