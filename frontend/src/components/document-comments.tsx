import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Loader2, MessageSquare } from "lucide-react";
import { useState } from "react";
import { api, ApiError } from "@/lib/api";
import { useAuthStore } from "@/lib/store";
import { useI18n } from "@/lib/i18n";
import { formatDate } from "@/lib/utils";

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
}

export function DocumentComments({ documentId }: DocumentCommentsProps) {
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
      <div key={c.id} className={depth > 0 ? "ml-4 mt-2 border-l border-default pl-3" : ""}>
        <div className="rounded-lg bg-surface-muted px-3 py-2 text-sm">
          <div className="flex items-center justify-between gap-2">
            <span className="font-medium text-fg">{c.author_name || t("comments.anonymous")}</span>
            <span className="text-xs text-subtle">{formatDate(c.created_at)}</span>
          </div>
          <p className="mt-1 whitespace-pre-wrap text-fg">{c.body}</p>
        </div>
        {c.replies?.map((r) => renderComment(r, depth + 1))}
      </div>
    );
  }

  return (
    <section className="mt-8 border-t border-default pt-6">
      <h3 className="mb-3 flex items-center gap-2 text-sm font-semibold text-fg">
        <MessageSquare className="h-4 w-4" />
        {t("comments.title")}
      </h3>
      {isLoading ? (
        <Loader2 className="h-5 w-5 animate-spin text-primary" />
      ) : (
        <div className="space-y-3">
          {(data?.items ?? []).map((c) => renderComment(c))}
          {!data?.items?.length && <p className="text-xs text-muted">{t("comments.empty")}</p>}
        </div>
      )}
      <div className="mt-4 space-y-2">
        <textarea
          className="input-field min-h-[80px] w-full"
          placeholder={t("comments.placeholder")}
          value={body}
          onChange={(e) => setBody(e.target.value)}
        />
        {error && <p className="text-xs text-danger-soft">{error}</p>}
        <button
          type="button"
          className="btn-secondary !text-xs"
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
