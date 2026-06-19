import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Loader2, Pencil, Plus, Trash2, UserPlus, Users, X } from "lucide-react";
import { api, ApiError } from "@/lib/api";
import { FadeIn } from "@/components/motion-wrapper";
import { SelectField } from "@/components/select-field";
import { useI18n } from "@/lib/i18n";
import { useAdminGuard } from "./layout";

interface GroupRow {
  id: string;
  name: string;
  description?: string;
  external_id?: string;
  member_count: number;
  space_count: number;
}

interface GroupMember {
  user_id: string;
  email: string;
  display_name: string;
}

interface UserOption {
  id: string;
  email: string;
  display_name: string;
}

export function AdminGroupsPage() {
  const { ready } = useAdminGuard();
  const { t } = useI18n();
  const qc = useQueryClient();

  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [error, setError] = useState("");

  const [editingId, setEditingId] = useState<string | null>(null);
  const [editName, setEditName] = useState("");
  const [editDescription, setEditDescription] = useState("");
  const [editExternalId, setEditExternalId] = useState("");
  const [editError, setEditError] = useState("");
  const [selectedUserId, setSelectedUserId] = useState("");
  const [removingMemberId, setRemovingMemberId] = useState<string | null>(null);
  const [deletingId, setDeletingId] = useState<string | null>(null);

  const { data, isLoading } = useQuery({
    queryKey: ["admin-groups"],
    queryFn: () => api<{ items: GroupRow[] }>("/api/admin/groups"),
    enabled: ready,
  });

  const { data: members, isLoading: membersLoading } = useQuery({
    queryKey: ["admin-group-members", editingId],
    queryFn: () => api<{ items: GroupMember[] }>(`/api/admin/groups/${editingId}/members`),
    enabled: ready && editingId !== null,
  });

  const { data: usersData } = useQuery({
    queryKey: ["admin-users"],
    queryFn: () => api<{ items: UserOption[] }>("/api/admin/users"),
    enabled: ready && editingId !== null,
  });

  const createGroup = useMutation({
    mutationFn: (body: { name: string; description?: string }) =>
      api("/api/admin/groups", { method: "POST", body: JSON.stringify(body) }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["admin-groups"] });
      setName("");
      setDescription("");
      setError("");
    },
    onError: (e) => {
      if (e instanceof ApiError && e.message.includes("already exists")) {
        setError(t("admin.groups.nameExists"));
      } else {
        setError(e instanceof ApiError ? e.message : t("admin.groups.createFailed"));
      }
    },
  });

  const updateGroup = useMutation({
    mutationFn: ({ id, body }: { id: string; body: Record<string, string> }) =>
      api(`/api/admin/groups/${id}`, { method: "PUT", body: JSON.stringify(body) }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["admin-groups"] });
      setEditError("");
    },
    onError: (e) => {
      setEditError(e instanceof ApiError ? e.message : t("admin.groups.saveFailed"));
    },
  });

  const addMember = useMutation({
    mutationFn: ({ groupId, userId }: { groupId: string; userId: string }) =>
      api(`/api/admin/groups/${groupId}/members`, {
        method: "POST",
        body: JSON.stringify({ user_id: userId }),
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["admin-group-members", editingId] });
      qc.invalidateQueries({ queryKey: ["admin-groups"] });
      setSelectedUserId("");
      setEditError("");
    },
    onError: (e) => {
      if (e instanceof ApiError && e.message.includes("already in group")) {
        setEditError(t("admin.groups.memberExists"));
      } else {
        setEditError(e instanceof ApiError ? e.message : t("common.failed"));
      }
    },
  });

  async function removeMember(userId: string) {
    if (!editingId) return;
    setRemovingMemberId(userId);
    setEditError("");
    try {
      await api(`/api/admin/groups/${editingId}/members/${userId}`, { method: "DELETE" });
      qc.invalidateQueries({ queryKey: ["admin-group-members", editingId] });
      qc.invalidateQueries({ queryKey: ["admin-groups"] });
    } catch (e) {
      setEditError(e instanceof ApiError ? e.message : t("common.failed"));
    } finally {
      setRemovingMemberId(null);
    }
  }

  async function deleteGroup(row: GroupRow) {
    const ok = window.confirm(t("admin.groups.deleteConfirm", { name: row.name }));
    if (!ok) return;
    setDeletingId(row.id);
    setError("");
    try {
      await api(`/api/admin/groups/${row.id}`, { method: "DELETE" });
      qc.invalidateQueries({ queryKey: ["admin-groups"] });
      if (editingId === row.id) setEditingId(null);
    } catch (e) {
      setError(e instanceof ApiError ? e.message : t("admin.groups.deleteFailed"));
    } finally {
      setDeletingId(null);
    }
  }

  function startEdit(row: GroupRow) {
    setEditingId(row.id);
    setEditName(row.name);
    setEditDescription(row.description || "");
    setEditExternalId(row.external_id || "");
    setEditError("");
    setSelectedUserId("");
  }

  function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    createGroup.mutate({
      name: name.trim(),
      description: description.trim() || undefined,
    });
  }

  function handleUpdate(e: React.FormEvent) {
    e.preventDefault();
    if (!editingId) return;
    updateGroup.mutate({
      id: editingId,
      body: {
        name: editName.trim(),
        description: editDescription.trim(),
        external_id: editExternalId.trim(),
      },
    });
  }

  const memberIds = new Set(members?.items.map((m) => m.user_id) ?? []);
  const availableUsers =
    usersData?.items.filter((u) => !memberIds.has(u.id)) ?? [];

  if (!ready) return null;

  return (
    <FadeIn>
      <div className="glass p-6">
        <h2 className="text-lg font-semibold text-fg">{t("admin.nav.groups")}</h2>
        <p className="mt-1 text-sm text-muted">{t("admin.groups.subtitle")}</p>

        <form onSubmit={handleCreate} className="mt-6 space-y-4 rounded-xl border border-default p-4">
          <h3 className="flex items-center gap-2 text-sm font-semibold text-fg">
            <Users className="h-4 w-4 text-primary" />
            {t("admin.groups.createTitle")}
          </h3>
          <input
            className="input-field"
            placeholder={t("admin.groups.name")}
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
          />
          <input
            className="input-field"
            placeholder={t("admin.groups.descriptionOptional")}
            value={description}
            onChange={(e) => setDescription(e.target.value)}
          />
          {error && <p className="text-sm text-danger-soft">{error}</p>}
          <button type="submit" className="btn-primary" disabled={createGroup.isPending}>
            {createGroup.isPending ? (
              <Loader2 className="h-4 w-4 animate-spin" />
            ) : (
              <Plus className="h-4 w-4" />
            )}
            {createGroup.isPending ? t("admin.groups.creating") : t("admin.groups.create")}
          </button>
        </form>

        {editingId && (
          <div className="mt-6 space-y-4 rounded-xl border border-brand-500/30 bg-surface-muted p-4 ring-1 ring-brand-500/20">
            <div className="flex items-center justify-between">
              <h3 className="flex items-center gap-2 text-sm font-semibold text-fg">
                <Pencil className="h-4 w-4 text-primary" />
                {t("admin.groups.editTitle")}
              </h3>
              <button type="button" className="btn-ghost !px-2" onClick={() => setEditingId(null)}>
                <X className="h-4 w-4" />
              </button>
            </div>

            <form onSubmit={handleUpdate} className="space-y-3">
              <input
                className="input-field"
                value={editName}
                onChange={(e) => setEditName(e.target.value)}
                required
              />
              <input
                className="input-field"
                placeholder={t("admin.groups.descriptionOptional")}
                value={editDescription}
                onChange={(e) => setEditDescription(e.target.value)}
              />
              <input
                className="input-field"
                placeholder={t("admin.groups.externalIdOptional")}
                value={editExternalId}
                onChange={(e) => setEditExternalId(e.target.value)}
              />
              <button type="submit" className="btn-secondary" disabled={updateGroup.isPending}>
                {updateGroup.isPending ? (
                  <Loader2 className="h-4 w-4 animate-spin" />
                ) : (
                  t("admin.groups.save")
                )}
              </button>
            </form>

            <div className="border-t border-default pt-4">
              <h4 className="text-sm font-semibold text-fg">{t("admin.groups.members")}</h4>
              <div className="mt-3 flex flex-wrap gap-2">
                <SelectField
                  className="min-w-[12rem] flex-1"
                  value={selectedUserId}
                  onChange={(e) => setSelectedUserId(e.target.value)}
                >
                  <option value="">{t("admin.groups.selectUser")}</option>
                  {availableUsers.map((u) => (
                    <option key={u.id} value={u.id}>
                      {u.display_name || u.email} ({u.email})
                    </option>
                  ))}
                </SelectField>
                <button
                  type="button"
                  className="btn-primary"
                  disabled={!selectedUserId || addMember.isPending}
                  onClick={() =>
                    editingId &&
                    selectedUserId &&
                    addMember.mutate({ groupId: editingId, userId: selectedUserId })
                  }
                >
                  {addMember.isPending ? (
                    <Loader2 className="h-4 w-4 animate-spin" />
                  ) : (
                    <UserPlus className="h-4 w-4" />
                  )}
                  {t("admin.groups.addMember")}
                </button>
              </div>

              {membersLoading ? (
                <div className="flex justify-center py-6">
                  <Loader2 className="h-5 w-5 animate-spin text-primary" />
                </div>
              ) : (
                <div className="mt-3 space-y-2">
                  {members?.items.map((m) => (
                    <div
                      key={m.user_id}
                      className="flex items-center justify-between rounded-lg border border-default bg-surface px-3 py-2"
                    >
                      <div>
                        <p className="text-sm font-medium text-fg">{m.display_name || m.email}</p>
                        <p className="text-xs text-subtle">{m.email}</p>
                      </div>
                      <button
                        type="button"
                        className="btn-ghost !px-2 text-danger-soft"
                        disabled={removingMemberId === m.user_id}
                        onClick={() => removeMember(m.user_id)}
                        aria-label={t("admin.groups.removeMember")}
                      >
                        {removingMemberId === m.user_id ? (
                          <Loader2 className="h-4 w-4 animate-spin" />
                        ) : (
                          <Trash2 className="h-4 w-4" />
                        )}
                      </button>
                    </div>
                  ))}
                  {members?.items.length === 0 && (
                    <p className="text-sm text-subtle">{t("admin.groups.noMembers")}</p>
                  )}
                </div>
              )}
            </div>

            {editError && <p className="text-sm text-danger-soft">{editError}</p>}
          </div>
        )}

        {isLoading ? (
          <div className="flex justify-center py-12">
            <Loader2 className="h-6 w-6 animate-spin text-primary" />
          </div>
        ) : (
          <div className="mt-6 space-y-2">
            {data?.items.map((row) => (
              <div
                key={row.id}
                className="flex flex-wrap items-center justify-between gap-2 rounded-xl border border-default bg-surface-muted px-4 py-3"
              >
                <div>
                  <p className="font-medium text-fg">{row.name}</p>
                  {row.description && <p className="text-sm text-muted">{row.description}</p>}
                  <p className="mt-1 text-xs text-subtle">
                    {t("admin.groups.membersCount", { count: row.member_count })} ·{" "}
                    {t("admin.groups.spacesCount", { count: row.space_count })}
                  </p>
                </div>
                <div className="flex items-center gap-2">
                  <button
                    type="button"
                    className="btn-ghost !px-2"
                    onClick={() => startEdit(row)}
                    aria-label={t("admin.groups.editTitle")}
                  >
                    <Pencil className="h-4 w-4" />
                  </button>
                  <button
                    type="button"
                    className="btn-ghost !px-2 text-danger-soft"
                    disabled={deletingId === row.id}
                    onClick={() => deleteGroup(row)}
                    aria-label={t("admin.groups.deleteGroup")}
                  >
                    {deletingId === row.id ? (
                      <Loader2 className="h-4 w-4 animate-spin" />
                    ) : (
                      <Trash2 className="h-4 w-4" />
                    )}
                  </button>
                </div>
              </div>
            ))}
            {data?.items.length === 0 && (
              <p className="text-sm text-subtle">{t("admin.groups.noGroups")}</p>
            )}
          </div>
        )}
      </div>
    </FadeIn>
  );
}
