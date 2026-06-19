import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Loader2, Pencil, Plus, Trash2, UserPlus, X } from "lucide-react";
import { api, ApiError } from "@/lib/api";
import { FadeIn } from "@/components/motion-wrapper";
import { useI18n } from "@/lib/i18n";
import { useAdminGuard } from "./layout";

interface UserRow {
  id: string;
  email: string;
  display_name: string;
  is_active: boolean;
  roles: string[];
}

const ALL_ROLES = ["viewer", "editor", "admin", "super_admin"] as const;
type RoleName = (typeof ALL_ROLES)[number];

const ELEVATED_ROLES = new Set<RoleName>(["admin", "super_admin"]);

function hasElevatedRole(roles: string[]) {
  return roles.some((role) => ELEVATED_ROLES.has(role as RoleName));
}

function canManageUser(actorRoles: string[], targetRoles: string[]) {
  if (actorRoles.includes("super_admin")) return true;
  return actorRoles.includes("admin") && !hasElevatedRole(targetRoles);
}

function assignableRoles(isSuperAdmin: boolean): RoleName[] {
  return isSuperAdmin ? [...ALL_ROLES] : ["viewer", "editor"];
}

export function AdminUsersPage() {
  const { ready, user } = useAdminGuard();
  const { t } = useI18n();
  const qc = useQueryClient();
  const actorRoles = user?.roles ?? [];
  const isSuperAdmin = actorRoles.includes("super_admin");
  const canManageUsers = actorRoles.includes("super_admin") || actorRoles.includes("admin");

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [displayName, setDisplayName] = useState("");
  const [roles, setRoles] = useState<string[]>(["viewer"]);
  const [isActive, setIsActive] = useState(true);
  const [error, setError] = useState("");

  const [editingId, setEditingId] = useState<string | null>(null);
  const [editEmail, setEditEmail] = useState("");
  const [editPassword, setEditPassword] = useState("");
  const [editDisplayName, setEditDisplayName] = useState("");
  const [editRoles, setEditRoles] = useState<string[]>(["viewer"]);
  const [editActive, setEditActive] = useState(true);
  const [editError, setEditError] = useState("");
  const [deletingId, setDeletingId] = useState<string | null>(null);
  const [listError, setListError] = useState("");

  const { data, isLoading } = useQuery({
    queryKey: ["admin-users"],
    queryFn: () => api<{ items: UserRow[] }>("/api/admin/users"),
    enabled: ready && canManageUsers,
  });

  const createUser = useMutation({
    mutationFn: (body: {
      email: string;
      password: string;
      display_name?: string;
      roles: string[];
      is_active: boolean;
    }) =>
      api<UserRow>("/api/admin/users", {
        method: "POST",
        body: JSON.stringify(body),
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["admin-users"] });
      setEmail("");
      setPassword("");
      setDisplayName("");
      setRoles(["viewer"]);
      setIsActive(true);
      setError("");
    },
    onError: (e) => {
      if (e instanceof ApiError && e.message.includes("already exists")) {
        setError(t("admin.users.emailExists"));
      } else if (e instanceof ApiError && e.message.includes("invalid email")) {
        setError(t("admin.users.invalidEmail"));
      } else {
        setError(e instanceof ApiError ? e.message : t("admin.users.createFailed"));
      }
    },
  });

  const updateUser = useMutation({
    mutationFn: ({
      id,
      body,
    }: {
      id: string;
      body: {
        email: string;
        password?: string;
        display_name: string;
        roles: string[];
        is_active: boolean;
      };
    }) =>
      api<UserRow>(`/api/admin/users/${id}`, {
        method: "PUT",
        body: JSON.stringify({
          email: body.email,
          display_name: body.display_name,
          roles: body.roles,
          is_active: body.is_active,
          ...(body.password ? { password: body.password } : {}),
        }),
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["admin-users"] });
      setEditingId(null);
      setEditError("");
    },
    onError: (e) => {
      if (e instanceof ApiError && e.message.includes("forbidden")) {
        setEditError(t("admin.users.forbiddenUser"));
      } else if (e instanceof ApiError && e.message.includes("already exists")) {
        setEditError(t("admin.users.emailExists"));
      } else if (e instanceof ApiError && e.message.includes("invalid email")) {
        setEditError(t("admin.users.invalidEmail"));
      } else {
        setEditError(e instanceof ApiError ? e.message : t("admin.users.saveFailed"));
      }
    },
  });

  function toggleRole(current: string[], role: string, setter: (roles: string[]) => void) {
    setter(current.includes(role) ? current.filter((r) => r !== role) : [...current, role]);
  }

  function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    const selectedRoles = roles.length > 0 ? roles : ["viewer"];
    createUser.mutate({
      email: email.trim(),
      password,
      display_name: displayName.trim() || undefined,
      roles: selectedRoles,
      is_active: isActive,
    });
  }

  function startEdit(row: UserRow) {
    setEditingId(row.id);
    setEditEmail(row.email);
    setEditPassword("");
    setEditDisplayName(row.display_name);
    setEditRoles(row.roles.length > 0 ? [...row.roles] : ["viewer"]);
    setEditActive(row.is_active);
    setEditError("");
  }

  function handleUpdate(e: React.FormEvent) {
    e.preventDefault();
    if (!editingId) return;
    setEditError("");
    const selectedRoles = editRoles.length > 0 ? editRoles : ["viewer"];
    updateUser.mutate({
      id: editingId,
      body: {
        email: editEmail.trim(),
        password: editPassword.trim() || undefined,
        display_name: editDisplayName.trim(),
        roles: selectedRoles,
        is_active: editActive,
      },
    });
  }

  async function removeUser(row: UserRow) {
    const ok = window.confirm(t("admin.users.deleteConfirm", { email: row.email }));
    if (!ok) return;
    setDeletingId(row.id);
    setListError("");
    try {
      await api(`/api/admin/users/${row.id}`, { method: "DELETE" });
      qc.invalidateQueries({ queryKey: ["admin-users"] });
      if (editingId === row.id) setEditingId(null);
    } catch (e) {
      if (e instanceof ApiError && e.message.includes("your own account")) {
        setListError(t("admin.users.cannotDeleteSelf"));
      } else if (e instanceof ApiError && e.message.includes("last super admin")) {
        setListError(t("admin.users.cannotDeleteLastSuperAdmin"));
      } else if (e instanceof ApiError && e.message.includes("forbidden")) {
        setListError(t("admin.users.forbiddenUser"));
      } else {
        setListError(e instanceof ApiError ? e.message : t("admin.users.deleteFailed"));
      }
    } finally {
      setDeletingId(null);
    }
  }

  if (!ready) return null;

  if (!canManageUsers) return null;

  const roleOptions = assignableRoles(isSuperAdmin);

  return (
    <FadeIn>
      <div className="glass p-6">
        <h2 className="text-lg font-semibold text-fg">{t("admin.nav.users")}</h2>
        <p className="mt-1 text-sm text-muted">{t("admin.users.subtitle")}</p>

        {isSuperAdmin ? (
          <form onSubmit={handleCreate} className="mt-6 space-y-4 rounded-xl border border-default p-4">
            <h3 className="flex items-center gap-2 text-sm font-semibold text-fg">
              <UserPlus className="h-4 w-4 text-primary" />
              {t("admin.users.createTitle")}
            </h3>

            <div className="grid gap-4 sm:grid-cols-2">
              <label className="block sm:col-span-1">
                <span className="mb-1 block text-sm text-muted">{t("admin.users.email")}</span>
                <input
                  className="input-field"
                  type="email"
                  autoComplete="off"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  required
                />
              </label>
              <label className="block sm:col-span-1">
                <span className="mb-1 block text-sm text-muted">{t("admin.users.password")}</span>
                <input
                  className="input-field"
                  type="password"
                  autoComplete="new-password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  minLength={8}
                  required
                />
              </label>
              <label className="block sm:col-span-2">
                <span className="mb-1 block text-sm text-muted">{t("admin.users.displayNameOptional")}</span>
                <input
                  className="input-field"
                  type="text"
                  value={displayName}
                  onChange={(e) => setDisplayName(e.target.value)}
                />
              </label>
            </div>

            <fieldset>
              <legend className="mb-2 text-sm text-muted">{t("admin.users.roles")}</legend>
              <div className="flex flex-wrap gap-3">
                {ALL_ROLES.map((role) => (
                  <label key={role} className="checkbox-label">
                    <input
                      type="checkbox"
                      className="checkbox-field"
                      checked={roles.includes(role)}
                      onChange={() => toggleRole(roles, role, setRoles)}
                    />
                    <span>{t(`admin.users.roleLabels.${role}`)}</span>
                  </label>
                ))}
              </div>
            </fieldset>

            <label className="checkbox-label">
              <input
                type="checkbox"
                className="checkbox-field"
                checked={isActive}
                onChange={(e) => setIsActive(e.target.checked)}
              />
              <span>{t("admin.users.active")}</span>
            </label>

            {error && <p className="text-sm text-danger-soft">{error}</p>}

            <button type="submit" className="btn-primary" disabled={createUser.isPending}>
              {createUser.isPending ? (
                <Loader2 className="h-4 w-4 animate-spin" />
              ) : (
                <Plus className="h-4 w-4" />
              )}
              {createUser.isPending ? t("admin.users.creating") : t("admin.users.create")}
            </button>
          </form>
        ) : (
          <p className="mt-4 text-sm text-subtle">{t("admin.users.createSuperAdminOnly")}</p>
        )}

        {editingId && (
          <form
            onSubmit={handleUpdate}
            className="mt-6 space-y-4 rounded-xl border border-brand-500/30 bg-surface-muted p-4 ring-1 ring-brand-500/20"
          >
            <div className="flex items-center justify-between">
              <h3 className="flex items-center gap-2 text-sm font-semibold text-fg">
                <Pencil className="h-4 w-4 text-primary" />
                {t("admin.users.editTitle")}
              </h3>
              <button
                type="button"
                className="btn-ghost !px-2"
                onClick={() => setEditingId(null)}
                aria-label={t("common.cancel")}
              >
                <X className="h-4 w-4" />
              </button>
            </div>

            <div className="grid gap-4 sm:grid-cols-2">
              <label className="block sm:col-span-1">
                <span className="mb-1 block text-sm text-muted">{t("admin.users.email")}</span>
                <input
                  className="input-field"
                  type="email"
                  autoComplete="off"
                  value={editEmail}
                  onChange={(e) => setEditEmail(e.target.value)}
                  required
                />
              </label>
              <label className="block sm:col-span-1">
                <span className="mb-1 block text-sm text-muted">{t("admin.users.passwordOptional")}</span>
                <input
                  className="input-field"
                  type="password"
                  autoComplete="new-password"
                  value={editPassword}
                  onChange={(e) => setEditPassword(e.target.value)}
                  minLength={editPassword ? 8 : undefined}
                  placeholder={t("admin.users.passwordKeep")}
                />
              </label>
              <label className="block sm:col-span-2">
                <span className="mb-1 block text-sm text-muted">{t("admin.users.displayName")}</span>
                <input
                  className="input-field"
                  type="text"
                  value={editDisplayName}
                  onChange={(e) => setEditDisplayName(e.target.value)}
                />
              </label>
            </div>

            <fieldset>
              <legend className="mb-2 text-sm text-muted">{t("admin.users.roles")}</legend>
              <div className="flex flex-wrap gap-3">
                {roleOptions.map((role) => (
                  <label key={role} className="checkbox-label">
                    <input
                      type="checkbox"
                      className="checkbox-field"
                      checked={editRoles.includes(role)}
                      onChange={() => toggleRole(editRoles, role, setEditRoles)}
                    />
                    <span>{t(`admin.users.roleLabels.${role}`)}</span>
                  </label>
                ))}
              </div>
            </fieldset>

            <label className="checkbox-label">
              <input
                type="checkbox"
                className="checkbox-field"
                checked={editActive}
                onChange={(e) => setEditActive(e.target.checked)}
              />
              <span>{t("admin.users.active")}</span>
            </label>

            {editError && <p className="text-sm text-danger-soft">{editError}</p>}

            <button type="submit" className="btn-primary" disabled={updateUser.isPending}>
              {updateUser.isPending ? (
                <Loader2 className="h-4 w-4 animate-spin" />
              ) : (
                <Pencil className="h-4 w-4" />
              )}
              {updateUser.isPending ? t("admin.users.saving") : t("admin.users.save")}
            </button>
          </form>
        )}

        {isLoading ? (
          <div className="flex justify-center py-12">
            <Loader2 className="h-6 w-6 animate-spin text-primary" />
          </div>
        ) : (
          <div className="mt-6 space-y-2">
            {listError && <p className="text-sm text-danger-soft">{listError}</p>}
            {data?.items.map((row) => {
              const editable = canManageUser(actorRoles, row.roles);
              return (
                <div
                  key={row.id}
                  className="flex flex-wrap items-center justify-between gap-2 rounded-xl border border-default bg-surface-muted px-4 py-3"
                >
                  <div>
                    <p className="font-medium text-fg">{row.display_name || row.email}</p>
                    <p className="text-xs text-subtle">{row.email}</p>
                  </div>
                  <div className="flex flex-wrap items-center gap-2">
                    {row.roles.map((role) => (
                      <span key={role} className="badge badge-primary">
                        {t(`admin.users.roleLabels.${role as RoleName}`) || role}
                      </span>
                    ))}
                    {!row.is_active && (
                      <span className="badge badge-danger">{t("admin.users.inactive")}</span>
                    )}
                    {editable ? (
                      <>
                        <button
                          type="button"
                          className="btn-ghost !px-2"
                          onClick={() => startEdit(row)}
                          aria-label={t("admin.users.edit")}
                        >
                          <Pencil className="h-4 w-4" />
                        </button>
                        <button
                          type="button"
                          className="btn-ghost !px-2 text-danger-soft hover:text-danger"
                          onClick={() => removeUser(row)}
                          disabled={deletingId === row.id}
                          aria-label={t("admin.users.delete")}
                        >
                          {deletingId === row.id ? (
                            <Loader2 className="h-4 w-4 animate-spin" />
                          ) : (
                            <Trash2 className="h-4 w-4" />
                          )}
                        </button>
                      </>
                    ) : (
                      <span className="text-xs text-subtle">{t("admin.users.protectedUser")}</span>
                    )}
                  </div>
                </div>
              );
            })}
            {data?.items.length === 0 && (
              <p className="text-sm text-subtle">{t("admin.users.noUsers")}</p>
            )}
          </div>
        )}
      </div>
    </FadeIn>
  );
}
