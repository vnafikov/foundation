package queries

import "github.com/vnafikov/foundation/core/models/meta"

type metaProvider interface {
	LookupMeta() *meta.Lookup_
}
