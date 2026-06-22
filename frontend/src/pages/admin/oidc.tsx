import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Loader2, Trash2 } from "lucide-react";
import { api, ApiError } from "@/lib/api";
import { FadeIn } from "@/components/motion-wrapper";
import { useAdminGuard } from "./layout";
import { useI18n } from "@/lib/i18n";

const configManagedName = "Authentik (config)";

interface OIDCProvider {
  id: string;
  name: string;
  provider_type: string;
  issuer_url: string;
  client_id: string;
  redirect_url: string;
  scopes: string;
  role_claim?: string;
  group_claim?: string;
  sync_groups?: boolean;
  enabled: boolean;
}

export function AdminOIDCPage() {
  const { ready } = useAdminGuard();
  const { t } = useI18n();
  const qc = useQueryClient();
  const [error, setError] = useState("");

  const [name, setName] = useState("");
  const [issuerUrl, setIssuerUrl] = useState("");
  const [clientId, setClientId] = useState("");
  const [redirectUrl, setRedirectUrl] = useState("");
  const [scopes, setScopes] = useState("openid profile email");
  const [roleClaim, setRoleClaim] = useState("roles");
  const [groupClaim, setGroupClaim] = useState("groups");
  const [syncGroups, setSyncGroups] = useState(true);

  const { data, isLoading } = useQuery({
    queryKey: ["admin-oidc"],
    queryFn: () => api<{ items: OIDCProvider[] }>("/api/admin/oidc-providers"),
    enabled: ready,
  });

  const create = useMutation({
    mutationFn: (body: Record<string, unknown>) =>
      api("/api/admin/oidc-providers", { method: "POST", body: JSON.stringify(body) }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["admin-oidc"] });
      setName("");
      setIssuerUrl("");
      setClientId("");
      setRedirectUrl("");
      setError("");
    },
    onError: (e) => setError(e instanceof ApiError ? e.message : "Failed"),
  });

  const remove = useMutation({
    mutationFn: (id: string) => api(`/api/admin/oidc-providers/${id}`, { method: "DELETE" }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["admin-oidc"] }),
    onError: (e) => setError(e instanceof ApiError ? e.message : "Failed"),
  });

  if (!ready) return null;

  return (
    <FadeIn>
      <div className="glass p-6">
        <h2 className="text-lg font-semibold text-fg">{t("admin.nav.oidc")}</h2>
        <p className="mt-1 text-sm text-muted">{t("admin.oidcPageHint")}</p>

        {isLoading ? (
          <div className="flex justify-center py-12">
            <Loader2 className="h-6 w-6 animate-spin text-primary" />
          </div>
        ) : (
          <div className="mt-6 space-y-2">
            {data?.items.map((p) => (
              <div
                key={p.id}
                className="flex items-start justify-between gap-4 rounded-xl border border-default bg-surface-muted px-4 py-3"
              >
                <div>
                  <div className="flex flex-wrap items-center gap-2">
                    <p className="font-medium text-fg">{p.name}</p>
                    {p.name === configManagedName && (
                      <span className="badge badge-neutral">{t("admin.oidcConfigBadge")}</span>
                    )}
                  </div>
                  <p className="text-xs text-subtle">{p.issuer_url}</p>
                  <p className="mt-1 text-xs text-subtle">{t("admin.oidcClient")}: {p.client_id}</p>
                  {p.name === configManagedName && (
                    <p className="mt-2 text-xs text-muted">{t("admin.oidcConfigHint")}</p>
                  )}
                  <p className="mt-1 text-xs text-subtle">
                    Claims: {p.role_claim || "roles"} / {p.group_claim || "groups"}
                    {p.sync_groups && " · sync groups"}
                  </p>
                </div>
                <div className="flex items-center gap-2">
                  <span
                    className={
                      p.enabled
                        ? "badge badge-success"
                        : "badge badge-neutral"
                    }
                  >
                    {p.enabled ? "enabled" : "disabled"}
                  </span>
                  <button
                    type="button"
                    className="btn-ghost text-danger-soft hover:text-rose-300"
                    disabled={p.name === configManagedName}
                    title={p.name === configManagedName ? t("admin.oidcConfigHint") : undefined}
                    onClick={() => remove.mutate(p.id)}
                  >
                    <Trash2 className="h-4 w-4" />
                  </button>
                </div>
              </div>
            ))}
            {data?.items.length === 0 && (
              <p className="text-sm text-subtle">{t("admin.oidcNone")}</p>
            )}
          </div>
        )}
      </div>

      <div className="glass mt-6 p-6">
        <h2 className="text-lg font-semibold text-fg">Add Provider</h2>
        <form
          className="mt-4 space-y-4"
          onSubmit={(e) => {
            e.preventDefault();
            create.mutate({
              name,
              issuer_url: issuerUrl,
              client_id: clientId,
              redirect_url: redirectUrl,
              scopes,
              role_claim: roleClaim,
              group_claim: groupClaim,
              sync_groups: syncGroups,
              enabled: true,
            });
          }}
        >
          <input className="input-field" placeholder="Name" value={name} onChange={(e) => setName(e.target.value)} required />
          <input
            className="input-field"
            placeholder="Issuer URL"
            value={issuerUrl}
            onChange={(e) => setIssuerUrl(e.target.value)}
            required
          />
          <input
            className="input-field"
            placeholder="Client ID"
            value={clientId}
            onChange={(e) => setClientId(e.target.value)}
            required
          />
          <input
            className="input-field"
            placeholder="Redirect URL"
            value={redirectUrl}
            onChange={(e) => setRedirectUrl(e.target.value)}
            required
          />
          <input className="input-field" placeholder="Scopes" value={scopes} onChange={(e) => setScopes(e.target.value)} />
          <div className="grid gap-4 sm:grid-cols-2">
            <input
              className="input-field"
              placeholder="Role claim (e.g. roles)"
              value={roleClaim}
              onChange={(e) => setRoleClaim(e.target.value)}
            />
            <input
              className="input-field"
              placeholder="Group claim (e.g. groups)"
              value={groupClaim}
              onChange={(e) => setGroupClaim(e.target.value)}
            />
          </div>
          <label className="checkbox-label">
            <input
              type="checkbox"
              checked={syncGroups}
              onChange={(e) => setSyncGroups(e.target.checked)}
              className="checkbox-field"
            />
            Sync IdP groups to TreePage groups on login
          </label>
          {error && <p className="text-sm text-danger-soft">{error}</p>}
          <button type="submit" className="btn-primary" disabled={create.isPending}>
            {create.isPending ? <Loader2 className="h-4 w-4 animate-spin" /> : "Add Provider"}
          </button>
        </form>
      </div>
    </FadeIn>
  );
}
