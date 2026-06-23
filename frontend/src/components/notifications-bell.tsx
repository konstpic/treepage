import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Bell, Check, ExternalLink, Loader2 } from "lucide-react";
import { useState } from "react";
import { Link } from "react-router-dom";
import { api } from "@/lib/api";
import { useAuthStore } from "@/lib/store";
import { useI18n } from "@/lib/i18n";
import { cn, formatDate } from "@/lib/utils";

interface Notification {
  id: string;
  type: string;
  title: string;
  body: string;
  link?: string;
  resource_type?: string;
  resource_id?: string;
  read_at?: string;
  created_at: string;
}

function notificationTitle(n: Notification, t: (key: string) => string) {
  if (n.type === "comment.mention") return t("notifications.mentionTitle");
  return n.title;
}

export function NotificationsBell() {
  const { t } = useI18n();
  const { isAuthenticated } = useAuthStore();
  const [open, setOpen] = useState(false);
  const qc = useQueryClient();

  const { data: countData } = useQuery({
    queryKey: ["notifications-unread"],
    queryFn: () => api<{ count: number }>("/api/notifications/unread-count"),
    enabled: isAuthenticated,
    refetchInterval: 60_000,
  });

  const { data, isLoading } = useQuery({
    queryKey: ["notifications"],
    queryFn: () => api<{ items: Notification[] }>("/api/notifications?limit=20"),
    enabled: isAuthenticated && open,
  });

  const markRead = useMutation({
    mutationFn: (id: string) => api(`/api/notifications/${id}/read`, { method: "POST" }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["notifications"] });
      qc.invalidateQueries({ queryKey: ["notifications-unread"] });
    },
  });

  const markAll = useMutation({
    mutationFn: () => api("/api/notifications/read-all", { method: "POST" }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["notifications"] });
      qc.invalidateQueries({ queryKey: ["notifications-unread"] });
    },
  });

  if (!isAuthenticated) return null;

  const unread = countData?.count ?? 0;

  function openNotification(n: Notification) {
    if (!n.read_at) markRead.mutate(n.id);
    setOpen(false);
  }

  return (
    <div className="relative">
      <button
        type="button"
        className="btn-ghost relative"
        aria-label={t("notifications.title")}
        onClick={() => setOpen((o) => !o)}
      >
        <Bell className="h-4 w-4" />
        {unread > 0 && (
          <span className="absolute -right-0.5 -top-0.5 flex h-4 min-w-4 items-center justify-center rounded-full bg-primary px-1 text-[10px] font-bold text-on-primary">
            {unread > 9 ? "9+" : unread}
          </span>
        )}
      </button>
      {open && (
        <>
          <button type="button" className="fixed inset-0 z-40" aria-label={t("common.cancel")} onClick={() => setOpen(false)} />
          <div className="absolute right-0 z-50 mt-2 w-80 rounded-xl border border-default bg-surface shadow-lg sm:w-96">
            <div className="flex items-center justify-between border-b border-default px-4 py-3">
              <span className="font-semibold text-fg">{t("notifications.title")}</span>
              {unread > 0 && (
                <button
                  type="button"
                  className="text-xs text-primary hover:underline"
                  disabled={markAll.isPending}
                  onClick={() => markAll.mutate()}
                >
                  {t("notifications.markAllRead")}
                </button>
              )}
            </div>
            <div className="max-h-80 overflow-y-auto">
              {isLoading ? (
                <div className="flex justify-center py-8">
                  <Loader2 className="h-5 w-5 animate-spin text-primary" />
                </div>
              ) : !data?.items.length ? (
                <p className="px-4 py-6 text-center text-sm text-muted">{t("notifications.empty")}</p>
              ) : (
                data.items.map((n) => {
                  const title = notificationTitle(n, t);
                  const content = (
                    <>
                      <p className="text-sm font-medium text-fg">{title}</p>
                      {n.body && <p className="mt-0.5 text-xs text-muted">{n.body}</p>}
                      {n.link && (
                        <p className="mt-1.5 inline-flex items-center gap-1 text-xs font-medium text-primary">
                          <ExternalLink className="h-3 w-3" />
                          {t("notifications.openComment")}
                        </p>
                      )}
                      <p className="mt-1 text-[10px] text-subtle">{formatDate(n.created_at)}</p>
                    </>
                  );

                  return (
                    <div
                      key={n.id}
                      className={cn(
                        "border-b border-default px-4 py-3 last:border-0",
                        !n.read_at && "bg-surface-muted/50",
                      )}
                    >
                      <div className="flex items-start justify-between gap-2">
                        <div className="min-w-0 flex-1">
                          {n.link ? (
                            <Link
                              to={n.link}
                              className="block rounded-lg transition-colors hover:opacity-90"
                              onClick={() => openNotification(n)}
                            >
                              {content}
                            </Link>
                          ) : (
                            content
                          )}
                        </div>
                        {!n.read_at && (
                          <button
                            type="button"
                            className="btn-ghost !p-1"
                            title={t("notifications.markRead")}
                            onClick={() => markRead.mutate(n.id)}
                          >
                            <Check className="h-3.5 w-3.5" />
                          </button>
                        )}
                      </div>
                    </div>
                  );
                })
              )}
            </div>
            <div className="border-t border-default px-4 py-2 text-center">
              <Link to="/me" className="text-xs text-primary hover:underline" onClick={() => setOpen(false)}>
                {t("notifications.viewAll")}
              </Link>
            </div>
          </div>
        </>
      )}
    </div>
  );
}
