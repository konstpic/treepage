import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useParams, useNavigate } from "react-router-dom";
import { Edit3, Languages, Loader2, Star, Trash2 } from "lucide-react";
import { api, ApiError, optionalAuthApi } from "@/lib/api";
import { MarkdownRenderer } from "@/components/markdown-renderer";
import { DocumentSyncDiff } from "@/components/document-sync-diff";
import { DocumentEditor } from "@/components/document-editor";
import { DocumentAttachments } from "@/components/document-attachments";
import { DocumentComments } from "@/components/document-comments";
import type { PublishPRInput } from "@/components/publish-pr-dialog";
import { DocBreadcrumbs } from "@/components/doc-breadcrumbs";
import { DocumentHistory } from "@/components/document-history";
import { FadeIn } from "@/components/motion-wrapper";
import { formatDate } from "@/lib/utils";
import { useAuthStore } from "@/lib/store";
import { useI18n } from "@/lib/i18n";

interface Document {
  id: string;
  title: string;
  content: string;
  path: string;
  author_name?: string;
  updated_at: string;
  tags?: string[];
  repository_id?: string;
  has_pending_changes?: boolean;
  is_published?: boolean;
  workflow_state?: string;
  is_favorite?: boolean;
  translated?: boolean;
  source_language?: string;
  display_language?: string;
}

interface SpaceMeta {
  name: string;
  can_edit?: boolean;
}

interface Repository {
  id: string;
  branch: string;
}

interface PublishResult {
  branch: string;
  commit_sha?: string;
  pr_url?: string;
  message?: string;
}

