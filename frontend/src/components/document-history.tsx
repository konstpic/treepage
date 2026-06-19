import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { GitCompare, History, Loader2, X } from "lucide-react";
import { api } from "@/lib/api";
import { formatDate } from "@/lib/utils";
import { useI18n } from "@/lib/i18n";

interface VersionRow {
  id: string;
  version_number: number;
  title: string;
  author_name?: string;
  created_at: string;
}

interface DiffLine {
  type: "add" | "remove" | "same";
  content: string;
}

interface VersionDiff {
  from_version: number;
  to_version: number;
  lines: DiffLine[];
}

interface DocumentHistoryProps {
  documentId: string;
}

export function DocumentHistory({ documentId }: DocumentHistoryProps) {
  const { t } = useI18n();
  const [open, setOpen] = useState(false);
  const [diffVersions, setDiffVersions] = useState<{ from: number; to: number } | null>(null);

  const { data, isLoading } = useQuery({
    queryKey: ["doc-versions", documentId],
    queryFn: () => api<{ items: VersionRow[] }>(`/api/documents/${documentId}/versions`),
    enabled: open,
  });

  const { data: diff, isLoading: diffLoading } = useQuery({
    queryKey: ["doc-diff", documentId, diffVersions?.from, diffVersions?.to],
    queryFn: () =>
      api<VersionDiff>(
        `/api/documents/${documentId}/versions/${diffVersions!.to}/diff?from=${diffVersions!.from}`,
      ),
    enabled: diffVersions !== null,
  });

  return (
    <>
      <button type="button" className="btn-ghost !px-2" onClick={() => setOpen(true)} aria-label={t("document.history")}>
        <History className="h-4 w-4" />
      </button>
      {open && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/75 p-4 backdrop-blur-sm">
          <div className="glass flex max-h-[85vh] w-full max-w-2xl flex-col overflow-hidden">
            <div className="flex items-center justify-between border-b border-default px-5 py-4">
              <h2 className="flex items-center gap-2 text-lg font-semibold text-fg">
                <History className="h-5 w-5 text-primary" />
                {t("document.history")}
              </h2>
              <button type="button" className="btn-ghost !px-2" onClick={() => { setOpen(false); setDiffVersions(null); }}>
                <X className="h-4 w-4" />
              </button>
            </div>
            <div className="flex-1 overflow-y-auto p-5">
              {isLoading ? (
                <div className="flex justify-center py-10">
                  <Loader2 className="h-6 w-6 animate-spin text-primary" />
                </div>
              ) : (
                <div className="space-y-2">
                  {data?.items.map((v, i) => {
                    const prev = data.items[i + 1];
                    return (
                      <div
                        key={v.id}
                        className="flex flex-wrap items-center justify-between gap-3 rounded-xl border border-default bg-surface-muted px-4 py-3"
                      >
                        <div>
                          <p className="font-medium text-fg">
                            {t("document.versionN", { n: v.version_number })}
                          </p>
                          <p className="text-xs text-subtle">
                            {formatDate(v.created_at)}
                            {v.author_name && ` · ${v.author_name}`}
                          </p>
                        </div>
                        {prev && (
                          <button
                            type="button"
                            className="btn-ghost text-xs"
                            onClick={() => setDiffVersions({ from: prev.version_number, to: v.version_number })}
                          >
                            <GitCompare className="mr-1 inline h-3.5 w-3.5" />
                            {t("document.compareWith", { n: prev.version_number })}
                          </button>
                        )}
                      </div>
                    );
                  })}
                  {data?.items.length === 0 && (
                    <p className="text-sm text-subtle">{t("document.noVersions")}</p>
                  )}
                </div>
              )}
              {diffVersions && (
                <div className="mt-6 border-t border-default pt-5">
                  <h3 className="mb-3 text-sm font-semibold text-fg">
                    {t("document.diffTitle", { from: diffVersions.from, to: diffVersions.to })}
                  </h3>
                  {diffLoading ? (
                    <Loader2 className="h-5 w-5 animate-spin text-primary" />
                  ) : (
                    <pre className="max-h-64 overflow-auto rounded-lg bg-surface-muted p-3 font-mono text-xs leading-relaxed">
                      {diff?.lines.map((line, i) => (
                        <div
                          key={i}
                          className={
                            line.type === "add"
                              ? "bg-success-soft/20 text-success-soft"
                              : line.type === "remove"
                                ? "bg-danger-soft/20 text-danger-soft"
                                : "text-fg-secondary"
                          }
                        >
                          {line.type === "add" ? "+ " : line.type === "remove" ? "- " : "  "}
                          {line.content}
                        </div>
                      ))}
                    </pre>
                  )}
                </div>
              )}
            </div>
          </div>
        </div>
      )}
    </>
  );
}
