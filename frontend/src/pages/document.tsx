import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useParams } from "react-router-dom";
import { Edit3, Languages, Loader2 } from "lucide-react";
import { api, ApiError, optionalAuthApi } from "@/lib/api";
import { MarkdownRenderer } from "@/components/markdown-renderer";
import { DocumentEditor } from "@/components/document-editor";
import type { PublishPRInput } from "@/components/publish-pr-dialog";
import { DocBreadcrumbs } from "@/components/doc-breadcrumbs";
import { DocumentHistory } from "@/components/document-history";
import { FadeIn } from "@/components/motion-wrapper";
import { formatDate } from "@/lib/utils";
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
  const { t, localeId } = useI18n();
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
    mutationFn: () =>
      api<Document>(`/api/documents/${doc!.id}`, {
        method: "PUT",
        body: JSON.stringify({ title: editTitle, content: editContent }),
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
        {!editing && (
          <header className="mb-6 border-b border-default pb-5">
            <div className="flex flex-wrap items-start justify-between gap-3">
              <div>
                <p className="text-xs text-subtle">{doc.path}</p>
                <h1 className="mt-1 text-2xl font-bold text-fg sm:text-3xl">{doc.title}</h1>
              </div>
              {canEdit && (
                <div className="flex items-center gap-1">
                  <DocumentHistory documentId={doc.id} />
                  <button type="button" className="btn-secondary" onClick={startEdit}>
                    <Edit3 className="h-4 w-4" />
                    {t("document.edit")}
                  </button>
                </div>
              )}
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
              onSave={() => saveDoc.mutate()}
              onPublishPR={async (input) => {
                await publishPR.mutateAsync(input);
              }}
              onCancel={() => setEditing(false)}
            />
          </>
        ) : (
          <MarkdownRenderer
            content={doc.content}
            spaceSlug={slug}
            documents={allDocs?.items ?? []}
            docPath={doc.path}
          />
        )}
      </article>
    </FadeIn>
  );
}
