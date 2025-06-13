package storage

import "context"

// Store is an interface for a generic blob storage system.
// It abstracts the underlying implementation (e.g., local disk, S3, Azure Blob).
type Store interface {
	// Save writes the given content to a specified path.
	// The path is typically structured as: <intentID>/<artifactID>/<filename>
	Save(ctx context.Context, path string, content []byte) (string, error)

	// Get retrieves content from a specified path.
	Get(ctx context.Context, path string) ([]byte, error)

	// GetURL returns a publicly accessible URL for the given path, if applicable.
	// For local storage, this might be a file URI. For cloud storage, a signed URL.
	GetURL(ctx context.Context, path string) (string, error)
}
