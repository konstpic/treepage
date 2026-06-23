import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { Clock, Loader2, Star } from "lucide-react";
import { api } from "@/lib/api";
import { useAuthStore } from "@/lib/store";
import { useI18n } from "@/lib/i18n";
import { FadeIn } from "@/components/motion-wrapper";
import { formatDate } from "@/lib/utils";

interface FavoriteItem {
  document_id: string;
  space_slug: string;
  space_name: string;
  doc_slug: string;
  title: string;
  created_at: string;
}

interface RecentItem {
  document_id: string;
  space_slug: string;
  space_name: string;
  doc_slug: string;
  title: string;
  viewed_at: string;
}

export function MePage() {
  const { t } = useI18n();
  const { isAuthenticated } = useAuthStore();
  const qc = useQueryClient();

  const { data: favorites, isLoading: favLoading } = useQuery({
    queryKey: ["favorites"],
    queryFn: () => api<{ items: FavoriteItem[] }>("/api/me/favorites"),
    enabled: isAuthenticated,
  });

  const { data: recent, isLoading: recentLoading } = useQuery({
    queryKey: ["recent"],
    queryFn: () => api<{ items: RecentItem[] }>("/api/me/recent"),
    enabled: isAuthenticated,
  });

  const removeFavorite = useMutation({
    mutationFn: (documentId: string) =>
      api(`/api/me/favorites/${documentId}`, { method: "DELETE" }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["favorites"] }),
  });

  if (!isAuthenticated) {
    return (
      <div className="mx-auto max-w-3xl px-4 py-16 text-center">
        <p className="text-muted">{t("me.signInRequired")}</p>
        <Link to="/auth" className="btn-primary mt-4 inline-flex">
          {t("nav.signIn")}
        </Link>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-3xl px-4 py-10 sm:px-6" data-tour="me-main">
      <FadeIn>
        <h1 className="text-2xl font-bold text-fg">{t("me.title")}</h1>
        <p className="mt-1 text-sm text-muted">{t("me.subtitle")}</p>
      </FadeIn>

      <FadeIn delay={0.05} className="mt-8">
        <div className="flex items-center gap-2">
          <Star className="h-5 w-5 text-primary" />
          <h2 className="text-lg font-semibold text-fg">{t("me.favorites")}</h2>
        </div>
        {favLoading ? (
          <div className="flex justify-center py-8">
            <Loader2 className="h-6 w-6 animate-spin text-primary" />
          </div>
        ) : !favorites?.items.length ? (
          <p className="mt-3 text-sm text-muted">{t("me.noFavorites")}</p>
        ) : (
          <ul className="mt-3 space-y-2">
            {favorites.items.map((f) => (
              <li key={f.document_id} className="glass flex items-center justify-between gap-3 px-4 py-3">
                <div className="min-w-0">
                  <Link
                    to={`/spaces/${f.space_slug}/docs/${f.doc_slug}`}
                    className="font-medium text-fg hover:text-primary"
                  >
                    {f.title}
                  </Link>
                  <p className="text-xs text-subtle">
                    {f.space_name} · {formatDate(f.created_at)}
                  </p>
                </div>
                <button
                  type="button"
                  className="btn-ghost shrink-0 text-subtle"
                  onClick={() => removeFavorite.mutate(f.document_id)}
                >
                  <Star className="h-4 w-4 fill-primary text-primary" />
                </button>
              </li>
            ))}
          </ul>
        )}
      </FadeIn>

      <FadeIn delay={0.1} className="mt-10">
        <div className="flex items-center gap-2">
          <Clock className="h-5 w-5 text-primary" />
          <h2 className="text-lg font-semibold text-fg">{t("me.recent")}</h2>
        </div>
        {recentLoading ? (
          <div className="flex justify-center py-8">
            <Loader2 className="h-6 w-6 animate-spin text-primary" />
          </div>
        ) : !recent?.items.length ? (
          <p className="mt-3 text-sm text-muted">{t("me.noRecent")}</p>
        ) : (
          <ul className="mt-3 space-y-2">
            {recent.items.map((r) => (
              <li key={r.document_id} className="glass px-4 py-3">
                <Link
                  to={`/spaces/${r.space_slug}/docs/${r.doc_slug}`}
                  className="font-medium text-fg hover:text-primary"
                >
                  {r.title}
                </Link>
                <p className="text-xs text-subtle">
                  {r.space_name} · {formatDate(r.viewed_at)}
                </p>
              </li>
            ))}
          </ul>
        )}
      </FadeIn>
    </div>
  );
}
