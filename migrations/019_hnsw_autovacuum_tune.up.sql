-- Reduce autovacuum churn on document_chunks (HNSW repair is expensive and can fail under load).
ALTER TABLE document_chunks SET (
  autovacuum_vacuum_scale_factor = 0.15,
  autovacuum_analyze_scale_factor = 0.1,
  autovacuum_vacuum_cost_delay = 5
);
