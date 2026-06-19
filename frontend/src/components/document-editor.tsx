import { useState } from "react";
import { Eye, Loader2, Save } from "lucide-react";
import { MarkdownRenderer } from "@/components/markdown-renderer";
import { useI18n } from "@/lib/i18n";

interface DocumentEditorProps {
  title: string;
  content: string;
  path: string;
  spaceSlug: string;
  gitHint?: string;
  saving?: boolean;
  onTitleChange: (title: string) => void;
  onContentChange: (content: string) => void;
  onSave: () => void;
  onCancel: () => void;
}

export function DocumentEditor({
  title,
  content,
  path,
  spaceSlug,
  gitHint,
  saving,
  onTitleChange,
  onContentChange,
  onSave,
  onCancel,
}: DocumentEditorProps) {
  const { t } = useI18n();
  const [preview, setPreview] = useState(false);

  return (
    <div className="space-y-4">
      {gitHint && (
        <div className="rounded-xl border border-default bg-surface-muted px-4 py-3 text-sm text-muted">
          {t("document.gitHint", { path: gitHint })}
        </div>
      )}
      <div className="flex flex-wrap items-center gap-2">
        <button type="button" className="btn-primary" disabled={saving} onClick={onSave}>
          {saving ? <Loader2 className="h-4 w-4 animate-spin" /> : <Save className="h-4 w-4" />}
          {t("common.save")}
        </button>
        <button type="button" className="btn-secondary" onClick={onCancel}>
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
          <MarkdownRenderer content={content} spaceSlug={spaceSlug} documents={[]} docPath={path} />
        </div>
      ) : (
        <textarea
          className="input-field min-h-[28rem] resize-y font-mono text-sm leading-relaxed"
          value={content}
          onChange={(e) => onContentChange(e.target.value)}
          spellCheck={false}
          aria-label={t("document.contentLabel")}
        />
      )}
    </div>
  );
}
