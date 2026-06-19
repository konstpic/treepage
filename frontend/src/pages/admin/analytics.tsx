import { useQuery } from "@tanstack/react-query";
import { BarChart3, Loader2 } from "lucide-react";
import { Link } from "react-router-dom";
import { api } from "@/lib/api";
import { FadeIn } from "@/components/motion-wrapper";
import { formatDate } from "@/lib/utils";
import { useI18n } from "@/lib/i18n";
import { useAdminGuard } from "./layout";

interface Overview {
  total_views: number;
  top_documents: {
    document_id: string;
    space_slug: string;
    doc_slug: string;
    title: string;
    view_count: number;
  }[];
  stale_documents: {
    space_slug: string;
    doc_slug: string;
    title: string;
    updated_at: string;
  }[];
  top_searches: { query_text: string; count: number }[];
}

export function AdminAnalyticsPage() {
  const { ready } = useAdminGuard();
  const { t } = useI18n();

  const { data, isLoading } = useQuery({
    queryKey: ["admin-analytics"],
    queryFn: () => api<Overview>("/api/admin/analytics/overview"),
    enabled: ready,
  });

  if (!ready) return null;

  return (
    <FadeIn>
      <div className="mb-6 flex items-center gap-2">
        <BarChart3 className="h-6 w-6 text-primary" />
        <div>
          <h2 className="text-xl font-bold text-fg">{t("admin.analytics.title")}</h2>
          <p className="text-sm text-muted">{t("admin.analytics.subtitle")}</p>
        </div>
      </div>

      {isLoading ? (
        <div className="flex justify-center py-16">
          <Loader2 className="h-8 w-8 animate-spin text-primary" />
        </div>
      ) : (
        <div className="space-y-6">
          <div className="glass p-5">
            <p className="text-sm text-muted">{t("admin.analytics.totalViews")}</p>
            <p className="text-3xl font-bold text-fg">{data?.total_views ?? 0}</p>
          </div>

          <div className="glass p-5">
            <h3 className="font-semibold text-fg">{t("admin.analytics.topDocuments")}</h3>
            <ul className="mt-3 space-y-2 text-sm">
              {(data?.top_documents ?? []).map((d) => (
                <li key={d.document_id} className="flex justify-between gap-2">
                  <Link to={`/spaces/${d.space_slug}/docs/${d.doc_slug}`} className="text-primary hover:underline">
                    {d.title}
                  </Link>
                  <span className="text-subtle">{d.view_count}</span>
                </li>
              ))}
              {!data?.top_documents?.length && <li className="text-muted">{t("admin.analytics.empty")}</li>}
            </ul>
          </div>

          <div className="glass p-5">
            <h3 className="font-semibold text-fg">{t("admin.analytics.staleDocuments")}</h3>
            <ul className="mt-3 space-y-2 text-sm">
              {(data?.stale_documents ?? []).map((d) => (
                <li key={`${d.space_slug}-${d.doc_slug}`} className="flex justify-between gap-2">
                  <Link to={`/spaces/${d.space_slug}/docs/${d.doc_slug}`} className="text-primary hover:underline">
                    {d.title}
                  </Link>
                  <span className="text-subtle">{formatDate(d.updated_at)}</span>
                </li>
              ))}
              {!data?.stale_documents?.length && <li className="text-muted">{t("admin.analytics.empty")}</li>}
            </ul>
          </div>

          <div className="glass p-5">
            <h3 className="font-semibold text-fg">{t("admin.analytics.topSearches")}</h3>
            <ul className="mt-3 space-y-2 text-sm">
              {(data?.top_searches ?? []).map((s) => (
                <li key={s.query_text} className="flex justify-between gap-2">
                  <span>{s.query_text}</span>
                  <span className="text-subtle">{s.count}</span>
                </li>
              ))}
              {!data?.top_searches?.length && <li className="text-muted">{t("admin.analytics.empty")}</li>}
            </ul>
          </div>
        </div>
      )}
    </FadeIn>
  );
}
