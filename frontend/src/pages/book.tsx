import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Link, useParams } from "react-router-dom";
import { Download, Languages, Loader2, RefreshCw, Sparkles, Trash2 } from "lucide-react";
import { api, ApiError, optionalAuthApi } from "@/lib/api";
import { FadeIn } from "@/components/motion-wrapper";
import { MarkdownRenderer } from "@/components/markdown-renderer";
import { formatDate } from "@/lib/utils";
import { useAuthStore } from "@/lib/store";
import { useI18n } from "@/lib/i18n";
import { canManageBooksInSpace, type SpaceRole } from "@/lib/roles";

interface BookDetail {
  id: string;
  slug: string;
  title: string;
  description: string;
  root_path: string;
  audience: string;
  status: string;
  sources_stale: boolean;
  chapter_count: number;
  enhanced: boolean;
  error_message?: string;
  content_markdown?: string;
  generated_at?: string;
  updated_at: string;
  llm_available?: boolean;
  translated?: boolean;
}

export function BookReaderPage() {
  const { slug, bookSlug } = useParams<{ slug: string; bookSlug: string }>();
  const qc = useQueryClient();
  const { t, localeId, statusLabel, audienceLabel } = useI18n();
  const user = useAuthStore((s) => s.user);
  const [error, setError] = useState("");

  const { data: spaceMeta } = useQuery({
    queryKey: ["space", slug],
    queryFn: () =>
      optionalAuthApi<{ my_role?: SpaceRole; can_edit?: boolean }>(`/api/spaces/${slug}`),
    enabled: !!slug,
  });
  const canManage = canManageBooksInSpace(spaceMeta?.my_role, user, spaceMeta?.can_edit);

  const { data: book, isLoading } = useQuery({
    queryKey: ["book", slug, bookSlug, localeId],
    queryFn: () =>
      optionalAuthApi<BookDetail>(
        `/api/spaces/${slug}/books/${bookSlug}?content=true&lang=${localeId}`,
      ),
    enabled: !!slug && !!bookSlug,
    refetchInterval: (q) => (q.state.data?.status === "generating" ? 3000 : false),
  });

  const generate = useMutation({
    mutationFn: (force: boolean) =>
      api<BookDetail>(`/api/spaces/${slug}/books/${bookSlug}/generate${force ? "?force=true" : ""}`, {
        method: "POST",
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["book", slug, bookSlug] });
      qc.invalidateQueries({ queryKey: ["books", slug] });
      setError("");
    },
    onError: (e) => setError(e instanceof ApiError ? e.message : t("book.generationFailed")),
  });

  const remove = useMutation({
    mutationFn: () => api(`/api/spaces/${slug}/books/${bookSlug}`, { method: "DELETE" }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["books", slug] });
      window.location.href = `/spaces/${slug}/books`;
    },
    onError: (e) => setError(e instanceof ApiError ? e.message : t("book.deleteFailed")),
  });

  async function downloadMd() {
    const cfg = await import("@/lib/config").then((m) => m.getRuntimeConfig());
    const token = localStorage.getItem("access_token");
    const res = await fetch(`${cfg.apiUrl}/api/spaces/${slug}/books/${bookSlug}?format=md`, {
      headers: token ? { Authorization: `Bearer ${token}` } : {},
    });
    if (!res.ok) return;
    const blob = await res.blob();
    const a = document.createElement("a");
    a.href = URL.createObjectURL(blob);
    a.download = `${bookSlug}.md`;
    a.click();
    URL.revokeObjectURL(a.href);
  }

  if (isLoading) {
    return (
      <div className="flex justify-center py-20">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    );
  }

  if (!book) {
    return <p className="text-sm text-muted">{t("book.notFound")}</p>;
  }

  const hasContent = book.status === "ready" && !!book.content_markdown;
  const llmAvailable = book.llm_available ?? false;
  const shellMessage =
    !canManage && !hasContent
      ? t("book.shellViewer")
      : !canManage
        ? ""
        : book.audience === "architect" && llmAvailable
          ? t("book.shellArchitectLlm")
          : book.audience === "developer" && llmAvailable
            ? t("book.shellDeveloperLlm")
            : t("book.shellAuth");

  return (
    <FadeIn>
      <div className="glass p-4">
        <div className="flex flex-wrap items-start justify-between gap-4">
          <div>
            <h2 className="text-xl font-semibold text-fg">{book.title}</h2>
            <p className="text-sm text-muted">{book.description}</p>
            <div className="mt-2 flex flex-wrap gap-2 text-xs text-subtle">
              {book.translated && (
                <span className="badge badge-primary inline-flex items-center gap-1">
                  <Languages className="h-3 w-3" />
                  {t("document.autoTranslated")}
                </span>
              )}
              <span className="badge badge-neutral">{statusLabel(book.status)}</span>
              <span className="badge badge-neutral">{audienceLabel(book.audience)}</span>
              {canManage && (
                <>
                  {llmAvailable ? (
                    <span className="badge badge-success">{t("book.llmConnected")}</span>
                  ) : (
                    <span className="badge badge-neutral">{t("book.llmDisconnected")}</span>
                  )}
                  {book.audience === "developer" && llmAvailable && (
                    <span className="badge badge-neutral">{t("book.llmRoleDeveloper")}</span>
                  )}
                  {book.audience === "architect" && llmAvailable && (
                    <span className="badge badge-neutral">{t("book.llmRoleArchitect")}</span>
                  )}
                  {book.enhanced && (
                    <span className="badge badge-success">
                      <Sparkles className="mr-1 inline h-3 w-3" />
                      AI
                    </span>
                  )}
                </>
              )}
              {book.sources_stale && (
                <span className="badge badge-neutral">{t("book.sourcesChanged")}</span>
              )}
              {book.generated_at && (
                <span>{t("book.generated", { date: formatDate(book.generated_at) })}</span>
              )}
            </div>
            {book.error_message && <p className="mt-2 text-sm text-danger-soft">{book.error_message}</p>}
          </div>
          <div className="flex flex-wrap gap-2">
            {hasContent && (
              <button type="button" className="btn-secondary" onClick={downloadMd}>
                <Download className="h-4 w-4" />
                .md
              </button>
            )}
            {canManage && (
              <>
                <button
                  type="button"
                  className="btn-primary"
                  disabled={generate.isPending || book.status === "generating"}
                  onClick={() => generate.mutate(!hasContent ? false : true)}
                >
                  {generate.isPending || book.status === "generating" ? (
                    <Loader2 className="h-4 w-4 animate-spin" />
                  ) : (
                    <Sparkles className="h-4 w-4" />
                  )}
                  {hasContent ? t("book.rebuild") : t("book.build")}
                </button>
                {hasContent && (
                  <button
                    type="button"
                    className="btn-secondary"
                    disabled={generate.isPending}
                    onClick={() => generate.mutate(true)}
                  >
                    <RefreshCw className="h-4 w-4" />
                    {t("book.force")}
                  </button>
                )}
                <button
                  type="button"
                  className="btn-ghost text-danger-soft"
                  onClick={() => {
                    if (window.confirm(t("book.deleteConfirm", { title: book.title }))) remove.mutate();
                  }}
                >
                  <Trash2 className="h-4 w-4" />
                </button>
              </>
            )}
          </div>
        </div>
        {error && <p className="mt-3 text-sm text-danger-soft">{error}</p>}
        <p className="mt-3 text-xs text-subtle">
          {t("book.source")}: <code>{book.root_path}/</code> · {book.chapter_count} {t("common.chapters")} ·{" "}
          <Link to={`/spaces/${slug}/books`} className="text-primary hover:text-primary-hover">
            {t("book.allBooks")}
          </Link>
        </p>
      </div>

      {book.status === "generating" && (
        <div className="glass mt-4 flex items-center gap-3 p-6 text-sm text-muted">
          <Loader2 className="h-5 w-5 animate-spin text-primary" />
          {book.audience === "architect" && llmAvailable
            ? t("book.generatingArchitect")
            : book.audience === "developer" && llmAvailable
              ? t("book.generatingWithLlm")
              : t("book.generating")}
        </div>
      )}

      {hasContent && (
        <div className="glass mt-4 p-6">
          <div className="prose-doc max-w-none">
            <MarkdownRenderer content={book.content_markdown!} spaceSlug={slug!} />
          </div>
        </div>
      )}

      {!hasContent && book.status !== "generating" && shellMessage && (
        <div className="glass mt-4 p-8 text-center text-sm text-muted">
          {shellMessage}
        </div>
      )}
    </FadeIn>
  );
}
