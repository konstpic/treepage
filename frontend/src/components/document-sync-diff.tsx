import { useQuery } from "@tanstack/react-query";
import { GitCompare, Loader2 } from "lucide-react";
import { useState } from "react";
import { api } from "@/lib/api";
import { useI18n } from "@/lib/i18n";

interface DiffLine {
  type: "add" | "remove" | "same";
  content: string;
}

interface SyncDiff {
  document_id: string;
  git_content: string;
  local_content: string;
  lines: DiffLine[];
}

interface DocumentSyncDiffProps {
  documentId: string;
}

export function DocumentSyncDiff({ documentId }: DocumentSyncDiffProps) {
  const { t } = useI18n();
  const [open, setOpen] = useState(false);

  const { data, isLoading, isFetching, refetch } = useQuery({
    queryKey: ["sync-diff", documentId],
    queryFn: () => api<SyncDiff>(`/api/documents/${documentId}/sync-diff`),
    enabled: open,
  });

  return (
    <div className="mb-4 rounded-xl border border-warning/30 bg-warning/10 px-4 py-3 text-sm text-fg">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <span>{t("document.pendingChanges")}</span>
        <button
          type="button"
          className="btn-ghost text-xs"
          onClick={() => {
            setOpen(true);
            void refetch();
          }}
        >
          <GitCompare className="mr-1 inline h-3.5 w-3.5" />
          {t("document.viewSyncDiff")}
        </button>
      </div>
      {open && (
        <div className="mt-3 max-h-64 overflow-auto rounded-lg border border-default bg-surface p-3 font-mono text-xs">
          {(isLoading || isFetching) && (
            <div className="flex items-center gap-2 text-subtle">
              <Loader2 className="h-4 w-4 animate-spin" />
              {t("common.loading")}
            </div>
          )}
          {data?.lines.map((line, i) => (
            <div
              key={i}
              className={
                line.type === "add"
                  ? "bg-success/10 text-success"
                  : line.type === "remove"
                    ? "bg-danger/10 text-danger"
                    : "text-subtle"
              }
            >
              {line.type === "add" ? "+ " : line.type === "remove" ? "- " : "  "}
              {line.content || " "}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
