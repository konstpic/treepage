package rag

import "testing"

func TestScoreChunkNavigationWinsOverEditing(t *testing.T) {
	keywords := []string{"раздел", "страниц"}
	nav := chunkRow{
		Title:   "Навигация по интерфейсу",
		Path:    "ru/user/navigation.md",
		Content: "### Боковая панель\n\n- **Страницы** — дерево документов (вкладка по умолчанию)",
		Rank:    1.0,
	}
	edit := chunkRow{
		Title:   "Редактирование документов",
		Path:    "ru/user/editing-docs.md",
		Content: "Страница связана с Git. Изменения сохраняются в TreePage — создайте PR в репозитории.",
		Rank:    1.0,
	}
	navScore := scoreChunk(nav, keywords)
	editScore := scoreChunk(edit, keywords)
	if navScore <= editScore {
		t.Fatalf("navigation should outrank editing: nav=%v edit=%v", navScore, editScore)
	}
}

func TestRankChunksOrdersByScore(t *testing.T) {
	keywords := []string{"раздел", "страниц"}
	rows := rankChunks([]chunkRow{
		{DocumentID: "a", Title: "Редактирование", Path: "editing.md", Content: "страница git", Rank: 2},
		{DocumentID: "b", Title: "Навигация", Path: "ru/user/navigation.md", Content: "**Страницы** — дерево", Rank: 1},
	}, keywords)
	if rows[0].DocumentID != "b" {
		t.Fatalf("expected navigation first, got %s (scores: %v, %v)", rows[0].Title, rows[0].Rank, rows[1].Rank)
	}
}
