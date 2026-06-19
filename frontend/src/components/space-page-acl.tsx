import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Loader2, Shield } from "lucide-react";
import { useState } from "react";
import { api, ApiError } from "@/lib/api";
import { SelectField } from "@/components/select-field";
import { useI18n } from "@/lib/i18n";

interface ACLRule {
  id: string;
  path_prefix: string;
  subject_type: string;
  subject_id: string;
  role: string;
}

interface UserOption {
  id: string;
  email: string;
  display_name: string;
}

interface GroupOption {
  id: string;
  name: string;
}

export function SpacePageACLPanel({ spaceId }: { spaceId: string }) {
  const { t } = useI18n();
  const qc = useQueryClient();
  const [form, setForm] = useState({
    path_prefix: "",
    subject_type: "user",
    subject_id: "",
    role: "viewer",
  });
  const [error, setError] = useState("");

  const { data: rules, isLoading } = useQuery({
    queryKey: ["page-acl", spaceId],
    queryFn: () => api<{ items: ACLRule[] }>(`/api/admin/spaces/${spaceId}/page-acl`),
    enabled: !!spaceId,
  });

  const { data: users } = useQuery({
    queryKey: ["admin-users-acl"],
    queryFn: () => api<{ items: UserOption[] }>("/api/admin/users"),
    enabled: form.subject_type === "user",
  });

  const { data: groups } = useQuery({
    queryKey: ["admin-groups-acl"],
    queryFn: () => api<{ items: GroupOption[] }>("/api/admin/groups"),
    enabled: form.subject_type === "group",
  });

  const createRule = useMutation({
    mutationFn: () =>
      api(`/api/admin/spaces/${spaceId}/page-acl`, { method: "POST", body: JSON.stringify(form) }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["page-acl", spaceId] });
      setError("");
      setForm({ path_prefix: "", subject_type: "user", subject_id: "", role: "viewer" });
    },
    onError: (e) => setError(e instanceof ApiError ? e.message : t("common.failed")),
  });

  const removeRule = useMutation({
    mutationFn: (id: string) => api(`/api/admin/page-acl/${id}`, { method: "DELETE" }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["page-acl", spaceId] }),
  });

  return (
    <div className="mt-6 rounded-xl border border-default p-4">
      <h3 className="flex items-center gap-2 font-semibold text-fg">
        <Shield className="h-4 w-4" />
        {t("admin.pageAcl.title")}
      </h3>
      <p className="mt-1 text-xs text-muted">{t("admin.pageAcl.hint")}</p>
      {isLoading ? (
        <Loader2 className="mt-4 h-5 w-5 animate-spin text-primary" />
      ) : (
        <ul className="mt-3 space-y-1 text-sm">
          {(rules?.items ?? []).map((r) => (
            <li key={r.id} className="flex items-center justify-between gap-2 rounded bg-surface-muted px-2 py-1">
              <span>
                <code className="text-xs">{r.path_prefix || "*"}</code> → {r.subject_type}/{r.role}
              </span>
              <button type="button" className="btn-ghost !p-1 text-danger-soft" onClick={() => removeRule.mutate(r.id)}>
                ×
              </button>
            </li>
          ))}
          {!rules?.items?.length && <li className="text-muted">{t("admin.pageAcl.empty")}</li>}
        </ul>
      )}
      <div className="mt-4 grid gap-2 sm:grid-cols-2">
        <input
          className="input-field sm:col-span-2"
          placeholder={t("admin.pageAcl.pathPrefix")}
          value={form.path_prefix}
          onChange={(e) => setForm({ ...form, path_prefix: e.target.value })}
        />
        <SelectField value={form.subject_type} onChange={(e) => setForm({ ...form, subject_type: e.target.value, subject_id: "" })}>
          <option value="user">{t("admin.pageAcl.user")}</option>
          <option value="group">{t("admin.pageAcl.group")}</option>
        </SelectField>
        <SelectField
          value={form.subject_id}
          onChange={(e) => setForm({ ...form, subject_id: e.target.value })}
          required
        >
          <option value="">{t("admin.pageAcl.selectSubject")}</option>
          {form.subject_type === "user"
            ? users?.items.map((u) => (
                <option key={u.id} value={u.id}>
                  {u.display_name || u.email}
                </option>
              ))
            : groups?.items.map((g) => (
                <option key={g.id} value={g.id}>
                  {g.name}
                </option>
              ))}
        </SelectField>
        <SelectField value={form.role} onChange={(e) => setForm({ ...form, role: e.target.value })}>
          <option value="viewer">viewer</option>
          <option value="editor">editor</option>
          <option value="admin">admin</option>
          <option value="none">{t("admin.pageAcl.deny")}</option>
        </SelectField>
        <button type="button" className="btn-secondary" disabled={createRule.isPending || !form.subject_id} onClick={() => createRule.mutate()}>
          {t("admin.pageAcl.add")}
        </button>
      </div>
      {error && <p className="mt-2 text-xs text-danger-soft">{error}</p>}
    </div>
  );
}
