import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Eye, GitBranch, GitCompare, History, Loader2, RotateCcw, X } from "lucide-react";
import { api } from "@/lib/api";
import { formatDate } from "@/lib/utils";
import { useI18n } from "@/lib/i18n";

interface VersionRow {
  id: string;
  source: "local" | "git";
  version_number?: number;
  commit_sha?: string;
  short_sha?: string;
  title?: string;
  author_name?: string;
  message?: string;
  created_at: string;
}

interface VersionContent {
  title: string;
  content: string;
  source: "local" | "git";
  version_number?: number;
  commit_sha?: string;
  created_at: string;
  author_name?: string;
}

interface DiffLine {
  type: "add" | "remove" | "same";
  content: string;
}

interface VersionDiff {
  from_version?: number;
  to_version?: number;
  from_sha?: string;
  to_sha?: string;
  lines: DiffLine[];
}

interface DocumentHistoryProps {
  documentId: string;
  canEdit?: boolean;
  onReverted?: () => void;
}

function buildDiffQuery(from: VersionRow, to: VersionRow): string {
  const params = new URLSearchParams();
  if (from.source === "git" && from.commit_sha) {
    params.set("from_sha", from.commit_sha);
  } else if (from.version_number) {
    params.set("from", String(from.version_number));
  }
  if (to.source === "git" && to.commit_sha) {
    params.set("to_sha", to.commit_sha);
  } else if (to.version_number) {
    params.set("to", String(to.version_number));
  }
  return params.toString();
}

function buildViewQuery(v: VersionRow): string {
  const params = new URLSearchParams();
  if (v.source === "git" && v.commit_sha) {
    params.set("sha", v.commit_sha);
  } else if (v.version_number) {
    params.set("version", String(v.version_number));
  }
  return params.toString();
}

function versionLabel(v: VersionRow, t: (key: string, vars?: Record<string, string | number>) => string) {
  if (v.source === "git") {
    return t("document.gitVersion", { sha: v.short_sha || v.commit_sha?.slice(0, 8) || "?" });
  }
  return t("document.versionN", { n: v.version_number ?? 0 });
}

