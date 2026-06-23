import { useState } from "react";
import { Link, Outlet, useLocation, useParams } from "react-router-dom";
import { BookOpen, ArrowLeft, FileText, GitBranch, Globe, Loader2, Plus, RefreshCw } from "lucide-react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api, ApiError, optionalAuthApi } from "@/lib/api";
import { useAuthStore } from "@/lib/store";
import { DocumentTree } from "@/components/document-tree";
import { FadeIn } from "@/components/motion-wrapper";
import { formatDate } from "@/lib/utils";
import { cn } from "@/lib/utils";
import { useI18n } from "@/lib/i18n";
import { canManageBooks } from "@/lib/roles";
import type { DocItem } from "@/lib/doc-tree";

interface Space {
  id: string;
  slug: string;
  name: string;
  description?: string;
  is_public: boolean;
  can_edit?: boolean;
}

interface Repository {
  id: string;
  name: string;
  branch: string;
  sync_mode: string;
  last_sync_at?: string;
  last_sync_status?: string;
  last_sync_error?: string;
}

interface SavedBook {
  id: string;
  slug: string;
  title: string;
  status: string;
  root_path: string;
  sources_stale?: boolean;
}

function syncStatusClass(status?: string) {
  if (status === "completed" || status === "success") return "text-success-soft";
  if (status === "failed" || status === "error") return "text-danger-soft";
  return "text-subtle";
}

