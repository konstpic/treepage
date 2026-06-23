import { useEffect, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Loader2, Pencil, RefreshCw, Trash2, X } from "lucide-react";
import { api, ApiError } from "@/lib/api";
import { FadeIn } from "@/components/motion-wrapper";
import { SelectField } from "@/components/select-field";
import { formatDate } from "@/lib/utils";
import { useI18n } from "@/lib/i18n";
import { useAdminGuard } from "./layout";

interface Space {
  id: string;
  slug: string;
  name: string;
}

interface Repository {
  id: string;
  space_id: string;
  name: string;
  url: string;
  branch: string;
  provider?: string;
  docs_path: string;
  sync_mode: string;
  sync_interval_seconds: number;
  access_token_ref?: string;
  webhook_secret_ref?: string;
  last_sync_at?: string;
  last_sync_status?: string;
  last_sync_error?: string;
  enabled: boolean;
  space_slug: string;
  space_name: string;
}

const SYNC_MODES = ["manual", "scheduled", "webhook"];

const PROVIDERS = ["github", "gitlab", "gitea", "bitbucket", "generic"] as const;

type RepoForm = {
  space_id: string;
  name: string;
  url: string;
  branch: string;
  provider: string;
  docs_path: string;
  sync_mode: string;
  sync_interval_seconds: number;
  access_token_ref: string;
  webhook_secret_ref: string;
  enabled: boolean;
};

const emptyForm = (): RepoForm => ({
  space_id: "",
  name: "",
  url: "",
  branch: "main",
  provider: "generic",
  docs_path: "docs",
  sync_mode: "manual",
  sync_interval_seconds: 300,
  access_token_ref: "",
  webhook_secret_ref: "",
  enabled: true,
});

function repoToForm(repo: Repository): RepoForm {
  return {
    space_id: repo.space_id,
    name: repo.name,
    url: repo.url,
    branch: repo.branch || "main",
    provider: repo.provider || "generic",
    docs_path: repo.docs_path || "docs",
    sync_mode: repo.sync_mode || "manual",
    sync_interval_seconds: repo.sync_interval_seconds || 300,
    access_token_ref: repo.access_token_ref || "",
    webhook_secret_ref: repo.webhook_secret_ref || "",
    enabled: repo.enabled,
  };
}

function syncStatusClass(status?: string) {
  if (status === "success" || status === "completed") {
    return "badge badge-success";
  }
  if (status === "failed") {
    return "badge badge-danger";
  }
  return "badge badge-neutral";
}

function RepositoryFormFields({
  form,
  setForm,
  spaces,
  tokenHint,
}: {
  form: RepoForm;
  setForm: (f: RepoForm) => void;
  spaces?: Space[];
  tokenHint?: string;
}) {
  const { t } = useI18n();
  return (
    <>
      {spaces && (
        <SelectField
          className="sm:col-span-2"
          value={form.space_id}
          onChange={(e) => setForm({ ...form, space_id: e.target.value })}
          required
        >
          <option value="">{t("admin.repositories.selectSpace")}</option>
          {spaces.map((s) => (
            <option key={s.id} value={s.id}>
              {s.name} ({s.slug})
            </option>
          ))}
        </SelectField>
      )}
      <input
        className="input-field"
        placeholder={t("admin.repositories.namePlaceholder")}
        value={form.name}
        onChange={(e) => setForm({ ...form, name: e.target.value })}
        required
      />
      <input
        className="input-field"
        placeholder={t("admin.repositories.branchPlaceholder")}
        value={form.branch}
        onChange={(e) => setForm({ ...form, branch: e.target.value })}
      />
      <input
        className="input-field sm:col-span-2"
        placeholder={t("admin.repositories.urlPlaceholder")}
        value={form.url}
        onChange={(e) => setForm({ ...form, url: e.target.value })}
        required
      />
      <SelectField
        value={form.provider}
        onChange={(e) => setForm({ ...form, provider: e.target.value })}
      >
        {PROVIDERS.map((p) => (
          <option key={p} value={p}>
            {p}
          </option>
        ))}
      </SelectField>
      <input
        className="input-field"
        placeholder={t("admin.repositories.docsPathPlaceholder")}
        value={form.docs_path}
        onChange={(e) => setForm({ ...form, docs_path: e.target.value })}
      />
      <SelectField
        value={form.sync_mode}
        onChange={(e) => setForm({ ...form, sync_mode: e.target.value })}
      >
        {SYNC_MODES.map((m) => (
          <option key={m} value={m}>
            {m}
          </option>
        ))}
      </SelectField>
      <input
        className="input-field"
        type="number"
        min={60}
        placeholder={t("admin.repositories.syncIntervalPlaceholder")}
        value={form.sync_interval_seconds}
        onChange={(e) => setForm({ ...form, sync_interval_seconds: Number(e.target.value) || 300 })}
      />
      <input
        className="input-field sm:col-span-2"
        placeholder={tokenHint || t("admin.repositories.tokenPlaceholder")}
        value={form.access_token_ref}
        onChange={(e) => setForm({ ...form, access_token_ref: e.target.value })}
      />
      <input
        className="input-field sm:col-span-2"
        placeholder={t("admin.repositories.webhookPlaceholder")}
        value={form.webhook_secret_ref}
        onChange={(e) => setForm({ ...form, webhook_secret_ref: e.target.value })}
      />
      <label className="checkbox-label sm:col-span-2">
        <input
          type="checkbox"
          checked={form.enabled}
          onChange={(e) => setForm({ ...form, enabled: e.target.checked })}
          className="checkbox-field"
        />
        <span>{t("admin.repositories.repoEnabled")}</span>
      </label>
    </>
  );
}

