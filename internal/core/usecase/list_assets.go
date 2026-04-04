package usecase

import "github.com/ZouZhao321/distill/internal/core/port"

// ListItem 表示资产列表中的一个摘要条目。
type ListItem struct {
	Name      string
	Hash      string
	FileCount int
	TotalSize int64
	CreatedAt string
}

// ListAssetsUseCase 负责列出仓库中的所有资产。
type ListAssetsUseCase struct {
	repo port.AssetRepository
}

// NewListAssetsUseCase 创建新的 ListAssetsUseCase 实例。
func NewListAssetsUseCase(repo port.AssetRepository) *ListAssetsUseCase {
	return &ListAssetsUseCase{repo: repo}
}

// Execute 返回所有资产列表，自动跳过损坏的清单。
func (uc *ListAssetsUseCase) Execute() ([]ListItem, error) {
	refs, err := uc.repo.ListRefs()
	if err != nil {
		return nil, err
	}

	var items []ListItem
	for _, ref := range refs {
		m, err := uc.repo.GetManifest(ref.Manifest)
		if err != nil {
			continue // 跳过损坏的清单
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
