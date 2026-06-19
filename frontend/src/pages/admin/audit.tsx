import { useQuery } from "@tanstack/react-query";
import { ClipboardList, Loader2 } from "lucide-react";
import { api } from "@/lib/api";
import { formatDate } from "@/lib/utils";
import { useI18n } from "@/lib/i18n";
import { useAdminGuard } from "./layout";

interface AuditRow {
  id: string;
  action: string;
  resource_type: string;
  resource_id?: string;
  user_email?: string;
  ip_address?: string;
  created_at: string;
}

export function AdminAuditPage() {
  const { t } = useI18n();
  const { ready } = useAdminGuard();

  const { data, isLoading } = useQuery({
    queryKey: ["admin-audit"],
    queryFn: () => api<{ items: AuditRow[]; total: number }>("/api/admin/audit-logs?limit=100"),
    enabled: ready,
  });

  if (!ready) return null;

  return (
    <div className="glass p-6">
      <h2 className="mb-4 flex items-center gap-2 text-xl font-semibold text-fg">
        <ClipboardList className="h-5 w-5 text-primary" />
        {t("admin.audit.title")}
      </h2>
      <p className="mb-4 text-sm text-muted">{t("admin.audit.subtitle")}</p>

      {isLoading ? (
        <div className="flex justify-center py-10">
          <Loader2 className="h-6 w-6 animate-spin text-primary" />
        </div>
      ) : (
        <div className="overflow-x-auto">
          <table className="w-full text-left text-sm">
            <thead>
              <tr className="border-b border-default text-subtle">
                <th className="py-2 pr-3">{t("admin.audit.time")}</th>
                <th className="py-2 pr-3">{t("admin.audit.action")}</th>
                <th className="py-2 pr-3">{t("admin.audit.user")}</th>
                <th className="py-2 pr-3">{t("admin.audit.resource")}</th>
                <th className="py-2">{t("admin.audit.ip")}</th>
              </tr>
            </thead>
            <tbody>
              {data?.items.map((row) => (
                <tr key={row.id} className="border-b border-default/60">
                  <td className="py-2 pr-3 whitespace-nowrap text-subtle">{formatDate(row.created_at)}</td>
                  <td className="py-2 pr-3 font-mono text-xs">{row.action}</td>
                  <td className="py-2 pr-3">{row.user_email ?? "—"}</td>
                  <td className="py-2 pr-3 font-mono text-xs">
                    {row.resource_type}
                    {row.resource_id ? ` · ${row.resource_id.slice(0, 8)}…` : ""}
                  </td>
                  <td className="py-2 font-mono text-xs text-subtle">{row.ip_address ?? "—"}</td>
                </tr>
              ))}
            </tbody>
          </table>
          {data?.items.length === 0 && (
            <p className="py-8 text-center text-sm text-subtle">{t("admin.audit.empty")}</p>
          )}
        </div>
      )}
    </div>
  );
}
