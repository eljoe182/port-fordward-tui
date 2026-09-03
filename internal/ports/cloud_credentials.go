package ports

import "context"

// CredentialSyncFailure records a single cluster that failed during credential sync.
type CredentialSyncFailure struct {
	Name   string
	Detail string
}

// CredentialSyncResult summarizes a cloud credential sync run.
type CredentialSyncResult struct {
	Provider string
	Synced   int
	Failed   []CredentialSyncFailure
}

// CloudCredentialSyncer refreshes kubeconfig entries from a cloud provider.
type CloudCredentialSyncer interface {
	Provider() string
	Sync(ctx context.Context) (CredentialSyncResult, error)
}
