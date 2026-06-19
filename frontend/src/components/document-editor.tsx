import { useState } from "react";
import { Eye, GitPullRequest, Loader2, Save } from "lucide-react";
import { MarkdownRenderer } from "@/components/markdown-renderer";
import { MarkdownEditor } from "@/components/markdown-editor";
import { PublishPRDialog, type PublishPRInput } from "@/components/publish-pr-dialog";
import { useI18n } from "@/lib/i18n";
import type { LinkDoc } from "@/lib/wiki-markdown";

interface DocumentEditorProps {
  title: string;
  content: string;
  path: string;
  spaceSlug: string;
  gitHint?: string;
  gitLinked?: boolean;
  defaultBranch?: string;
  documents?: LinkDoc[];
  saving?: boolean;
  publishing?: boolean;
  publishError?: string;
  onTitleChange: (title: string) => void;
  onContentChange: (content: string) => void;
  onSave: () => void;
  onSaveDraft?: () => void;
  onPublishLocal?: () => void;
  onSubmitReview?: () => void;
  onApproveReview?: () => void;
  onPublishPR?: (input: PublishPRInput) => void | Promise<void>;
  onCancel: () => void;
}

export function DocumentEditor({
  title,
  content,
  path,
  spaceSlug,
  gitHint,
  gitLinked,
  defaultBranch = "main",
  documents = [],
  saving,
  publishing,
  publishError,
  onTitleChange,
  onContentChange,
  onSave,
  onSaveDraft,
  onPublishLocal,
  onSubmitReview,
  onApproveReview,
  onPublishPR,
  onCancel,
}: DocumentEditorProps) {
  const { t } = useI18n();
  const [preview, setPreview] = useState(false);
  const [prOpen, setPrOpen] = useState(false);

  return (
    <div className="space-y-4">
      {gitHint && (
        <div className="rounded-xl border border-default bg-surface-muted px-4 py-3 text-sm text-muted">
          {t("document.gitHintPublish", { path: gitHint })}
        </div>
      )}
      <div className="flex flex-wrap items-center gap-2">
        <button type="button" className="btn-primary" disabled={saving || publishing} onClick={onSave}>
          {saving ? <Loader2 className="h-4 w-4 animate-spin" /> : <Save className="h-4 w-4" />}
          {t("documentEditor.saveLocal")}
        </button>
        {onSaveDraft && (
          <button type="button" className="btn-secondary" disabled={saving || publishing} onClick={onSaveDraft}>
            {t("documentEditor.saveDraft")}
          </button>
        )}
        {onSubmitReview && (
          <button type="button" className="btn-secondary" disabled={saving || publishing} onClick={onSubmitReview}>
            {t("workflow.submitReview")}
          </button>
        )}
        {onApproveReview && (
          <button type="button" className="btn-secondary" disabled={saving || publishing} onClick={onApproveReview}>
            {t("workflow.approve")}
          </button>
        )}
        {onPublishLocal && (
          <button type="button" className="btn-secondary" disabled={saving || publishing} onClick={onPublishLocal}>
            {t("documentEditor.publishLocal")}
          </button>
        )}
        {gitLinked && onPublishPR && (
          <button
            type="button"
            className="btn-secondary"
            disabled={saving || publishing}
            onClick={() => setPrOpen(true)}
          >
            <GitPullRequest className="h-4 w-4" />
            {t("documentEditor.publishPr")}
          </button>
        )}
        <button type="button" className="btn-secondary" disabled={saving || publishing} onClick={onCancel}>
          {t("common.cancel")}
        </button>
        <button
          type="button"
          className="btn-ghost ml-auto"
          onClick={() => setPreview((p) => !p)}
        >
          <Eye className="h-4 w-4" />
          {preview ? t("document.editMode") : t("document.previewMode")}
        </button>
      </div>
      <input
        className="input-field text-lg font-semibold"
        value={title}
        onChange={(e) => onTitleChange(e.target.value)}
        placeholder={t("document.titlePlaceholder")}
      />
      <p className="text-xs text-subtle">{path}</p>
      {preview ? (
        <div className="glass min-h-[20rem] p-6">
          <MarkdownRenderer
            content={content}
            spaceSlug={spaceSlug}
            documents={documents}
            docPath={path}
          />
        </div>
      ) : (
        <MarkdownEditor
          value={content}
          onChange={onContentChange}
          documents={documents}
          ariaLabel={t("document.contentLabel")}
        />
      )}

      <PublishPRDialog
        open={prOpen}
        docTitle={title}
        docPath={path}
        defaultBranch={defaultBranch}
        publishing={publishing}
        error={publishError}
        onClose={() => setPrOpen(false)}
        onPublish={async (input) => {
          await onPublishPR?.(input);
          setPrOpen(false);
        }}
      />
    </div>
  );
}