export function SpaceDocLayout() {
  const { slug, docSlug, bookSlug } = useParams<{ slug: string; docSlug?: string; bookSlug?: string }>();
  const location = useLocation();
  const isBooks = location.pathname.includes("/books");
  const user = useAuthStore((s) => s.user);
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  const canManage = canManageBooks(user);
  const { t, statusLabel } = useI18n();
  const qc = useQueryClient();
  const [syncingId, setSyncingId] = useState<string | null>(null);
  const [syncError, setSyncError] = useState("");
  const [showCreateDoc, setShowCreateDoc] = useState(false);
  const [newDocTitle, setNewDocTitle] = useState("");
  const [newDocPath, setNewDocPath] = useState("");
  const [createError, setCreateError] = useState("");

  const { data: space } = useQuery({
    queryKey: ["space", slug],
    queryFn: () => optionalAuthApi<Space>(`/api/spaces/${slug}`),
    enabled: !!slug,
  });

  const canEdit = space?.can_edit === true;

  const { data: repos } = useQuery({
    queryKey: ["space-repos", slug],
    queryFn: () => optionalAuthApi<{ items: Repository[] }>(`/api/spaces/${slug}/repositories`),
    enabled: !!slug && isAuthenticated,
  });

  const { data: docs, isLoading: docsLoading } = useQuery({
    queryKey: ["documents", slug],
    queryFn: () => optionalAuthApi<{ items: DocItem[] }>(`/api/spaces/${slug}/documents`),
    enabled: !!slug && !isBooks,
  });

  const { data: booksData } = useQuery({
    queryKey: ["books", slug],
    queryFn: () =>
      optionalAuthApi<{ saved: SavedBook[]; candidates: unknown[]; llm_available: boolean }>(
        `/api/spaces/${slug}/books`,
      ),
    enabled: !!slug && isBooks,
  });

  const createDoc = useMutation({
    mutationFn: (body: { title: string; path: string; content: string }) =>
      api(`/api/spaces/${slug}/documents`, { method: "POST", body: JSON.stringify(body) }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["documents", slug] });
      setShowCreateDoc(false);
      setNewDocTitle("");
      setNewDocPath("");
      setCreateError("");
    },
    onError: (e) => setCreateError(e instanceof ApiError ? e.message : t("common.failed")),
  });

  async function triggerSync(repoId: string) {
    setSyncingId(repoId);
    setSyncError("");
    try {
      await api(`/api/spaces/${slug}/repositories/${repoId}/sync`, { method: "POST" });
      qc.invalidateQueries({ queryKey: ["space-repos", slug] });
      qc.invalidateQueries({ queryKey: ["documents", slug] });
    } catch (e) {
      setSyncError(e instanceof ApiError ? e.message : t("space.syncFailed"));
    } finally {
      setSyncingId(null);
    }
  }

  const activeDoc = docs?.items.find((d) => d.slug === docSlug);
  const pagesTab = `/spaces/${slug}`;
  const booksTab = `/spaces/${slug}/books`;

  return (
    <div className="mx-auto max-w-7xl px-4 py-8 sm:px-6">
      <FadeIn>
        <Link
          to="/spaces"
          className="mb-4 inline-flex items-center gap-1.5 text-sm text-muted transition-colors hover:text-primary"
        >
          <ArrowLeft className="h-4 w-4 shrink-0" />
          {t("space.backToSpaces")}
        </Link>
        <div className="flex flex-wrap items-center gap-3">
          <h1 className="text-2xl font-bold gradient-text sm:text-3xl">{space?.name || slug}</h1>
          {space?.is_public && (
            <span className="badge badge-success">
              <Globe className="mr-1 h-3 w-3" />
              {t("common.public")}
            </span>
          )}
        </div>
        {space?.description && <p className="mt-1 text-sm text-muted">{space.description}</p>}

        <nav className="mt-5 flex gap-1 border-b border-default">
          <Link
            to={pagesTab}
            className={cn(
              "flex items-center gap-2 border-b-2 px-4 py-2.5 text-sm font-medium transition-colors -mb-px",
              !isBooks ? "border-primary text-primary" : "border-transparent text-muted hover:text-fg",
            )}
          >
            <FileText className="h-4 w-4" />
            {t("space.pages")}
          </Link>
          <Link
            to={booksTab}
            className={cn(
              "flex items-center gap-2 border-b-2 px-4 py-2.5 text-sm font-medium transition-colors -mb-px",
              isBooks ? "border-primary text-primary" : "border-transparent text-muted hover:text-fg",
            )}
          >
            <BookOpen className="h-4 w-4" />
            {t("space.books")}
          </Link>
        </nav>
      </FadeIn>

      {isAuthenticated && repos && repos.items.length > 0 && !isBooks && (
        <details className="mt-4 glass group" open={canEdit}>
          <summary className="flex cursor-pointer list-none items-center gap-2 px-4 py-3 text-sm text-muted marker:content-none">
            <GitBranch className="h-4 w-4 text-primary" />
            <span>
              {repos.items.length === 1
                ? t("space.linkedReposOne")
                : t("space.linkedReposMany", { count: repos.items.length })}
            </span>
            <span className="ml-auto text-xs text-subtle group-open:hidden">{t("space.show")}</span>
          </summary>
          <div className="space-y-2 border-t border-default px-4 pb-4 pt-2">
            {syncError && <p className="text-xs text-danger-soft">{syncError}</p>}
            {repos.items.map((repo) => (
              <div key={repo.id} className="flex flex-wrap items-center justify-between gap-2 text-xs">
                <span className="text-fg-secondary">{repo.name}</span>
                <div className="flex flex-wrap items-center gap-2 text-subtle">
                  <span>
                    {repo.branch} · {repo.sync_mode}
                    {repo.last_sync_at && ` · ${formatDate(repo.last_sync_at)}`}
                  </span>
                  {repo.last_sync_status && (
                    <span className={syncStatusClass(repo.last_sync_status)}>
                      · {repo.last_sync_status}
                    </span>
                  )}
                  {repo.last_sync_error && (
                    <span className="text-danger-soft" title={repo.last_sync_error}>
                      · {t("space.syncError")}
                    </span>
                  )}
                  {canEdit && (
                    <button
                      type="button"
                      className="btn-ghost !px-2 !py-1 text-xs"
                      disabled={syncingId === repo.id}
                      onClick={() => triggerSync(repo.id)}
                    >
                      {syncingId === repo.id ? (
                        <Loader2 className="h-3.5 w-3.5 animate-spin" />
                      ) : (
                        <RefreshCw className="h-3.5 w-3.5" />
                      )}
                      {t("space.syncNow")}
                    </button>
                  )}
                </div>
              </div>
            ))}
          </div>
        </details>
      )}

      {!isBooks && docsLoading ? (
        <div className="flex justify-center py-20">
          <Loader2 className="h-8 w-8 animate-spin text-primary" />
        </div>
      ) : (
        <div className="mt-6 flex flex-col gap-6 lg:flex-row lg:items-start">
          <aside className="glass w-full shrink-0 p-3 lg:sticky lg:top-24 lg:w-72 lg:max-h-[calc(100vh-7rem)] lg:overflow-y-auto">
            {isBooks ? (
              <div className="space-y-1">
                <p className="px-2 pb-2 text-xs font-semibold uppercase tracking-wide text-subtle">
                  {t("space.savedBooks")}
                </p>
                {booksData?.saved.map((book) => (
                  <Link
                    key={book.id}
                    to={`/spaces/${slug}/books/${book.slug}`}
                    className={cn(
                      "block rounded-lg px-3 py-2 text-sm transition-colors",
                      bookSlug === book.slug
                        ? "bg-highlight-row text-fg font-medium"
                        : "text-fg-secondary hover:bg-surface-muted",
                    )}
                  >
                    <span className="line-clamp-2">{book.title}</span>
                    <span className="mt-0.5 flex flex-wrap gap-1 text-xs text-subtle">
                      <span>{statusLabel(book.status)}</span>
                      {book.sources_stale && (
                        <span className="text-warning-soft">· {t("space.sourcesChanged")}</span>
                      )}
                    </span>
                  </Link>
                ))}
                {booksData && booksData.saved.length === 0 && (
                  <p className="px-2 text-xs text-subtle">
                    {canManage ? t("space.noBooksSidebar") : t("space.noBooksSidebarReader")}
                  </p>
                )}
              </div>
            ) : (
              <>
                {canEdit && (
                  <div className="mb-2 px-1">
                    {!showCreateDoc ? (
                      <button
                        type="button"
                        className="btn-secondary w-full text-xs"
                        onClick={() => setShowCreateDoc(true)}
                      >
                        <Plus className="h-3.5 w-3.5" />
                        {t("document.createPage")}
                      </button>
                    ) : (
                      <form
                        className="space-y-2 rounded-lg border border-default bg-surface-muted p-2"
                        onSubmit={(e) => {
                          e.preventDefault();
                          const path = newDocPath.trim() || `${newDocTitle.trim().toLowerCase().replace(/\s+/g, "-")}.md`;
                          createDoc.mutate({
                            title: newDocTitle.trim(),
                            path: path.endsWith(".md") ? path : `${path}.md`,
                            content: `# ${newDocTitle.trim()}\n`,
                          });
                        }}
                      >
                        <input
                          className="input-field text-xs"
                          placeholder={t("document.titlePlaceholder")}
                          value={newDocTitle}
                          onChange={(e) => setNewDocTitle(e.target.value)}
                          required
                        />
                        <input
                          className="input-field text-xs"
                          placeholder={t("document.pathPlaceholder")}
                          value={newDocPath}
                          onChange={(e) => setNewDocPath(e.target.value)}
                        />
                        {createError && <p className="text-xs text-danger-soft">{createError}</p>}
                        <div className="flex gap-1">
                          <button type="submit" className="btn-primary flex-1 text-xs" disabled={createDoc.isPending}>
                            {createDoc.isPending ? <Loader2 className="h-3 w-3 animate-spin" /> : t("common.save")}
                          </button>
                          <button type="button" className="btn-ghost text-xs" onClick={() => setShowCreateDoc(false)}>
                            {t("common.cancel")}
                          </button>
                        </div>
                      </form>
                    )}
                  </div>
                )}
              <div data-tour="doc-tree">
                <DocumentTree
                  spaceSlug={slug!}
                  documents={docs?.items ?? []}
                  activeSlug={docSlug}
                  activePath={activeDoc?.path}
                />
              </div>
              </>
            )}
          </aside>

          <div className="min-w-0 flex-1">
            <Outlet />
          </div>
        </div>
      )}
    </div>
  );
}
