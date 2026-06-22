import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Loader2, RefreshCw, Sparkles } from "lucide-react";
import { api } from "@/lib/api";
import { useI18n } from "@/lib/i18n";
import { useAdminGuard } from "./layout";

interface RAGStatus {
  phase: string;
  running: boolean;
  documents_total: number;
  documents_done: number;
  chunks_embedded: number;
  chunks_pending: number;
  pgvector_enabled: boolean;
  error?: string;
}

export function AdminRAGPage() {
  const { t } = useI18n();
  const { ready } = useAdminGuard();
  const qc = useQueryClient();

  const { data, isLoading, refetch, isFetching } = useQuery({
    queryKey: ["admin-rag-status"],
    queryFn: () => api<RAGStatus>("/api/admin/rag/status"),
    enabled: ready,
    refetchInterval: (q) => (q.state.data?.running ? 3000 : false),
  });

  const reindex = useMutation({
    mutationFn: () => api("/api/admin/rag/reindex", { method: "POST" }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["admin-rag-status"] }),
  });

  if (!ready) return null;

  return (
    <div className="glass p-6">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <h2 className="flex items-center gap-2 text-xl font-semibold text-fg">
          <Sparkles className="h-5 w-5 text-primary" />
          {t("admin.rag.title")}
        </h2>
        <div className="flex gap-2">
          <button type="button" className="btn-secondary" onClick={() => refetch()} disabled={isFetching}>
            {isFetching ? <Loader2 className="h-4 w-4 animate-spin" /> : <RefreshCw className="h-4 w-4" />}
          </button>
          <button
            type="button"
            className="btn-primary"
            disabled={reindex.isPending || data?.running}
            onClick={() => reindex.mutate()}
          >
            {reindex.isPending ? t("common.loading") : t("admin.rag.reindex")}
          </button>
        </div>
      </div>

      {isLoading ? (
        <div className="mt-6 flex items-center gap-2 text-subtle">
          <Loader2 className="h-4 w-4 animate-spin" />
          {t("common.loading")}
        </div>
      ) : data ? (
        <dl className="mt-6 grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          <div className="rounded-xl border border-default bg-surface-muted p-4">
            <dt className="text-xs text-subtle">{t("admin.rag.phase")}</dt>
            <dd className="mt-1 font-medium text-fg">{data.phase}</dd>
          </div>
          <div className="rounded-xl border border-default bg-surface-muted p-4">
            <dt className="text-xs text-subtle">{t("admin.rag.documents")}</dt>
            <dd className="mt-1 font-medium text-fg">
              {data.documents_done}/{data.documents_total || "—"}
            </dd>
          </div>
          <div className="rounded-xl border border-default bg-surface-muted p-4">
            <dt className="text-xs text-subtle">{t("admin.rag.chunksPending")}</dt>
            <dd className="mt-1 font-medium text-fg">{data.chunks_pending}</dd>
          </div>
          <div className="rounded-xl border border-default bg-surface-muted p-4">
            <dt className="text-xs text-subtle">{t("admin.rag.chunksEmbedded")}</dt>
            <dd className="mt-1 font-medium text-fg">{data.chunks_embedded}</dd>
          </div>
          <div className="rounded-xl border border-default bg-surface-muted p-4">
            <dt className="text-xs text-subtle">pgvector</dt>
            <dd className="mt-1 font-medium text-fg">{data.pgvector_enabled ? "on" : "off"}</dd>
          </div>
          {data.error && (
            <div className="rounded-xl border border-danger/30 bg-danger/10 p-4 sm:col-span-2 lg:col-span-3">
              <dt className="text-xs text-subtle">{t("admin.rag.error")}</dt>
              <dd className="mt-1 text-sm text-fg">{data.error}</dd>
            </div>
          )}
        </dl>
      ) : null}
    </div>
  );
}
