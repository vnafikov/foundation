package parameters

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/vnafikov/common/num"

	"github.com/vnafikov/foundation/core/models/meta"
	"github.com/vnafikov/foundation/core/parameters/page"
	"github.com/vnafikov/foundation/core/parameters/page/filtering"
	"github.com/vnafikov/foundation/core/parameters/page/pagination"
	"github.com/vnafikov/foundation/core/parameters/page/sorting"
)

const (
	filterPrefix    = "filter-"
	filterPrefixLen = len(filterPrefix)
)

var searches = map[string]filtering.SearchOperator{
	"search-sub":   filtering.SearchSubstring,
	"search-fuzzy": filtering.SearchFuzzy,
}

func ReadPageByCursorQueryParameters(r *http.Request, lookupMeta *meta.Lookup_, options ReadPageOptions, p *page.Parameters) {
	values := r.URL.Query()
	p.Pagination.Type = pagination.TypeCursor
	readCursorPagination(values, p)
	readFiltering(values, lookupMeta, options.Allowed.Filtering, &p.Filtering)
	readSorting(values, lookupMeta, options.Allowed.Sorting, p)
}

func readCursorPagination(values url.Values, p *page.Parameters) {
	p.Pagination.Cursor.Next = values.Get("next")
	p.Pagination.Cursor.Previous = values.Get("previous")
	p.Pagination.Cursor.IsLastPage = values.Has("last")
	p.Pagination.Cursor.Limit = num.ParseIntOrZero(values.Get("limit"))
}

func readFiltering(values url.Values, lookupMeta *meta.Lookup_, allowed filtering.Allowed, p *filtering.Parameters) {
	readAllConditions(values, lookupMeta, allowed, p)
	readSearch(values, allowed.Search, p)
}

func readAllConditions(values url.Values, lookupMeta *meta.Lookup_, allowed filtering.Allowed, p *filtering.Parameters) {
	var (
		conditions       []filtering.Condition
		customConditions []filtering.CustomCondition
	)
	for key, list := range values {
		if !strings.HasPrefix(key, filterPrefix) {
			continue
		}

		key = key[filterPrefixLen:]
		property, operator, found := strings.Cut(key, "-")
		if !found {
			name := filtering.CustomConditionName(key)
			if customCondition, ok := filtering.NewCustomConditionFromString(name, list, allowed.CustomConditions); ok {
				customConditions = append(customConditions, customCondition)
			}
			continue
		}

		fieldMeta, ok := lookupMeta.ByProperty(property)
		if !ok {
			continue
		}

		if condition, ok := filtering.NewConditionFromString(fieldMeta, operator, list, allowed.Conditions); ok {
			conditions = append(conditions, condition)
		}
	}
	p.Conditions = conditions
	p.CustomConditions = customConditions
}

func readSearch(values url.Values, allowed filtering.AllowedSearch, p *filtering.Parameters) {
	for key, operator := range searches {
		value := values.Get(key)
		if value == "" {
			continue
		}

		if search, ok := filtering.NewSearch(operator, value, allowed); ok {
			p.Search = search
			return
		}
	}
}

func readSorting(values url.Values, lookupMeta *meta.Lookup_, allowed sorting.Allowed, p *page.Parameters) {
	value := values.Get("sort")
	if value == "" {
		return
	}

	sortValues := strings.Split(value, ",")
	valueCount := len(sortValues)
	seen := make(map[meta.Field]struct{}, valueCount)
	sortingParameters := make(sorting.Parameters, 0, valueCount)
	for _, sortValue := range sortValues {
		direction, sortValue := sortingDirection(sortValue)
		fieldMeta, ok := lookupMeta.ByProperty(sortValue)
		if !ok {
			continue
		}

		if sort, ok := sorting.NewSort(fieldMeta, direction, allowed); ok {
			if _, ok := seen[fieldMeta]; ok {
				continue
			}

			seen[fieldMeta] = struct{}{}
			sortingParameters = append(sortingParameters, sort)
		}
	}
	p.Sorting = sortingParameters
}

func sortingDirection(value string) (direction sorting.Direction, sortValue string) {
	if strings.HasPrefix(value, "-") {
		return sorting.Desc, value[1:]
	}
	return sorting.Asc, value
}

func ReadPageByOffsetQueryParameters(r *http.Request, lookupMeta *meta.Lookup_, options ReadPageOptions, p *page.Parameters) {
	values := r.URL.Query()
	p.Pagination.Type = pagination.TypeOffset
	readOffsetPagination(values, p)
	readFiltering(values, lookupMeta, options.Allowed.Filtering, &p.Filtering)
	readSorting(values, lookupMeta, options.Allowed.Sorting, p)
}

func readOffsetPagination(values url.Values, p *page.Parameters) {
	p.Pagination.Offset.Page = num.ParseIntOrZero(values.Get("page"))
	p.Pagination.Offset.Limit = num.ParseIntOrZero(values.Get("limit"))
}
