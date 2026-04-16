package tui

type CatalogItem struct {
	ID                 string
	Label              string
	RemotePort         int
	PreferredLocalPort int
}

type SelectedItem struct {
	TargetID   string
	Label      string
	LocalPort  int
	RemotePort int
}
