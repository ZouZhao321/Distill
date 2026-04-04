package usecase

import "github.com/ZouZhao321/distill/internal/core/port"

// ListItem represents a summary of an asset for listing.
type ListItem struct {
	Name      string
	Hash      string
	FileCount int
	TotalSize int64
	CreatedAt string
}

// ListAssetsUseCase lists all assets in the repository.
type ListAssetsUseCase struct {
	repo port.AssetRepository
}

// NewListAssetsUseCase creates a new ListAssetsUseCase.
func NewListAssetsUseCase(repo port.AssetRepository) *ListAssetsUseCase {
	return &ListAssetsUseCase{repo: repo}
}

// Execute returns all assets, skipping any with broken manifests.
func (uc *ListAssetsUseCase) Execute() ([]ListItem, error) {
	refs, err := uc.repo.ListRefs()
	if err != nil {
		return nil, err
	}

	var items []ListItem
	for _, ref := range refs {
		m, err := uc.repo.GetManifest(ref.Manifest)
		if err != nil {
			continue // skip broken manifests
		}
		items = append(items, ListItem{
			Name:      ref.Name,
			Hash:      m.Hash,
			FileCount: m.FileCount,
			TotalSize: m.TotalSize,
			CreatedAt: m.CreatedAt,
		})
	}
	return items, nil
}
