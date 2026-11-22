package parameters

import (
	"net/http"

	"github.com/vnafikov/foundation/core/models/meta"
	"github.com/vnafikov/foundation/core/parameters/page/filtering"
)

func ReadFilteringQueryParameters(r *http.Request, lookupMeta *meta.Lookup_, options ReadFilteringOptions, p *filtering.Parameters) {
	values := r.URL.Query()
	readFiltering(values, lookupMeta, options.Allowed, p)
}
