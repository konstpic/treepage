import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Globe, Loader2, Pencil, Trash2, UsersRound, X } from "lucide-react";
import { api, ApiError } from "@/lib/api";
import { FadeIn } from "@/components/motion-wrapper";
import { SelectField } from "@/components/select-field";
import { useI18n } from "@/lib/i18n";
import { useAdminGuard } from "./layout";
import { SpacePageACLPanel } from "@/components/space-page-acl";

interface Space {
  id: string;
  slug: string;
  name: string;
  description?: string;
  is_public: boolean;
}

interface SpaceRepository {
  id: string;
  name: string;
  url: string;
  branch: string;
  docs_path: string;
}

type SpaceEditForm = {
  name: string;
  description: string;
  is_public: boolean;
};

interface SpaceGroupRow {
  group_id: string;
  group_name: string;
  role: string;
  description?: string;
}

interface SpaceMemberRow {
  user_id: string;
  email: string;
  display_name: string;
  role: string;
}

interface GroupOption {
  id: string;
  name: string;
}

const SPACE_ROLES = ["viewer", "editor", "admin"] as const;

export function AdminSpacesPage() {
  const { ready } = useAdminGuard();
  const { t } = useI18n();
  const qc = useQueryClient();
  const [slug, setSlug] = useState("");
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [isPublic, setIsPublic] = useState(false);
  const [error, setError] = useState("");
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editForm, setEditForm] = useState<SpaceEditForm>({ name: "", description: "", is_public: false });
  const [removingRepoId, setRemovingRepoId] = useState<string | null>(null);
  const [selectedGroupId, setSelectedGroupId] = useState("");
  const [groupRole, setGroupRole] = useState<string>("viewer");
  const [removingGroupId, setRemovingGroupId] = useState<string | null>(null);
  const [selectedUserId, setSelectedUserId] = useState("");
  const [memberRole, setMemberRole] = useState<string>("viewer");
  const [removingMemberId, setRemovingMemberId] = useState<string | null>(null);

  const { data, isLoading } = useQuery({
    queryKey: ["admin-spaces"],
    queryFn: () => api<{ items: Space[] }>("/api/admin/spaces"),
    enabled: ready,
  });

  const { data: spaceRepos, isLoading: reposLoading } = useQuery({
    queryKey: ["admin-space-repos", editingId],
    queryFn: () => api<{ items: SpaceRepository[] }>(`/api/admin/spaces/${editingId}/repositories`),
    enabled: ready && editingId !== null,
  });

  const { data: spaceGroups, isLoading: groupsLoading } = useQuery({
    queryKey: ["admin-space-groups", editingId],
    queryFn: () => api<{ items: SpaceGroupRow[] }>(`/api/admin/spaces/${editingId}/groups`),
    enabled: ready && editingId !== null,
  });

  const { data: spaceMembers, isLoading: membersLoading } = useQuery({
    queryKey: ["admin-space-members", editingId],
    queryFn: () => api<{ items: SpaceMemberRow[] }>(`/api/admin/spaces/${editingId}/members`),
    enabled: ready && editingId !== null,
  });

  const { data: allGroups } = useQuery({
    queryKey: ["admin-groups"],
    queryFn: () => api<{ items: GroupOption[] }>("/api/admin/groups"),
    enabled: ready && editingId !== null,
  });

  const { data: allUsers } = useQuery({
    queryKey: ["admin-users-picker"],
    queryFn: () =>
      api<{ items: { id: string; email: string; display_name: string }[] }>("/api/admin/users"),
    enabled: ready && editingId !== null,
  });

  const createSpace = useMutation({
    mutationFn: (body: { slug: string; name: string; description?: string; is_public: boolean }) =>
      api("/api/spaces", { method: "POST", body: JSON.stringify(body) }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["admin-spaces"] });
      qc.invalidateQueries({ queryKey: ["spaces"] });
      setSlug("");
      setName("");
      setDescription("");
      setIsPublic(false);
      setError("");
    },
    onError: (e) => setError(e instanceof ApiError ? e.message : t("admin.spaces.failed")),
  });

  const updateSpace = useMutation({
    mutationFn: ({ id, form }: { id: string; form: SpaceEditForm }) =>
      api<Space>(`/api/admin/spaces/${id}`, {
        method: "PATCH",
        body: JSON.stringify({
          name: form.name,
          description: form.description || "",
          is_public: form.is_public,
        }),
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["admin-spaces"] });
      qc.invalidateQueries({ queryKey: ["spaces"] });
      setError("");
    },
    onError: (e) => setError(e instanceof ApiError ? e.message : t("admin.spaces.updateFailed")),
  });

  const unbindRepo = useMutation({
    mutationFn: ({ spaceId, repoId }: { spaceId: string; repoId: string }) =>
      api(`/api/admin/spaces/${spaceId}/repositories/${repoId}`, { method: "DELETE" }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["admin-space-repos", editingId] });
      qc.invalidateQueries({ queryKey: ["admin-repositories"] });
      setRemovingRepoId(null);
      setError("");
    },
    onError: (e) => {
      setRemovingRepoId(null);
      setError(e instanceof ApiError ? e.message : t("admin.spaces.removeRepoFailed"));
    },
  });

  const assignGroup = useMutation({
    mutationFn: ({ spaceId, groupId, role }: { spaceId: string; groupId: string; role: string }) =>
      api(`/api/admin/spaces/${spaceId}/groups`, {
        method: "POST",
        body: JSON.stringify({ group_id: groupId, role }),
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["admin-space-groups", editingId] });
      qc.invalidateQueries({ queryKey: ["admin-groups"] });
      qc.invalidateQueries({ queryKey: ["spaces"] });
      setSelectedGroupId("");
      setGroupRole("viewer");
      setError("");
    },
    onError: (e) => {
      setError(e instanceof ApiError ? e.message : t("admin.spaceGroups.assignFailed"));
    },
  });

  const assignMember = useMutation({
    mutationFn: ({ spaceId, userId, role }: { spaceId: string; userId: string; role: string }) =>
      api(`/api/admin/spaces/${spaceId}/members`, {
        method: "POST",
        body: JSON.stringify({ user_id: userId, role }),
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["admin-space-members", editingId] });
      qc.invalidateQueries({ queryKey: ["spaces"] });
      setSelectedUserId("");
      setMemberRole("viewer");
      setError("");
    },
    onError: (e) => {
      setError(e instanceof ApiError ? e.message : t("admin.spaceMembers.assignFailed"));
    },
  });

  function startEdit(space: Space) {
    setEditingId(space.id);
    setEditForm({
      name: space.name,
      description: space.description || "",
      is_public: space.is_public,
    });
    setError("");
  }

  function closeEdit() {
    setEditingId(null);
    setRemovingRepoId(null);
    setRemovingGroupId(null);
    setRemovingMemberId(null);
    setSelectedGroupId("");
    setGroupRole("viewer");
    setSelectedUserId("");
    setMemberRole("viewer");
    setError("");
  }

  function confirmRemoveRepo(repo: SpaceRepository) {
    if (!editingId) return;
    const ok = window.confirm(
      t("admin.spaces.removeRepoConfirm", { name: repo.name }),
    );
    if (!ok) return;
    setRemovingRepoId(repo.id);
    unbindRepo.mutate({ spaceId: editingId, repoId: repo.id });
  }

  async function removeSpaceMember(userId: string) {
    if (!editingId) return;
    setRemovingMemberId(userId);
    setError("");
    try {
      await api(`/api/admin/spaces/${editingId}/members/${userId}`, { method: "DELETE" });
      qc.invalidateQueries({ queryKey: ["admin-space-members", editingId] });
      qc.invalidateQueries({ queryKey: ["spaces"] });
    } catch (e) {
      setError(e instanceof ApiError ? e.message : t("common.failed"));
    } finally {
      setRemovingMemberId(null);
    }
  }

  async function removeSpaceGroup(groupId: string) {
    if (!editingId) return;
    setRemovingGroupId(groupId);
    setError("");
    try {
      await api(`/api/admin/spaces/${editingId}/groups/${groupId}`, { method: "DELETE" });
      qc.invalidateQueries({ queryKey: ["admin-space-groups", editingId] });
      qc.invalidateQueries({ queryKey: ["admin-groups"] });
      qc.invalidateQueries({ queryKey: ["spaces"] });
    } catch (e) {
      setError(e instanceof ApiError ? e.message : t("common.failed"));
    } finally {
      setRemovingGroupId(null);
    }
  }

  const editingSpace = data?.items.find((s) => s.id === editingId);
  const assignedGroupIds = new Set(spaceGroups?.items.map((g) => g.group_id) ?? []);
  const availableGroups = allGroups?.items.filter((g) => !assignedGroupIds.has(g.id)) ?? [];
  const assignedUserIds = new Set(spaceMembers?.items.map((m) => m.user_id) ?? []);
  const availableUsers = allUsers?.items.filter((u) => !assignedUserIds.has(u.id)) ?? [];

  if (!ready) return null;

  return (
    <FadeIn>
      {editingId && editingSpace && (
        <div className="glass mb-6 p-6 ring-1 ring-brand-500/20">
          <div className="mb-4 flex items-center justify-between">
            <div>
              <h2 className="text-lg font-semibold text-fg">{t("admin.spaces.editTitle")}</h2>
              <p className="text-xs text-subtle">/{editingSpace.slug}</p>
            </div>
            <button type="button" className="btn-ghost !px-2" onClick={closeEdit} aria-label={t("admin.spaces.closeEditor")}>
              <X className="h-4 w-4" />
            </button>
          </div>
          <form
            className="space-y-4"
            onSubmit={(e) => {
              e.preventDefault();
              updateSpace.mutate({ id: editingId, form: editForm });
            }}
          >
            <input
              className="input-field"
              placeholder={t("admin.spaces.namePlaceholder")}
              value={editForm.name}
              onChange={(e) => setEditForm({ ...editForm, name: e.target.value })}
              required
            />
            <input
              className="input-field"
              placeholder={t("admin.spaces.descriptionPlaceholder")}
              value={editForm.description}
              onChange={(e) => setEditForm({ ...editForm, description: e.target.value })}
            />
            <label className="flex items-center gap-3 text-sm text-fg-secondary">
              <Globe className="h-4 w-4 text-primary" />
              <input
                type="checkbox"
                checked={editForm.is_public}
                onChange={(e) => setEditForm({ ...editForm, is_public: e.target.checked })}
                className="checkbox-field"
              />
              {t("admin.spaces.publishPublic")}
            </label>
            {error && editingId && <p className="text-sm text-danger-soft">{error}</p>}
            <div className="flex flex-wrap gap-3">
              <button type="submit" className="btn-primary" disabled={updateSpace.isPending}>
                {updateSpace.isPending ? <Loader2 className="h-4 w-4 animate-spin" /> : t("admin.spaces.saveChanges")}
              </button>
              <button type="button" className="btn-secondary" onClick={closeEdit}>
                {t("common.cancel")}
              </button>
            </div>
          </form>

          <div className="mt-8 border-t border-default pt-6">
            <h3 className="flex items-center gap-2 text-sm font-semibold text-fg">
              <UsersRound className="h-4 w-4 text-primary" />
              {t("admin.spaceAccess.title")}
            </h3>
            <p className="mt-1 text-xs text-muted">{t("admin.spaceAccess.hint")}</p>

            <h4 className="mt-5 text-xs font-semibold uppercase tracking-wide text-subtle">
              {t("admin.spaceMembers.title")}
            </h4>
            <div className="mt-3 flex flex-wrap gap-2">
              <SelectField
                className="min-w-[10rem] flex-1"
                value={selectedUserId}
                onChange={(e) => setSelectedUserId(e.target.value)}
              >
                <option value="">{t("admin.spaceMembers.selectUser")}</option>
                {availableUsers.map((u) => (
                  <option key={u.id} value={u.id}>
                    {u.display_name || u.email}
                  </option>
                ))}
              </SelectField>
              <SelectField className="w-36" value={memberRole} onChange={(e) => setMemberRole(e.target.value)}>
                {SPACE_ROLES.map((role) => (
                  <option key={role} value={role}>
                    {t(`admin.groups.roleLabels.${role}`)}
                  </option>
                ))}
              </SelectField>
              <button
                type="button"
                className="btn-primary"
                disabled={!selectedUserId || assignMember.isPending}
                onClick={() =>
                  editingId &&
                  selectedUserId &&
                  assignMember.mutate({ spaceId: editingId, userId: selectedUserId, role: memberRole })
                }
              >
                {assignMember.isPending ? (
                  <Loader2 className="h-4 w-4 animate-spin" />
                ) : (
                  t("admin.spaceMembers.assignButton")
                )}
              </button>
            </div>
            {membersLoading ? (
              <div className="flex justify-center py-6">
                <Loader2 className="h-5 w-5 animate-spin text-primary" />
              </div>
            ) : (
              <div className="mt-3 space-y-2">
                {spaceMembers?.items.map((member) => (
                  <div
                    key={member.user_id}
                    className="flex flex-wrap items-center justify-between gap-3 rounded-xl border border-default bg-surface-muted px-4 py-3"
                  >
                    <div>
                      <p className="font-medium text-fg">{member.display_name || member.email}</p>
                      <p className="text-xs text-subtle">{member.email}</p>
                    </div>
                    <div className="flex items-center gap-2">
                      <span className="badge badge-primary">
                        {t(`admin.groups.roleLabels.${member.role as (typeof SPACE_ROLES)[number]}`) || member.role}
                      </span>
                      <button
                        type="button"
                        className="btn-ghost text-danger-soft !px-2"
                        disabled={removingMemberId === member.user_id}
                        onClick={() => removeSpaceMember(member.user_id)}
                        aria-label={t("admin.spaceMembers.remove")}
                      >
                        {removingMemberId === member.user_id ? (
                          <Loader2 className="h-4 w-4 animate-spin" />
                        ) : (
                          <Trash2 className="h-4 w-4" />
                        )}
                      </button>
                    </div>
                  </div>
                ))}
                {spaceMembers?.items.length === 0 && (
                  <p className="text-sm text-subtle">{t("admin.spaceMembers.noMembers")}</p>
                )}
              </div>
            )}

            <h4 className="mt-6 text-xs font-semibold uppercase tracking-wide text-subtle">
              {t("admin.spaceGroups.title")}
            </h4>
            <p className="mt-1 text-xs text-muted">{t("admin.spaceGroups.hint")}</p>

            <div className="mt-4 flex flex-wrap gap-2">
              <SelectField
                className="min-w-[10rem] flex-1"
                value={selectedGroupId}
                onChange={(e) => setSelectedGroupId(e.target.value)}
              >
                <option value="">{t("admin.spaceGroups.selectGroup")}</option>
                {availableGroups.map((g) => (
                  <option key={g.id} value={g.id}>
                    {g.name}
                  </option>
                ))}
              </SelectField>
              <SelectField
                className="w-36"
                value={groupRole}
                onChange={(e) => setGroupRole(e.target.value)}
              >
                {SPACE_ROLES.map((role) => (
                  <option key={role} value={role}>
                    {t(`admin.groups.roleLabels.${role}`)}
                  </option>
                ))}
              </SelectField>
              <button
                type="button"
                className="btn-primary"
                disabled={!selectedGroupId || assignGroup.isPending}
                onClick={() =>
                  editingId &&
                  selectedGroupId &&
                  assignGroup.mutate({ spaceId: editingId, groupId: selectedGroupId, role: groupRole })
                }
              >
                {assignGroup.isPending ? (
                  <Loader2 className="h-4 w-4 animate-spin" />
                ) : (
                  t("admin.spaceGroups.assignButton")
                )}
              </button>
            </div>

            {groupsLoading ? (
              <div className="flex justify-center py-8">
                <Loader2 className="h-5 w-5 animate-spin text-primary" />
              </div>
            ) : (
              <div className="mt-4 space-y-2">
                {spaceGroups?.items.map((grp) => (
                  <div
                    key={grp.group_id}
                    className="flex flex-wrap items-center justify-between gap-3 rounded-xl border border-default bg-surface-muted px-4 py-3"
                  >
                    <div>
                      <p className="font-medium text-fg">{grp.group_name}</p>
                      {grp.description && <p className="text-xs text-subtle">{grp.description}</p>}
                    </div>
                    <div className="flex items-center gap-2">
                      <span className="badge badge-primary">
                        {t(`admin.groups.roleLabels.${grp.role as (typeof SPACE_ROLES)[number]}`) || grp.role}
                      </span>
                      <button
                        type="button"
                        className="btn-ghost text-danger-soft !px-2"
                        disabled={removingGroupId === grp.group_id}
                        onClick={() => removeSpaceGroup(grp.group_id)}
                        aria-label={t("admin.spaceGroups.remove")}
                      >
                        {removingGroupId === grp.group_id ? (
                          <Loader2 className="h-4 w-4 animate-spin" />
                        ) : (
                          <Trash2 className="h-4 w-4" />
                        )}
                      </button>
                    </div>
                  </div>
                ))}
                {spaceGroups?.items.length === 0 && (
                  <p className="text-sm text-subtle">{t("admin.spaceGroups.noGroups")}</p>
                )}
              </div>
            )}
          </div>

          <div className="mt-8 border-t border-default pt-6">
            <h3 className="text-sm font-semibold text-fg">{t("admin.spaces.linkedRepos")}</h3>
            <p className="mt-1 text-xs text-muted">{t("admin.spaces.linkedReposHint")}</p>
            {reposLoading ? (
              <div className="flex justify-center py-8">
                <Loader2 className="h-5 w-5 animate-spin text-primary" />
              </div>
            ) : (
              <div className="mt-4 space-y-2">
                {spaceRepos?.items.map((repo) => (
                  <div
                    key={repo.id}
                    className="flex flex-wrap items-center justify-between gap-3 rounded-xl border border-default bg-surface-muted px-4 py-3"
                  >
                    <div className="min-w-0">
                      <p className="font-medium text-fg">{repo.name}</p>
                      <p className="truncate text-xs text-subtle">{repo.url}</p>
                      <p className="text-xs text-subtle">
                        {repo.branch} · {repo.docs_path}
                      </p>
                    </div>
                    <button
                      type="button"
                      className="btn-ghost text-danger-soft !px-2"
                      disabled={removingRepoId === repo.id}
                      onClick={() => confirmRemoveRepo(repo)}
                      aria-label={t("admin.spaces.removeRepo")}
                    >
                      {removingRepoId === repo.id ? (
                        <Loader2 className="h-4 w-4 animate-spin" />
                      ) : (
                        <Trash2 className="h-4 w-4" />
                      )}
                    </button>
                  </div>
                ))}
                {spaceRepos?.items.length === 0 && (
                  <p className="text-sm text-subtle">{t("admin.spaces.noRepos")}</p>
                )}
              </div>
            )}
          </div>
          <SpacePageACLPanel spaceId={editingId} />
        </div>
      )}

      <div className="glass p-6">
        <h2 className="text-lg font-semibold text-fg">{t("admin.spaces.title")}</h2>
        <p className="mt-1 text-sm text-muted">{t("admin.spaces.subtitle")}</p>

        {isLoading ? (
          <div className="flex justify-center py-12">
            <Loader2 className="h-6 w-6 animate-spin text-primary" />
          </div>
        ) : (
          <div className="mt-6 space-y-2">
            {data?.items.map((space) => (
              <div
                key={space.id}
                className={`flex flex-wrap items-center justify-between gap-4 rounded-xl border border-default px-4 py-3 ${
                  editingId === space.id ? "bg-highlight-row" : "bg-surface-muted"
                }`}
              >
                <div>
                  <p className="font-medium text-fg">{space.name}</p>
                  <p className="text-xs text-subtle">/{space.slug}</p>
                  {space.description && <p className="mt-1 text-sm text-muted">{space.description}</p>}
                </div>
                <div className="flex items-center gap-3">
                  {space.is_public && (
                    <span className="badge badge-primary">
                      <Globe className="mr-1 inline h-3 w-3" />
                      {t("common.public")}
                    </span>
                  )}
                  <button
                    type="button"
                    className="btn-ghost !px-2"
                    onClick={() => startEdit(space)}
                    aria-label={t("admin.spaces.editSpace", { name: space.name })}
                  >
                    <Pencil className="h-4 w-4" />
                  </button>
                </div>
              </div>
            ))}
            {data?.items.length === 0 && <p className="text-sm text-subtle">{t("admin.spaces.noSpaces")}</p>}
          </div>
        )}
        {error && !editingId && <p className="mt-4 text-sm text-danger-soft">{error}</p>}
      </div>

      <div className="glass mt-6 p-6">
        <h2 className="text-lg font-semibold text-fg">{t("admin.spaces.createTitle")}</h2>
        <form
          className="mt-4 space-y-4"
          onSubmit={(e) => {
            e.preventDefault();
            createSpace.mutate({
              slug,
              name,
              description: description || undefined,
              is_public: isPublic,
            });
          }}
        >
          <input
            className="input-field"
            placeholder={t("admin.spaces.slugPlaceholder")}
            value={slug}
            onChange={(e) => setSlug(e.target.value)}
            required
          />
          <input
            className="input-field"
            placeholder={t("admin.spaces.namePlaceholder")}
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
          />
          <input
            className="input-field"
            placeholder={t("admin.spaces.descriptionPlaceholder")}
            value={description}
            onChange={(e) => setDescription(e.target.value)}
          />
          <label className="checkbox-label sm:col-span-2">
            <input
              type="checkbox"
              checked={isPublic}
              onChange={(e) => setIsPublic(e.target.checked)}
              className="checkbox-field"
            />
            {t("admin.spaces.publishPublic")}
          </label>
          {error && !editingId && <p className="text-sm text-danger-soft">{error}</p>}
          <button type="submit" className="btn-primary" disabled={createSpace.isPending}>
            {createSpace.isPending ? <Loader2 className="h-4 w-4 animate-spin" /> : t("admin.spaces.create")}
          </button>
        </form>
      </div>
    </FadeIn>
  );
}
