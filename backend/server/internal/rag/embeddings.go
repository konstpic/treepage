package rag

import (
	"context"
	"encoding/json"

	"github.com/konstpic/treepage/backend/pkg/embeddings"
	"github.com/konstpic/treepage/backend/pkg/models"
)

func (s *Service) rerankWithEmbeddings(ctx context.Context, rows []chunkRow, question string, _ []string) []chunkRow {
	if s.embed == nil || !s.embed.Available() || len(rows) == 0 {
		return rows
	}
	qEmb, err := s.embed.Embed(ctx, question)
	if err != nil || len(qEmb) == 0 {
		return rows
	}

	ids := make([]string, 0, len(rows))
	for _, r := range rows {
		if r.ChunkID != "" {
			ids = append(ids, r.ChunkID)
		}
	}
	if len(ids) == 0 {
		return rows
	}
	var dbChunks []models.DocumentChunk
	if err := s.db.WithContext(ctx).Where("id IN ?", ids).Find(&dbChunks).Error; err != nil {
		return rows
	}
	embByID := map[string]embeddings.Vector{}
	for _, c := range dbChunks {
		if len(c.Embedding) > 0 {
			embByID[c.ID] = c.Embedding
		}
	}

	maxFTS := rows[0].Rank
	if maxFTS <= 0 {
		maxFTS = 1
	}
	for i := range rows {
		sim := 0.0
		if emb, ok := embByID[rows[i].ChunkID]; ok {
			sim = embeddings.CosineSimilarity(qEmb, emb)
			rows[i].VectorSim = sim
		}
		ftsNorm := rows[i].Rank / maxFTS
		if sim > 0 {
			rows[i].Rank = 0.4*ftsNorm + 0.6*sim
		}
	}
	for i := 1; i < len(rows); i++ {
		j := i
		for j > 0 && rows[j].Rank > rows[j-1].Rank {
			rows[j], rows[j-1] = rows[j-1], rows[j]
			j--
		}
	}
	return rows
}

func (s *Service) backfillEmbeddings(ctx context.Context, limit int) (int, error) {
	if s.embed == nil || !s.embed.Available() {
		return 0, nil
	}
	var chunks []models.DocumentChunk
	if err := s.db.WithContext(ctx).
		Where("embedding IS NULL").
		Order("document_id, chunk_index").
		Limit(limit).
		Find(&chunks).Error; err != nil {
		return 0, err
	}
	n := 0
	for _, c := range chunks {
		emb, err := s.embed.Embed(ctx, c.Content)
		if err != nil || len(emb) == 0 {
			continue
		}
		if err := s.db.WithContext(ctx).Model(&models.DocumentChunk{}).
			Where("id = ?", c.ID).
			Update("embedding", emb).Error; err != nil {
			return n, err
		}
		_ = embeddings.SetChunkVector(ctx, s.db, c.ID, emb)
		n++
	}
	return n, nil
}

func (s *Service) BackfillEmbeddings(ctx context.Context) (int, error) {
	total := 0
	for {
		n, err := s.backfillEmbeddings(ctx, 20)
		total += n
		if err != nil {
			return total, err
		}
		if n == 0 {
			break
		}
	}
	return total, nil
}

func (s *Service) backfillPgVectors(ctx context.Context, limit int) (int, error) {
	return embeddings.BackfillVectorsFromJSONB(ctx, s.db, limit)
}

func (s *Service) vectorSearchChunks(ctx context.Context, question string, allowed []string, limit int) ([]chunkRow, error) {
	if s.embed == nil || !s.embed.Available() {
		return nil, nil
	}
	qEmb, err := s.embed.Embed(ctx, question)
	if err != nil || len(qEmb) == 0 {
		return nil, err
	}

	if embeddings.PgVectorAvailable(s.db) {
		return s.vectorSearchPgVector(ctx, qEmb, allowed, limit)
	}
	return s.vectorSearchJSONB(ctx, qEmb, allowed, limit)
}

