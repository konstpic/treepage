import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Loader2, MessageSquare } from "lucide-react";
import { useState } from "react";
import { api, ApiError } from "@/lib/api";
import { MentionTextarea } from "@/components/mention-textarea";
import { useAuthStore } from "@/lib/store";
import { useI18n } from "@/lib/i18n";
import { formatDate } from "@/lib/utils";
import { cn } from "@/lib/utils";

interface Comment {
  id: string;
  author_id?: string;
  author_name?: string;
  body: string;
  created_at: string;
  replies?: Comment[];
}

interface DocumentCommentsProps {
  documentId: string;
  variant?: "inline" | "sidebar";
  className?: string;
}

export function DocumentComments({ documentId, variant = "inline", className }: DocumentCommentsProps) {
  const { t } = useI18n();
  const { isAuthenticated } = useAuthStore();
  const qc = useQueryClient();
  const [body, setBody] = useState("");
  const [error, setError] = useState("");

  const { data, isLoading } = useQuery({
    queryKey: ["comments", documentId],
    queryFn: () => api<{ items: Comment[] }>(`/api/documents/${documentId}/comments`),
    enabled: isAuthenticated && !!documentId,
  });

  const addComment = useMutation({
    mutationFn: () =>
      api(`/api/documents/${documentId}/comments`, {
        method: "POST",
        body: JSON.stringify({ body }),
      }),
    onSuccess: () => {
      setBody("");
      setError("");
      qc.invalidateQueries({ queryKey: ["comments", documentId] });
    },
    onError: (e) => setError(e instanceof ApiError ? e.message : t("common.failed")),
  });

  if (!isAuthenticated) return null;

  function renderComment(c: Comment, depth = 0) {
    return (
      <div key={c.id} className={depth > 0 ? "ml-3 mt-2 border-l border-default pl-2" : ""}>
        <div className="rounded-lg bg-surface-muted px-3 py-2 text-sm">
          <div className="flex items-center justify-between gap-2">
            <span className="font-medium text-fg">{c.author_name || t("comments.anonymous")}</span>
            <span className="text-xs text-subtle">{formatDate(c.created_at)}</span>
          </div>
          <p className="mt-1 whitespace-pre-wrap break-words text-fg">{c.body}</p>
        </div>
        {c.replies?.map((r) => renderComment(r, depth + 1))}
      </div>
    );
  }

  const isSidebar = variant === "sidebar";

  return (
    <section
      className={cn(
        isSidebar
          ? "flex h-full min-h-[320px] flex-col rounded-xl border border-default bg-surface p-4 lg:sticky lg:top-4 lg:max-h-[calc(100vh-6rem)]"
          : "mt-8 border-t border-default pt-6",
        className,
      )}
      data-tour="doc-comments"
    >
      <h3 className="mb-3 flex shrink-0 items-center gap-2 text-sm font-semibold text-fg">
        <MessageSquare className="h-4 w-4" />
        {t("comments.title")}
      </h3>
      <div className={cn("min-h-0 flex-1 space-y-3 overflow-y-auto", isSidebar && "pr-1")}>
        {isLoading ? (
          <Loader2 className="h-5 w-5 animate-spin text-primary" />
        ) : (
          <>
            {(data?.items ?? []).map((c) => renderComment(c))}
            {!data?.items?.length && <p className="text-xs text-muted">{t("comments.empty")}</p>}
          </>
        )}
      </div>
      <div className={cn("mt-4 shrink-0 space-y-2", isSidebar && "border-t border-default pt-3")}>
        <MentionTextarea
          value={body}
          onChange={setBody}
          placeholder={t("comments.placeholder")}
          minRows={isSidebar ? 3 : 4}
        />
        {error && <p className="text-xs text-danger-soft">{error}</p>}
        <button
          type="button"
          className="btn-secondary w-full !text-xs"
          disabled={!body.trim() || addComment.isPending}
          onClick={() => addComment.mutate()}
        >
          {addComment.isPending ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : t("comments.post")}
        </button>
        <p className="text-[10px] text-subtle">{t("comments.mentionHint")}</p>
      </div>
    </section>
  );
}