export function AdminRepositoriesPage() {
  const { ready } = useAdminGuard();
  const { t } = useI18n();
  const qc = useQueryClient();
  const [error, setError] = useState("");
  const [syncingId, setSyncingId] = useState<string | null>(null);
  const [removingId, setRemovingId] = useState<string | null>(null);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [originalSpaceId, setOriginalSpaceId] = useState("");

  const [createForm, setCreateForm] = useState<RepoForm>(emptyForm());
  const [editForm, setEditForm] = useState<RepoForm>(emptyForm());

  const { data: spaces } = useQuery({
    queryKey: ["admin-spaces"],
    queryFn: () => api<{ items: Space[] }>("/api/admin/spaces"),
    enabled: ready,
  });

  const { data: repos, isLoading } = useQuery({
    queryKey: ["admin-repositories"],
    queryFn: () => api<{ items: Repository[] }>("/api/admin/repositories"),
    enabled: ready,
  });

  useEffect(() => {
    if (!editingId || !repos?.items) return;
    const repo = repos.items.find((r) => r.id === editingId);
    if (repo) {
      setEditForm(repoToForm(repo));
      setOriginalSpaceId(repo.space_id);
    }
  }, [editingId, repos]);

  const createRepo = useMutation({
    mutationFn: (body: Record<string, unknown>) =>
      api("/api/admin/repositories", { method: "POST", body: JSON.stringify(body) }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["admin-repositories"] });
      setCreateForm(emptyForm());
      setError("");
    },
    onError: (e) => setError(e instanceof ApiError ? e.message : t("admin.repositories.failed")),
  });

  const updateRepo = useMutation({
    mutationFn: async ({ id, form, prevSpaceId }: { id: string; form: RepoForm; prevSpaceId: string }) => {
      const body: Record<string, unknown> = {
        name: form.name,
        url: form.url,
        branch: form.branch,
        docs_path: form.docs_path,
        sync_mode: form.sync_mode,
        sync_interval_seconds: form.sync_interval_seconds,
        enabled: form.enabled,
      };
      if (form.access_token_ref) body.access_token_ref = form.access_token_ref;
      if (form.webhook_secret_ref) body.webhook_secret_ref = form.webhook_secret_ref;

      await api(`/api/admin/repositories/${id}`, { method: "PUT", body: JSON.stringify(body) });

      if (form.space_id && form.space_id !== prevSpaceId) {
        await api(`/api/admin/spaces/${form.space_id}/bind-repo`, {
          method: "POST",
          body: JSON.stringify({ repository_id: id }),
        });
      }
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["admin-repositories"] });
      setEditingId(null);
      setError("");
    },
    onError: (e) => setError(e instanceof ApiError ? e.message : t("admin.repositories.saveFailed")),
  });

  async function triggerSync(repoId: string) {
    setSyncingId(repoId);
    setError("");
    try {
      await api(`/api/admin/sync/${repoId}`, { method: "POST" });
      qc.invalidateQueries({ queryKey: ["admin-repositories"] });
    } catch (e) {
      setError(e instanceof ApiError ? e.message : t("admin.repositories.syncFailed"));
    } finally {
      setSyncingId(null);
    }
  }

  async function removeFromSpace(repo: Repository) {
    const ok = window.confirm(
      t("admin.repositories.removeFromSpaceConfirm", { name: repo.name, space: repo.space_name }),
    );
    if (!ok) return;
    setRemovingId(repo.id);
    setError("");
    try {
      await api(`/api/admin/spaces/${repo.space_id}/repositories/${repo.id}`, { method: "DELETE" });
      qc.invalidateQueries({ queryKey: ["admin-repositories"] });
      qc.invalidateQueries({ queryKey: ["admin-space-repos"] });
      if (editingId === repo.id) setEditingId(null);
    } catch (e) {
      setError(e instanceof ApiError ? e.message : t("admin.repositories.removeFailed"));
    } finally {
      setRemovingId(null);
    }
  }

  function startEdit(repo: Repository) {
    setEditingId(repo.id);
    setEditForm(repoToForm(repo));
    setOriginalSpaceId(repo.space_id);
    setError("");
  }

  if (!ready) return null;

  return (
    <FadeIn>
      {editingId && (
        <div className="glass mb-6 p-6 ring-1 ring-brand-500/20">
          <div className="mb-4 flex items-center justify-between">
            <h2 className="text-lg font-semibold text-fg">{t("admin.repositories.editTitle")}</h2>
            <button
              type="button"
              className="btn-ghost !px-2"
              onClick={() => setEditingId(null)}
              aria-label={t("admin.repositories.closeEditor")}
            >
              <X className="h-4 w-4" />
            </button>
          </div>
          <form
            className="grid gap-4 sm:grid-cols-2"
            onSubmit={(e) => {
              e.preventDefault();
              updateRepo.mutate({ id: editingId, form: editForm, prevSpaceId: originalSpaceId });
            }}
          >
            <RepositoryFormFields
              form={editForm}
              setForm={setEditForm}
              spaces={spaces?.items}
              tokenHint={t("admin.repositories.tokenEditHint")}
            />
            {error && editingId && <p className="text-sm text-danger-soft sm:col-span-2">{error}</p>}
            <div className="flex flex-wrap gap-3 sm:col-span-2">
              <button type="submit" className="btn-primary" disabled={updateRepo.isPending}>
                {updateRepo.isPending ? <Loader2 className="h-4 w-4 animate-spin" /> : t("admin.repositories.saveChanges")}
              </button>
              <button type="button" className="btn-secondary" onClick={() => setEditingId(null)}>
                {t("common.cancel")}
              </button>
              <button
                type="button"
                className="btn-secondary"
                disabled={syncingId === editingId}
                onClick={() => triggerSync(editingId)}
              >
                {syncingId === editingId ? (
                  <Loader2 className="h-4 w-4 animate-spin" />
                ) : (
                  <RefreshCw className="h-4 w-4" />
                )}
                {t("admin.repositories.syncNow")}
              </button>
              <button
                type="button"
                className="btn-secondary text-danger-soft"
                disabled={removingId === editingId}
                onClick={() => {
                  const repo = repos?.items.find((r) => r.id === editingId);
                  if (repo) removeFromSpace(repo);
                }}
              >
                {removingId === editingId ? (
                  <Loader2 className="h-4 w-4 animate-spin" />
                ) : (
                  <Trash2 className="h-4 w-4" />
                )}
                {t("admin.repositories.removeFromSpace")}
              </button>
            </div>
          </form>
        </div>
      )}

      <div className="glass p-6">
        <h2 className="text-lg font-semibold text-fg">{t("admin.repositories.title")}</h2>
        <p className="mt-1 text-sm text-muted">{t("admin.repositories.subtitle")}</p>

        {isLoading ? (
          <div className="flex justify-center py-12">
            <Loader2 className="h-6 w-6 animate-spin text-primary" />
          </div>
        ) : (
          <div className="mt-6 overflow-x-auto">
            <table className="w-full text-left text-sm">
              <thead>
                <tr className="border-b border-default text-subtle">
                  <th className="pb-3 pr-4 font-medium">{t("admin.repositories.colName")}</th>
                  <th className="pb-3 pr-4 font-medium">{t("admin.repositories.colSpace")}</th>
                  <th className="pb-3 pr-4 font-medium">{t("admin.repositories.colBranch")}</th>
                  <th className="pb-3 pr-4 font-medium">{t("admin.repositories.colSync")}</th>
                  <th className="pb-3 pr-4 font-medium">{t("admin.repositories.colStatus")}</th>
                  <th className="pb-3 font-medium">{t("admin.repositories.colActions")}</th>
                </tr>
              </thead>
              <tbody>
                {repos?.items.map((repo) => (
                  <tr
                    key={repo.id}
                    className={`border-b border-default ${editingId === repo.id ? "bg-highlight-row" : ""}`}
                  >
                    <td className="py-3 pr-4">
                      <p className="font-medium text-fg">{repo.name}</p>
                      <p className="text-xs text-subtle truncate max-w-xs">{repo.url}</p>
                      <p className="text-xs text-subtle">{repo.docs_path}</p>
                    </td>
                    <td className="py-3 pr-4 text-fg-secondary">{repo.space_name}</td>
                    <td className="py-3 pr-4 text-muted">{repo.branch}</td>
                    <td className="py-3 pr-4 text-muted">{repo.sync_mode}</td>
                    <td className="py-3 pr-4">
                      <span className={syncStatusClass(repo.last_sync_status)}>
                        {repo.last_sync_status || t("admin.repositories.syncStatusNever")}
                      </span>
                      {repo.last_sync_at && (
                        <p className="mt-1 text-xs text-subtle">{formatDate(repo.last_sync_at)}</p>
                      )}
                      {repo.last_sync_error && (
                        <p className="mt-1 max-w-xs text-xs text-danger-soft line-clamp-3" title={repo.last_sync_error}>
                          {repo.last_sync_error}
                        </p>
                      )}
                    </td>
                    <td className="py-3">
                      <div className="flex gap-1">
                        <button
                          type="button"
                          className="btn-ghost !px-2"
                          onClick={() => startEdit(repo)}
                          aria-label={t("admin.repositories.editTitle")}
                        >
                          <Pencil className="h-4 w-4" />
                        </button>
                        <button
                          type="button"
                          className="btn-ghost !px-2"
                          disabled={syncingId === repo.id}
                          onClick={() => triggerSync(repo.id)}
                          aria-label={t("admin.repositories.syncNow")}
                        >
                          {syncingId === repo.id ? (
                            <Loader2 className="h-4 w-4 animate-spin" />
                          ) : (
                            <RefreshCw className="h-4 w-4" />
                          )}
                        </button>
                        <button
                          type="button"
                          className="btn-ghost !px-2 text-danger-soft"
                          disabled={removingId === repo.id}
                          onClick={() => removeFromSpace(repo)}
                          aria-label={t("admin.repositories.removeFromSpace")}
                        >
                          {removingId === repo.id ? (
                            <Loader2 className="h-4 w-4 animate-spin" />
                          ) : (
                            <Trash2 className="h-4 w-4" />
                          )}
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
            {repos?.items.length === 0 && (
              <p className="py-6 text-sm text-subtle">{t("admin.repositories.noRepos")}</p>
            )}
          </div>
        )}
        {error && !editingId && <p className="mt-4 text-sm text-danger-soft">{error}</p>}
      </div>

      <div className="glass mt-6 p-6">
        <h2 className="text-lg font-semibold text-fg">{t("admin.repositories.createTitle")}</h2>
        <form
          className="mt-4 grid gap-4 sm:grid-cols-2"
          onSubmit={(e) => {
            e.preventDefault();
            createRepo.mutate({
              space_id: createForm.space_id,
              name: createForm.name,
              url: createForm.url,
              branch: createForm.branch,
              docs_path: createForm.docs_path,
              sync_mode: createForm.sync_mode,
              sync_interval_seconds: createForm.sync_interval_seconds,
              access_token_ref: createForm.access_token_ref || undefined,
              webhook_secret_ref: createForm.webhook_secret_ref || undefined,
              enabled: createForm.enabled,
            });
          }}
        >
          <RepositoryFormFields form={createForm} setForm={setCreateForm} spaces={spaces?.items} />
          {error && !editingId && <p className="text-sm text-danger-soft sm:col-span-2">{error}</p>}
          <button type="submit" className="btn-primary sm:col-span-2" disabled={createRepo.isPending}>
            {createRepo.isPending ? <Loader2 className="h-4 w-4 animate-spin" /> : t("admin.repositories.create")}
          </button>
        </form>
      </div>
    </FadeIn>
  );
}
