package pagination

const (
	DefaultLimit = 50
	MaxLimit     = 200
)

func NormalizeLimit(options NormalizeLimitOptions) int {
	if options.Limit <= 0 {
		if options.DefaultLimit == 0 {
			return DefaultLimit
		}
		return options.DefaultLimit
	}

	if options.MaxLimit == 0 {
		options.MaxLimit = MaxLimit
	}
	return min(options.Limit, options.MaxLimit)
}
