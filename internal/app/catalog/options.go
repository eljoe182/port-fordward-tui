package catalog

type FilterMode string

const (
	FilterAll       FilterMode = "all"
	FilterServices  FilterMode = "services"
	FilterPods      FilterMode = "pods"
	FilterFavorites FilterMode = "favorites"
)

type SortMode string

const (
	SortSmart     SortMode = "smart"
	SortName      SortMode = "name"
	SortRecent    SortMode = "recent"
	SortFavorites SortMode = "favorites"
	SortType      SortMode = "type"
)

type LoadOptions struct {
	Query  string
	Filter FilterMode
	Sort   SortMode
}

func (o LoadOptions) WithDefaults() LoadOptions {
	if o.Filter == "" {
		o.Filter = FilterAll
	}
	if o.Sort == "" {
		o.Sort = SortSmart
	}
	return o
}