func (s *Service) vectorSearchPgVector(ctx context.Context, qEmb embeddings.Vector, allowed []string, limit int) ([]chunkRow, error) {
	lit := embeddings.VectorLiteral(qEmb)
	spaceFilter := ""
	args := []any{lit}
	if allowed != nil {
		spaceFilter = " AND d.space_id IN ?"
		args = append(args, allowed)
	}
	args = append(args, lit, limit*3)
	query := `
		SELECT dc.id AS chunk_id, dc.document_id, dc.content, d.title, d.slug, sp.slug AS space_slug,
			d.space_id, d.path,
			(1 - (dc.embedding_vector <=> ?::vector)) AS rank
		FROM document_chunks dc
		JOIN documents d ON d.id = dc.document_id
		JOIN spaces sp ON sp.id = d.space_id
		WHERE d.is_published = true
		  AND dc.embedding_vector IS NOT NULL` + spaceFilter + `
		ORDER BY dc.embedding_vector <=> ?::vector
		LIMIT ?`

	var rows []chunkRow
	if err := s.db.WithContext(ctx).Raw(query, args...).Scan(&rows).Error; err != nil {
		return nil, err
	}
	for i := range rows {
		rows[i].VectorSim = rows[i].Rank
	}
	return rows, nil
}

func (s *Service) vectorSearchJSONB(ctx context.Context, qEmb embeddings.Vector, allowed []string, limit int) ([]chunkRow, error) {
	q := s.db.WithContext(ctx).Table("document_chunks dc").
		Select(`dc.id AS chunk_id, dc.document_id, dc.content, d.title, d.slug, sp.slug AS space_slug,
			d.space_id, d.path, dc.embedding`).
		Joins("JOIN documents d ON d.id = dc.document_id").
		Joins("JOIN spaces sp ON sp.id = d.space_id").
		Where("d.is_published = ?", true).
		Where("dc.embedding IS NOT NULL")
	if allowed != nil {
		q = q.Where("d.space_id IN ?", allowed)
	}

	type vectorChunkRow struct {
		ChunkID    string `gorm:"column:chunk_id"`
		DocumentID string `gorm:"column:document_id"`
		Content    string `gorm:"column:content"`
		Title      string `gorm:"column:title"`
		Slug       string `gorm:"column:slug"`
		SpaceSlug  string `gorm:"column:space_slug"`
		SpaceID    string `gorm:"column:space_id"`
		Path       string `gorm:"column:path"`
		Embedding  []byte `gorm:"column:embedding"`
	}
	var raw []vectorChunkRow
	if err := q.Limit(limit * 15).Scan(&raw).Error; err != nil {
		return nil, err
	}

	type scored struct {
		row chunkRow
		sim float64
	}
	scoredRows := make([]scored, 0, len(raw))
	for _, r := range raw {
		if r.ChunkID == "" {
			continue
		}
		var emb embeddings.Vector
		if len(r.Embedding) > 0 {
			if err := json.Unmarshal(r.Embedding, &emb); err != nil {
				continue
			}
		}
		if len(emb) == 0 {
			continue
		}
		sim := embeddings.CosineSimilarity(qEmb, emb)
		if sim < 0.25 {
			continue
		}
		row := chunkRow{
			ChunkID:    r.ChunkID,
			DocumentID: r.DocumentID,
			Content:    r.Content,
			Title:      r.Title,
			Slug:       r.Slug,
			SpaceSlug:  r.SpaceSlug,
			SpaceID:    r.SpaceID,
			Path:       r.Path,
			Rank:       sim,
			VectorSim:  sim,
		}
		scoredRows = append(scoredRows, scored{row: row, sim: sim})
	}
	for i := 1; i < len(scoredRows); i++ {
		j := i
		for j > 0 && scoredRows[j].sim > scoredRows[j-1].sim {
			scoredRows[j], scoredRows[j-1] = scoredRows[j-1], scoredRows[j]
			j--
		}
	}
	out := make([]chunkRow, 0, limit*3)
	for _, sr := range scoredRows {
		out = append(out, sr.row)
		if len(out) >= limit*3 {
			break
		}
	}
	return out, nil
}

func mergeChunkRows(primary, secondary []chunkRow) []chunkRow {
	merged := map[string]chunkRow{}
	order := make([]string, 0, len(primary)+len(secondary))
	add := func(r chunkRow) {
		if r.ChunkID == "" {
			return
		}
		if prev, ok := merged[r.ChunkID]; ok {
			if r.Rank > prev.Rank {
				merged[r.ChunkID] = r
			}
			return
		}
		merged[r.ChunkID] = r
		order = append(order, r.ChunkID)
	}
	for _, r := range primary {
		add(r)
	}
	for _, r := range secondary {
		add(r)
	}
	out := make([]chunkRow, 0, len(order))
	for _, id := range order {
		out = append(out, merged[id])
	}
	return out
}