export function DocumentPage() {
  const { slug, docSlug } = useParams<{ slug: string; docSlug: string }>();
  const navigate = useNavigate();
  const { t, localeId } = useI18n();
  const { isAuthenticated } = useAuthStore();
  const qc = useQueryClient();
  const [editing, setEditing] = useState(false);
  const [editTitle, setEditTitle] = useState("");
  const [editContent, setEditContent] = useState("");
  const [saveError, setSaveError] = useState("");
  const [publishError, setPublishError] = useState("");
  const [publishNotice, setPublishNotice] = useState("");

  const { data: space } = useQuery({
    queryKey: ["space", slug],
    queryFn: () => optionalAuthApi<SpaceMeta>(`/api/spaces/${slug}`),
    enabled: !!slug,
  });

  const { data: doc, isLoading } = useQuery({
    queryKey: ["document", slug, docSlug, editing ? "raw" : localeId],
    queryFn: () =>
      optionalAuthApi<Document>(
        editing
          ? `/api/spaces/${slug}/documents/${docSlug}`
          : `/api/spaces/${slug}/documents/${docSlug}?lang=${localeId}`,
      ),
    enabled: !!slug && !!docSlug,
  });

  const { data: allDocs } = useQuery({
    queryKey: ["documents", slug],
    queryFn: () =>
      optionalAuthApi<{ items: { id: string; slug: string; title: string; path: string }[] }>(
        `/api/spaces/${slug}/documents`,
      ),
    enabled: !!slug,
  });

  const { data: repos } = useQuery({
    queryKey: ["repositories", slug],
    queryFn: () => optionalAuthApi<{ items: Repository[] }>(`/api/spaces/${slug}/repositories`),
    enabled: !!slug && editing && !!doc?.repository_id,
  });

  const linkedRepo = repos?.items.find((r) => r.id === doc?.repository_id);

  const saveDoc = useMutation({
    mutationFn: (opts?: { draft?: boolean }) =>
      api<Document>(`/api/documents/${doc!.id}`, {
        method: "PUT",
        body: JSON.stringify({
          title: editTitle,
          content: editContent,
          is_published: opts?.draft ? false : true,
        }),
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["document", slug, docSlug] });
      qc.invalidateQueries({ queryKey: ["documents", slug] });
      qc.invalidateQueries({ queryKey: ["doc-versions", doc!.id] });
      setEditing(false);
      setSaveError("");
    },
    onError: (e) => setSaveError(e instanceof ApiError ? e.message : t("common.failed")),
  });

  const publishWorkflow = useMutation({
    mutationFn: () => api(`/api/documents/${doc!.id}/publish-workflow`, { method: "POST" }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["document", slug, docSlug] });
      qc.invalidateQueries({ queryKey: ["documents", slug] });
    },
  });

  const submitReview = useMutation({
    mutationFn: () => api(`/api/documents/${doc!.id}/submit-review`, { method: "POST" }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["document", slug, docSlug] }),
  });

  const approveReview = useMutation({
    mutationFn: () => api(`/api/documents/${doc!.id}/approve`, { method: "POST" }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["document", slug, docSlug] }),
  });

  const toggleFavorite = useMutation({
    mutationFn: (add: boolean) =>
      add
        ? api(`/api/me/favorites/${doc!.id}`, { method: "POST" })
        : api(`/api/me/favorites/${doc!.id}`, { method: "DELETE" }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["document", slug, docSlug] });
      qc.invalidateQueries({ queryKey: ["favorites"] });
    },
  });

  const publishPR = useMutation({
    mutationFn: (input: PublishPRInput) =>
      api<PublishResult>(`/api/documents/${doc!.id}/publish`, {
        method: "POST",
        body: JSON.stringify({
          title: editTitle,
          content: editContent,
          branch: input.branch,
          commit_message: input.commitMessage,
          pr_title: input.prTitle,
          pr_body: input.prBody,
        }),
      }),
    onSuccess: (result) => {
      qc.invalidateQueries({ queryKey: ["document", slug, docSlug] });
      qc.invalidateQueries({ queryKey: ["doc-versions", doc!.id] });
      setPublishError("");
      if (result.pr_url) {
        setPublishNotice(t("documentEditor.publishSuccessPr", { url: result.pr_url }));
      } else if (result.message) {
        setPublishNotice(result.message);
      } else {
        setPublishNotice(t("documentEditor.publishSuccessPush", { branch: result.branch }));
      }
    },
    onError: (e) =>
      setPublishError(e instanceof ApiError ? e.message : t("documentEditor.publishFailed")),
  });

  const deleteDoc = useMutation({
    mutationFn: () => api(`/api/documents/${doc!.id}`, { method: "DELETE" }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["documents", slug] });
      navigate(`/spaces/${slug}`, { replace: true });
    },
  });

  function startEdit() {
    if (!doc) return;
    setEditTitle(doc.title);
    setEditContent(doc.content);
    setEditing(true);
    setSaveError("");
    setPublishError("");
    setPublishNotice("");
  }

  if (isLoading) {
    return (
      <div className="flex justify-center py-20">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    );
  }

  if (!doc) return null;

  const canEdit = space?.can_edit === true;

  return (
    <FadeIn>
      <DocBreadcrumbs
        spaceSlug={slug!}
        spaceName={space?.name}
        docPath={doc.path}
        docTitle={doc.title}
      />
      <article className="glass p-6 sm:p-8">
        {doc.has_pending_changes && doc.repository_id && (
          <DocumentSyncDiff documentId={doc.id} />
        )}
        {doc.workflow_state && doc.workflow_state !== "published" && (
          <div className="mb-4 rounded-xl border border-default bg-surface-muted px-4 py-3 text-sm text-fg">
            <span className="badge badge-neutral mr-2">{t(`workflow.${doc.workflow_state}`)}</span>
            {t("workflow.statusHint")}
          </div>
        )}
        {!editing && (
          <header className="mb-6 border-b border-default pb-5">
            <div className="flex flex-wrap items-start justify-between gap-3">
              <div>
                <p className="text-xs text-subtle">{doc.path}</p>
                <h1 className="mt-1 text-2xl font-bold text-fg sm:text-3xl">{doc.title}</h1>
              </div>
              <div className="flex items-center gap-1">
                {isAuthenticated && (
                  <button
                    type="button"
                    className="btn-ghost"
                    title={doc.is_favorite ? t("document.removeFavorite") : t("document.addFavorite")}
                    disabled={toggleFavorite.isPending}
                    onClick={() => toggleFavorite.mutate(!doc.is_favorite)}
                  >
                    <Star
                      className={`h-4 w-4 ${doc.is_favorite ? "fill-primary text-primary" : ""}`}
                    />
                  </button>
                )}
                {canEdit && (
                  <>
                    <DocumentHistory
                    documentId={doc.id}
                    canEdit
                    onReverted={() => qc.invalidateQueries({ queryKey: ["document", slug, docSlug] })}
                  />
                  <button type="button" className="btn-secondary" onClick={startEdit}>
                    <Edit3 className="h-4 w-4" />
                    {t("document.edit")}
                  </button>
                  <button
                    type="button"
                    className="btn-ghost text-danger-soft"
                    disabled={deleteDoc.isPending}
                    onClick={() => {
                      if (window.confirm(t("document.deleteConfirm"))) deleteDoc.mutate();
                    }}
                  >
                    <Trash2 className="h-4 w-4" />
                  </button>
                  </>
                )}
              </div>
            </div>
            <div className="mt-3 flex flex-wrap items-center gap-3 text-sm text-subtle">
              {doc.translated && (
                <span className="badge badge-primary inline-flex items-center gap-1">
                  <Languages className="h-3 w-3" />
                  {t("document.autoTranslated")}
                </span>
              )}
              {doc.author_name && <span>{doc.author_name}</span>}
              <span>{formatDate(doc.updated_at)}</span>
              {doc.tags?.map((tag) => (
                <span key={tag} className="badge badge-primary">
                  {tag}
                </span>
              ))}
            </div>
          </header>
        )}

        {editing ? (
          <>
            {saveError && <p className="mb-3 text-sm text-danger-soft">{saveError}</p>}
            {publishNotice && (
              <p className="mb-3 rounded-xl border border-default bg-surface-muted px-4 py-3 text-sm text-fg">
                {publishNotice}
              </p>
            )}
            <DocumentEditor
              title={editTitle}
              content={editContent}
              path={doc.path}
              spaceSlug={slug!}
              gitHint={doc.repository_id ? doc.path : undefined}
              gitLinked={!!doc.repository_id}
              defaultBranch={linkedRepo?.branch ?? "main"}
              documents={allDocs?.items ?? []}
              saving={saveDoc.isPending}
              publishing={publishPR.isPending}
              publishError={publishError}
              onTitleChange={setEditTitle}
              onContentChange={setEditContent}
              onSave={() => saveDoc.mutate({})}
              onSaveDraft={() => saveDoc.mutate({ draft: true })}
              onPublishLocal={
                doc.is_published === false || doc.workflow_state === "approved"
                  ? () => publishWorkflow.mutate()
                  : undefined
              }
              onSubmitReview={
                canEdit && (doc.workflow_state === "draft" || !doc.workflow_state)
                  ? () => submitReview.mutate()
                  : undefined
              }
              onApproveReview={
                canEdit && doc.workflow_state === "in_review" ? () => approveReview.mutate() : undefined
              }
              onPublishPR={async (input) => {
                await publishPR.mutateAsync(input);
              }}
              onCancel={() => setEditing(false)}
            />
          </>
        ) : (
          <>
            <MarkdownRenderer
              content={doc.content}
              spaceSlug={slug}
              documents={allDocs?.items ?? []}
              docPath={doc.path}
            />
            <DocumentAttachments documentId={doc.id} canEdit={canEdit} />
            <DocumentComments documentId={doc.id} />
          </>
        )}
      </article>
    </FadeIn>
  );
}
