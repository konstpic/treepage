package rag

import (
	"context"

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

	ids := make([]string, len(rows))
	for i, r := range rows {
		ids[i] = r.ChunkID
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

func (s *Service) vectorSearchChunks(ctx context.Context, question string, allowed []string, limit int) ([]chunkRow, error) {
	if s.embed == nil || !s.embed.Available() {
		return nil, nil
	}
	qEmb, err := s.embed.Embed(ctx, question)
	if err != nil || len(qEmb) == 0 {
		return nil, err
	}

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

	type rowWithEmb struct {
		chunkRow
		Embedding embeddings.Vector `gorm:"column:embedding"`
	}
	var raw []rowWithEmb
	if err := q.Limit(limit * 15).Scan(&raw).Error; err != nil {
		return nil, err
	}

	type scored struct {
		row chunkRow
		sim float64
	}
	scoredRows := make([]scored, 0, len(raw))
	for _, r := range raw {
		sim := embeddings.CosineSimilarity(qEmb, r.Embedding)
		if sim < 0.25 {
			continue
		}
		r.chunkRow.Rank = sim
		r.chunkRow.VectorSim = sim
		scoredRows = append(scoredRows, scored{row: r.chunkRow, sim: sim})
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
