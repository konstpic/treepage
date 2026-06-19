import { useEffect, useState } from "react";
import { GitPullRequest, Loader2, X } from "lucide-react";
import { useI18n } from "@/lib/i18n";

export interface PublishPRInput {
  branch: string;
  commitMessage: string;
  prTitle: string;
  prBody: string;
}

interface PublishPRDialogProps {
  open: boolean;
  docTitle: string;
  docPath: string;
  defaultBranch: string;
  publishing?: boolean;
  error?: string;
  onClose: () => void;
  onPublish: (input: PublishPRInput) => void;
}

function defaultBranchName(path: string): string {
  const slug = path
    .replace(/\.md$/i, "")
    .replace(/[^a-z0-9]+/gi, "-")
    .replace(/^-+|-+$/g, "")
    .toLowerCase()
    .slice(0, 40);
  return `treepage/docs-${slug || "update"}`;
}

export function PublishPRDialog({
  open,
  docTitle,
  docPath,
  defaultBranch,
  publishing,
  error,
  onClose,
  onPublish,
}: PublishPRDialogProps) {
  const { t } = useI18n();
  const [branch, setBranch] = useState("");
  const [commitMessage, setCommitMessage] = useState("");
  const [prTitle, setPrTitle] = useState("");
  const [prBody, setPrBody] = useState("");

  useEffect(() => {
    if (!open) return;
    setBranch(defaultBranchName(docPath));
    setCommitMessage(t("documentEditor.defaultCommitMessage", { path: docPath }));
    setPrTitle(t("documentEditor.defaultPrTitle", { title: docTitle }));
    setPrBody(t("documentEditor.defaultPrBody", { path: docPath }));
  }, [open, docPath, docTitle, t]);

  if (!open) return null;

  return (
    <div className="fixed inset-0 z-[1000] flex items-center justify-center p-4">
      <button
        type="button"
        className="absolute inset-0 bg-black/40"
        aria-label={t("common.cancel")}
        onClick={onClose}
      />
      <div className="relative z-10 w-full max-w-lg rounded-2xl border border-default bg-surface p-6 shadow-xl">
        <div className="mb-4 flex items-start justify-between gap-3">
          <div>
            <h2 className="text-lg font-semibold text-fg">{t("documentEditor.publishPr")}</h2>
            <p className="mt-1 text-sm text-muted">{t("documentEditor.publishPrHint")}</p>
          </div>
          <button type="button" className="btn-ghost p-1" onClick={onClose}>
            <X className="h-5 w-5" />
          </button>
        </div>

        {error && <p className="mb-3 text-sm text-danger-soft">{error}</p>}

        <div className="space-y-3">
          <label className="block text-sm">
            <span className="mb-1 block text-muted">{t("documentEditor.targetBranch")}</span>
            <span className="mb-1 block text-xs text-subtle">
              {t("documentEditor.baseBranch", { branch: defaultBranch })}
            </span>
            <input
              className="input-field font-mono text-sm"
              value={branch}
              onChange={(e) => setBranch(e.target.value)}
              placeholder="treepage/docs-my-page"
            />
          </label>
          <label className="block text-sm">
            <span className="mb-1 block text-muted">{t("documentEditor.commitMessage")}</span>
            <input
              className="input-field font-mono text-sm"
              value={commitMessage}
              onChange={(e) => setCommitMessage(e.target.value)}
            />
          </label>
          <label className="block text-sm">
            <span className="mb-1 block text-muted">{t("documentEditor.prTitle")}</span>
            <input
              className="input-field text-sm"
              value={prTitle}
              onChange={(e) => setPrTitle(e.target.value)}
            />
          </label>
          <label className="block text-sm">
            <span className="mb-1 block text-muted">{t("documentEditor.prBody")}</span>
            <textarea
              className="input-field min-h-[5rem] resize-y text-sm"
              value={prBody}
              onChange={(e) => setPrBody(e.target.value)}
            />
          </label>
        </div>

        <div className="mt-5 flex flex-wrap justify-end gap-2">
          <button type="button" className="btn-secondary" disabled={publishing} onClick={onClose}>
            {t("common.cancel")}
          </button>
          <button
            type="button"
            className="btn-primary"
            disabled={publishing || !branch.trim() || !commitMessage.trim()}
            onClick={() =>
              onPublish({
                branch: branch.trim(),
                commitMessage: commitMessage.trim(),
                prTitle: prTitle.trim() || commitMessage.trim(),
                prBody: prBody.trim(),
              })
            }
          >
            {publishing ? (
              <Loader2 className="h-4 w-4 animate-spin" />
            ) : (
              <GitPullRequest className="h-4 w-4" />
            )}
            {t("documentEditor.publishPrSubmit")}
          </button>
        </div>
      </div>
    </div>
  );
}