export function DocumentHistory({ documentId, canEdit, onReverted }: DocumentHistoryProps) {
  const { t } = useI18n();
  const qc = useQueryClient();
  const [open, setOpen] = useState(false);
  const [diffQuery, setDiffQuery] = useState<string | null>(null);
  const [viewQuery, setViewQuery] = useState<string | null>(null);

  const { data, isLoading } = useQuery({
    queryKey: ["doc-versions", documentId],
    queryFn: () => api<{ items: VersionRow[] }>(`/api/documents/${documentId}/versions`),
    enabled: open,
    staleTime: 60_000,
  });

  const { data: diff, isLoading: diffLoading } = useQuery({
    queryKey: ["doc-history-diff", documentId, diffQuery],
    queryFn: () => api<VersionDiff>(`/api/documents/${documentId}/history/diff?${diffQuery}`),
    enabled: diffQuery !== null,
  });

  const { data: viewed, isLoading: viewLoading } = useQuery({
    queryKey: ["doc-history-content", documentId, viewQuery],
    queryFn: () => api<VersionContent>(`/api/documents/${documentId}/history/content?${viewQuery}`),
    enabled: viewQuery !== null,
  });

  const revert = useMutation({
    mutationFn: (version: number) =>
      api(`/api/documents/${documentId}/revert/${version}`, { method: "POST" }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["doc-versions", documentId] });
      onReverted?.();
      setOpen(false);
    },
  });

  const close = () => {
    setOpen(false);
    setDiffQuery(null);
    setViewQuery(null);
  };

  return (
    <>
      <button type="button" className="btn-ghost !px-2" onClick={() => setOpen(true)} aria-label={t("document.history")}>
        <History className="h-4 w-4" />
      </button>
      {open && (
        <div className="fixed inset-0 z-50 flex items-start justify-center bg-black/75 p-4 pt-[min(10vh,5rem)] backdrop-blur-sm">
          <div className="glass flex max-h-[min(85vh,52rem)] w-full max-w-2xl flex-col overflow-hidden">
            <div className="flex items-center justify-between border-b border-default px-5 py-4">
              <h2 className="flex items-center gap-2 text-lg font-semibold text-fg">
                <History className="h-5 w-5 text-primary" />
                {t("document.history")}
              </h2>
              <button type="button" className="btn-ghost !px-2" onClick={close}>
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
                    const newer = data.items[i + 1];
                    const viewKey = buildViewQuery(v);
                    return (
                      <div
                        key={`${v.source}-${v.id}`}
                        className="flex flex-wrap items-center justify-between gap-3 rounded-xl border border-default bg-surface-muted px-4 py-3"
                      >
                        <div className="min-w-0">
                          <p className="flex items-center gap-2 font-medium text-fg">
                            {v.source === "git" && <GitBranch className="h-3.5 w-3.5 shrink-0 text-primary" />}
                            {versionLabel(v, t)}
                          </p>
                          {v.message && (
                            <p className="mt-0.5 line-clamp-2 text-xs text-muted">{v.message}</p>
                          )}
                          {v.title && v.source === "local" && !v.message && (
                            <p className="mt-0.5 line-clamp-1 text-xs text-muted">{v.title}</p>
                          )}
                          <p className="mt-1 text-xs text-subtle">
                            {formatDate(v.created_at)}
                            {v.author_name && ` · ${v.author_name}`}
                          </p>
                        </div>
                        <div className="flex flex-wrap gap-1">
                          {viewKey && (
                            <button
                              type="button"
                              className="btn-ghost text-xs"
                              onClick={() => {
                                setDiffQuery(null);
                                setViewQuery(viewKey);
                              }}
                            >
                              <Eye className="mr-1 inline h-3.5 w-3.5" />
                              {t("document.viewVersion")}
                            </button>
                          )}
                          {canEdit && v.source === "local" && v.version_number && i > 0 && (
                            <button
                              type="button"
                              className="btn-ghost text-xs"
                              disabled={revert.isPending}
                              onClick={() => {
                                if (window.confirm(t("document.revertConfirm", { n: v.version_number! }))) {
                                  revert.mutate(v.version_number!);
                                }
                              }}
                            >
                              <RotateCcw className="mr-1 inline h-3.5 w-3.5" />
                              {t("document.revert")}
                            </button>
                          )}
                          {newer && (
                            <button
                              type="button"
                              className="btn-ghost text-xs"
                              onClick={() => {
                                setViewQuery(null);
                                setDiffQuery(buildDiffQuery(newer, v));
                              }}
                            >
                              <GitCompare className="mr-1 inline h-3.5 w-3.5" />
                              {t("document.compareWithLabel", { label: versionLabel(newer, t) })}
                            </button>
                          )}
                        </div>
                      </div>
                    );
                  })}
                  {data?.items.length === 0 && (
                    <p className="text-sm text-subtle">{t("document.noVersions")}</p>
                  )}
                </div>
              )}
              {viewQuery && (
                <div className="mt-6 border-t border-default pt-5">
                  <div className="mb-3 flex items-center justify-between gap-2">
                    <h3 className="text-sm font-semibold text-fg">{t("document.versionContentTitle")}</h3>
                    <button type="button" className="btn-ghost !px-2 text-xs" onClick={() => setViewQuery(null)}>
                      <X className="h-3.5 w-3.5" />
                    </button>
                  </div>
                  {viewLoading ? (
                    <Loader2 className="h-5 w-5 animate-spin text-primary" />
                  ) : (
                    <>
                      {viewed?.title && (
                        <p className="mb-2 text-sm font-medium text-fg">{viewed.title}</p>
                      )}
                      <pre className="max-h-72 overflow-auto whitespace-pre-wrap rounded-lg bg-surface-muted p-3 font-mono text-xs leading-relaxed text-fg-secondary">
                        {viewed?.content || ""}
                      </pre>
                    </>
                  )}
                </div>
              )}
              {diffQuery && (
                <div className="mt-6 border-t border-default pt-5">
                  <div className="mb-3 flex items-center justify-between gap-2">
                    <h3 className="text-sm font-semibold text-fg">{t("document.diffTitleGeneric")}</h3>
                    <button type="button" className="btn-ghost !px-2 text-xs" onClick={() => setDiffQuery(null)}>
                      <X className="h-3.5 w-3.5" />
                    </button>
                  </div>
                  {diffLoading ? (
                    <Loader2 className="h-5 w-5 animate-spin text-primary" />
                  ) : (
                    <pre className="max-h-64 overflow-auto rounded-lg bg-surface-muted p-3 font-mono text-xs leading-relaxed">
                      {diff?.lines.map((line, idx) => (
                        <div
                          key={idx}
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
