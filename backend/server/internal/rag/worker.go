package rag

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// WorkerStatus describes background RAG indexing progress.
type WorkerStatus struct {
	Phase               string     `json:"phase"`
	Running             bool       `json:"running"`
	StartedAt           *time.Time `json:"started_at,omitempty"`
	UpdatedAt           *time.Time `json:"updated_at,omitempty"`
	DocumentsTotal      int        `json:"documents_total"`
	DocumentsDone       int        `json:"documents_done"`
	DocumentsWithChunks int        `json:"documents_with_chunks"`
	PublishedDocuments  int        `json:"published_documents"`
	ChunksTotal         int64      `json:"chunks_total"`
	ChunksEmbedded      int64      `json:"chunks_embedded"`
	ChunksEmbeddedRun   int        `json:"chunks_embedded_run"`
	ChunksPending       int64      `json:"chunks_pending"`
	EmbeddingsEnabled   bool       `json:"embeddings_enabled"`
	PgVectorEnabled     bool       `json:"pgvector_enabled"`
	Error               string     `json:"error,omitempty"`
}

// Worker runs RAG backfill asynchronously without blocking server readiness.
type Worker struct {
	svc    *Service
	db     *gorm.DB
	logger *zap.Logger

	mu     sync.RWMutex
	status WorkerStatus
}

func NewWorker(svc *Service, db *gorm.DB, logger *zap.Logger) *Worker {
	return &Worker{
		svc: svc,
		db:  db,
		logger: logger,
		status: WorkerStatus{Phase: "idle", PgVectorEnabled: PgVectorEnabled(db)},
	}
}

func PgVectorEnabled(db *gorm.DB) bool {
	var ok bool
	_ = db.Raw(`SELECT EXISTS(SELECT 1 FROM pg_extension WHERE extname = 'vector')`).Scan(&ok).Error
	return ok
}

func (w *Worker) Status() WorkerStatus {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.status
}

func (w *Worker) setStatus(update func(*WorkerStatus)) {
	w.mu.Lock()
	defer w.mu.Unlock()
	update(&w.status)
	now := time.Now()
	w.status.UpdatedAt = &now
}

// Start launches background indexing; safe to call once at startup.
func (w *Worker) Start(ctx context.Context) {
	go w.run(ctx)
}

func (w *Worker) refreshPersistedStats(ctx context.Context) {
	stats, err := w.svc.IndexStats(ctx)
	if err != nil {
		return
	}
	w.setStatus(func(s *WorkerStatus) {
		s.PublishedDocuments = stats.PublishedDocuments
		s.DocumentsWithChunks = stats.DocumentsWithChunks
		s.ChunksTotal = stats.ChunksTotal
		s.ChunksEmbedded = stats.ChunksEmbedded
		s.ChunksPending = stats.ChunksPending
		s.EmbeddingsEnabled = stats.EmbeddingsEnabled
		s.PgVectorEnabled = PgVectorEnabled(w.db)
	})
}

func (w *Worker) run(ctx context.Context) {
	w.setStatus(func(s *WorkerStatus) {
		s.Phase = "starting"
		s.Running = true
		now := time.Now()
		s.StartedAt = &now
		s.Error = ""
	})
	w.refreshPersistedStats(ctx)

	var chunkCount int64
	if err := w.db.WithContext(ctx).Table("document_chunks").Count(&chunkCount).Error; err != nil {
		w.fail(err)
		return
	}

	needsReindex := chunkCount == 0
	if !needsReindex {
		stats, _ := w.svc.IndexStats(ctx)
		needsReindex = stats.PublishedDocuments > stats.DocumentsWithChunks
	}

	if needsReindex {
		w.setStatus(func(s *WorkerStatus) { s.Phase = "reindex" })
		n, err := w.svc.ReindexAllPublished(ctx)
		w.setStatus(func(s *WorkerStatus) {
			s.DocumentsTotal = n
			s.DocumentsDone = n
		})
		if err != nil {
			w.fail(err)
			return
		}
		if w.logger != nil {
			w.logger.Info("rag worker reindex completed", zap.Int("documents", n))
		}
		w.refreshPersistedStats(ctx)
	}

	if w.svc.embed != nil && w.svc.embed.Available() {
		w.setStatus(func(s *WorkerStatus) { s.Phase = "embeddings" })
		for {
			select {
			case <-ctx.Done():
				w.setStatus(func(s *WorkerStatus) {
					s.Phase = "stopped"
					s.Running = false
				})
				return
			default:
			}
			n, err := w.svc.backfillEmbeddings(ctx, 20)
			w.setStatus(func(s *WorkerStatus) { s.ChunksEmbeddedRun += n })
			if err != nil {
				w.fail(err)
				return
			}
			if n == 0 {
				break
			}
		}
		if PgVectorEnabled(w.db) {
			w.setStatus(func(s *WorkerStatus) { s.Phase = "pgvector" })
			for {
				n, err := w.svc.backfillPgVectors(ctx, 50)
				if err != nil {
					w.fail(err)
					return
				}
				if n == 0 {
					break
				}
			}
		}
	}

	var pending int64
	_ = w.db.WithContext(ctx).Table("document_chunks").Where("embedding IS NULL").Count(&pending).Error
	w.refreshPersistedStats(ctx)
	w.setStatus(func(s *WorkerStatus) {
		s.ChunksPending = pending
		s.Phase = "done"
		s.Running = false
	})
	if w.logger != nil {
		w.logger.Info("rag worker finished", zap.Int64("chunks_pending", pending))
	}
}

func (w *Worker) fail(err error) {
	if w.logger != nil {
		w.logger.Warn("rag worker failed", zap.Error(err))
	}
	w.setStatus(func(s *WorkerStatus) {
		s.Phase = "error"
		s.Running = false
		s.Error = err.Error()
	})
}

// TriggerReindex starts a full reindex in the background (admin action).
func (w *Worker) TriggerReindex(ctx context.Context) {
	go func() {
		w.setStatus(func(s *WorkerStatus) {
			s.Phase = "reindex"
			s.Running = true
			s.DocumentsDone = 0
			s.ChunksEmbeddedRun = 0
			s.Error = ""
			now := time.Now()
			s.StartedAt = &now
		})
		n, err := w.svc.ReindexAllPublished(ctx)
		w.setStatus(func(s *WorkerStatus) {
			s.DocumentsTotal = n
			s.DocumentsDone = n
		})
		if err != nil {
			w.fail(err)
			return
		}
		w.refreshPersistedStats(ctx)

		if w.svc.embed != nil && w.svc.embed.Available() {
			w.setStatus(func(s *WorkerStatus) { s.Phase = "embeddings" })
			for {
				embedded, err := w.svc.backfillEmbeddings(ctx, 20)
				w.setStatus(func(s *WorkerStatus) { s.ChunksEmbeddedRun += embedded })
				if err != nil {
					w.fail(err)
					return
				}
				if embedded == 0 {
					break
				}
			}
		}
		w.refreshPersistedStats(ctx)
		w.setStatus(func(s *WorkerStatus) {
			s.Phase = "done"
			s.Running = false
		})
	}()
}
