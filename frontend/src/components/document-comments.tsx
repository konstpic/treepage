import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Loader2, MessageSquare } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { useLocation } from "react-router-dom";
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
  mention_labels?: { email: string; display_name: string }[];
  created_at: string;
  replies?: Comment[];
}

const mentionInBodyRe = /@([a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+)/g;

function CommentBody({
  body,
  mentionLabels,
}: {
  body: string;
  mentionLabels?: Comment["mention_labels"];
}) {
  const nameByEmail = new Map(
    (mentionLabels ?? []).map((l) => [l.email.toLowerCase(), l.display_name || l.email]),
  );

  const parts: React.ReactNode[] = [];
  let last = 0;
  let match: RegExpExecArray | null;
  const re = new RegExp(mentionInBodyRe);
  while ((match = re.exec(body)) !== null) {
    const idx = match.index;
    if (idx > last) parts.push(body.slice(last, idx));
    const email = match[1];
    const label = nameByEmail.get(email.toLowerCase()) ?? email;
    parts.push(
      <span key={`${idx}-${email}`} className="font-medium text-primary">
        @{label}
      </span>,
    );
    last = idx + match[0].length;
  }
  if (last < body.length) parts.push(body.slice(last));
  if (parts.length === 0) return <>{body}</>;
  return <>{parts}</>;
}

function countComments(items: Comment[]): number {
  return items.reduce((n, c) => n + 1 + countComments(c.replies ?? []), 0);
}

function railWidthClass(count: number): string {
  if (count === 0) return "lg:w-[13.5rem]";
  if (count <= 2) return "lg:w-[16rem]";
  if (count <= 5) return "lg:w-[19rem]";
  if (count <= 10) return "lg:w-[22rem]";
  return "lg:w-[25rem]";
}

interface DocumentCommentsProps {
  documentId: string;
  variant?: "inline" | "rail";
  className?: string;
}

export function DocumentComments({ documentId, variant = "inline", className }: DocumentCommentsProps) {
  const { t } = useI18n();
  const location = useLocation();
  const { isAuthenticated } = useAuthStore();
  const qc = useQueryClient();
  const [body, setBody] = useState("");
  const [error, setError] = useState("");

  const { data, isLoading } = useQuery({
    queryKey: ["comments", documentId],
    queryFn: () => api<{ items: Comment[] }>(`/api/documents/${documentId}/comments`),
    enabled: isAuthenticated && !!documentId,
  });

  const commentCount = useMemo(() => countComments(data?.items ?? []), [data?.items]);

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

  useEffect(() => {
    if (isLoading || !data?.items?.length) return;
    const hash = location.hash;
    if (!hash.startsWith("#comment-")) return;
    const el = document.getElementById(hash.slice(1));
    if (!el) return;
    el.scrollIntoView({ behavior: "smooth", block: "center" });
    el.classList.add("comment-highlight");
    const timer = window.setTimeout(() => el.classList.remove("comment-highlight"), 2500);
    return () => window.clearTimeout(timer);
  }, [isLoading, data, location.hash]);

  if (!isAuthenticated) return null;

  function renderComment(c: Comment, depth = 0) {
    return (
      <div key={c.id} className={depth > 0 ? "ml-2 mt-2 border-l border-default pl-2" : ""}>
        <div
          id={`comment-${c.id}`}
          className="rounded-lg bg-surface-muted px-2.5 py-2 text-sm scroll-mt-24"
        >
          <div className="flex flex-wrap items-baseline justify-between gap-x-2 gap-y-0.5">
            <span className="text-xs font-medium text-fg">{c.author_name || t("comments.anonymous")}</span>
            <span className="text-[10px] text-subtle">{formatDate(c.created_at)}</span>
          </div>
          <p className="mt-1 whitespace-pre-wrap break-words text-xs leading-relaxed text-fg">
            <CommentBody body={c.body} mentionLabels={c.mention_labels} />
          </p>
        </div>
        {c.replies?.map((r) => renderComment(r, depth + 1))}
      </div>
    );
  }

  const isRail = variant === "rail";

  const compose = (
    <div className="shrink-0 space-y-2">
      <MentionTextarea
        value={body}
        onChange={setBody}
        placeholder={t("comments.placeholder")}
        minRows={isRail ? 2 : 4}
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
      <p className="text-[10px] leading-snug text-subtle">{t("comments.mentionHint")}</p>
    </div>
  );

  return (
    <aside
      className={cn(
        isRail
          ? cn(
              "glass flex w-full flex-col rounded-xl p-3 transition-[width] duration-300 ease-out lg:sticky lg:top-24 lg:max-h-[calc(100vh-7rem)] lg:shrink-0",
              railWidthClass(commentCount),
            )
          : "mt-8 border-t border-default pt-6",
        className,
      )}
      data-tour="doc-comments"
    >
      <h3 className="mb-2 flex shrink-0 items-center gap-1.5 text-xs font-semibold uppercase tracking-wide text-subtle">
        <MessageSquare className="h-3.5 w-3.5 text-primary" />
        {t("comments.title")}
        {commentCount > 0 && (
          <span className="rounded-full bg-surface-muted px-1.5 py-0.5 text-[10px] font-medium text-muted">
            {commentCount}
          </span>
        )}
      </h3>

      {compose}

      <div
        className={cn(
          "mt-3 min-h-0 flex-1 space-y-2 overflow-y-auto",
          isRail && commentCount > 0 && "border-t border-default pt-3",
        )}
      >
        {isLoading ? (
          <div className="flex justify-center py-4">
            <Loader2 className="h-4 w-4 animate-spin text-primary" />
          </div>
        ) : (
          <>
            {(data?.items ?? []).map((c) => renderComment(c))}
            {!data?.items?.length && isRail && (
              <p className="text-[11px] text-muted">{t("comments.empty")}</p>
            )}
            {!data?.items?.length && !isRail && (
              <p className="text-xs text-muted">{t("comments.empty")}</p>
            )}
          </>
        )}
      </div>
    </aside>
  );
}
