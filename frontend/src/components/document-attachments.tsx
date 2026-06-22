import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Loader2, Paperclip, Trash2, Upload } from "lucide-react";
import { useRef } from "react";
import { api, ApiError } from "@/lib/api";
import { useAuthStore } from "@/lib/store";
import { useI18n } from "@/lib/i18n";
import { formatDate } from "@/lib/utils";

interface Attachment {
  id: string;
  filename: string;
  mime_type: string;
  size_bytes: number;
  created_at: string;
}

interface DocumentAttachmentsProps {
  documentId: string;
  canEdit: boolean;
}

function formatBytes(n: number) {
  if (n < 1024) return `${n} B`;
  if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KB`;
  return `${(n / (1024 * 1024)).toFixed(1)} MB`;
}

export function DocumentAttachments({ documentId, canEdit }: DocumentAttachmentsProps) {
  const { t } = useI18n();
  const qc = useQueryClient();
  const inputRef = useRef<HTMLInputElement>(null);

  const { isAuthenticated } = useAuthStore();
  const { data, isLoading } = useQuery({
    queryKey: ["attachments", documentId],
    queryFn: () => api<{ items: Attachment[] }>(`/api/documents/${documentId}/attachments`),
    enabled: isAuthenticated && !!documentId,
  });

  const upload = useMutation({
    mutationFn: async (file: File) => {
      const form = new FormData();
      form.append("file", file);
      const token = localStorage.getItem("access_token");
      const res = await fetch(`/api/documents/${documentId}/attachments`, {
        method: "POST",
        headers: token ? { Authorization: `Bearer ${token}` } : {},
        body: form,
      });
      if (!res.ok) {
        const body = await res.json().catch(() => ({}));
        throw new ApiError(body.error || res.statusText, res.status);
      }
      return res.json();
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ["attachments", documentId] }),
  });

  const remove = useMutation({
    mutationFn: (id: string) => api(`/api/attachments/${id}`, { method: "DELETE" }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["attachments", documentId] }),
  });

  const items = data?.items ?? [];
  if (!canEdit && items.length === 0) return null;

  return (
    <section className="mt-8 border-t border-default pt-6">
      <div className="mb-3 flex flex-wrap items-center justify-between gap-2">
        <h3 className="flex items-center gap-2 text-sm font-semibold text-fg">
          <Paperclip className="h-4 w-4" />
          {t("attachments.title")}
        </h3>
        {canEdit && (
          <>
            <input
              ref={inputRef}
              type="file"
              className="hidden"
              onChange={(e) => {
                const file = e.target.files?.[0];
                if (file) upload.mutate(file);
                e.target.value = "";
              }}
            />
            <button
              type="button"
              className="btn-secondary !py-1.5 !text-xs"
              disabled={upload.isPending}
              onClick={() => inputRef.current?.click()}
            >
              {upload.isPending ? (
                <Loader2 className="h-3.5 w-3.5 animate-spin" />
              ) : (
                <Upload className="h-3.5 w-3.5" />
              )}
              {t("attachments.upload")}
            </button>
          </>
        )}
      </div>
      {upload.isError && (
        <p className="mb-2 text-xs text-danger-soft">
          {upload.error instanceof ApiError ? upload.error.message : t("common.failed")}
        </p>
      )}
      {isLoading ? (
        <Loader2 className="h-5 w-5 animate-spin text-primary" />
      ) : items.length === 0 ? (
        <p className="text-xs text-muted">{t("attachments.empty")}</p>
      ) : (
        <ul className="space-y-2">
          {items.map((a) => (
            <li key={a.id} className="flex items-center justify-between gap-2 rounded-lg bg-surface-muted px-3 py-2 text-sm">
              <a
                href={`/api/attachments/${a.id}/download`}
                target="_blank"
                rel="noopener noreferrer"
                className="truncate text-primary hover:underline"
              >
                {a.filename}
              </a>
              <span className="shrink-0 text-xs text-subtle">
                {formatBytes(a.size_bytes)} · {formatDate(a.created_at)}
              </span>
              {canEdit && (
                <button
                  type="button"
                  className="btn-ghost !p-1 text-danger-soft"
                  disabled={remove.isPending}
                  onClick={() => remove.mutate(a.id)}
                >
                  <Trash2 className="h-3.5 w-3.5" />
                </button>
              )}
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}
