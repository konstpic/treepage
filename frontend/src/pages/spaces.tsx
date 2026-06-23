import { useEffect, useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { AnimatePresence, motion } from "framer-motion";
import { ArrowDownAZ, ArrowUpAZ, FolderOpen, Hash, LayoutGrid, List, Loader2, Plus, Search, Table2 } from "lucide-react";
import { optionalAuthApi, getPublicSpaces, ApiError } from "@/lib/api";
import { useAuthStore } from "@/lib/store";
import { FadeIn } from "@/components/motion-wrapper";
import { useI18n } from "@/lib/i18n";
import { cn } from "@/lib/utils";
import {
  readSpacesSort,
  sortSpaces,
  writeSpacesSort,
  type SpacesSortMode,
} from "@/lib/spaces-sort";
import {
  readSpacesView,
  writeSpacesView,
  type SpacesViewMode,
} from "@/lib/spaces-view";

const searchEase: [number, number, number, number] = [0.25, 0.46, 0.45, 0.94];

const COLLAPSED_HEIGHT = 32;
const EXPANDED_HEIGHT = 44;

interface SpacesToolbarProps {
  search: string;
  onSearchChange: (value: string) => void;
  placeholder: string;
  resultText?: string;
  sortMode: SpacesSortMode;
  onSortModeChange: (mode: SpacesSortMode) => void;
  sortModeLabel: string;
  viewMode: SpacesViewMode;
  onViewModeChange: (mode: SpacesViewMode) => void;
  viewModeLabel: string;
  label: (key: string) => string;
}

const SORT_OPTIONS: { id: SpacesSortMode; icon: typeof ArrowDownAZ; labelKey: string }[] = [
  { id: "name_asc", icon: ArrowDownAZ, labelKey: "spaces.sortNameAsc" },
  { id: "name_desc", icon: ArrowUpAZ, labelKey: "spaces.sortNameDesc" },
  { id: "slug_asc", icon: Hash, labelKey: "spaces.sortSlugAsc" },
];

function SpacesToolbar({
  search,
  onSearchChange,
  placeholder,
  resultText,
  sortMode,
  onSortModeChange,
  sortModeLabel,
  viewMode,
  onViewModeChange,
  viewModeLabel,
  label,
}: SpacesToolbarProps) {
  const [hovered, setHovered] = useState(false);
  const [searchFocused, setSearchFocused] = useState(false);
  const expanded = hovered || searchFocused || search.trim().length > 0;

  return (
    <div className="mb-6">
      <div
        className="flex items-start gap-3"
        onMouseEnter={() => setHovered(true)}
        onMouseLeave={() => setHovered(false)}
      >
        <div className="min-w-0 flex-1">
          <motion.div
            className={cn(
              "spaces-search flex items-center",
              expanded && "spaces-search--expanded"
            )}
            initial={false}
            animate={{ height: expanded ? EXPANDED_HEIGHT : COLLAPSED_HEIGHT }}
            transition={{ duration: 0.28, ease: searchEase }}
          >
            <motion.div
              className="pointer-events-none flex shrink-0 items-center justify-center pl-3"
              animate={{ opacity: expanded ? 1 : 0.55 }}
              transition={{ duration: 0.2 }}
            >
              <Search className="h-4 w-4 text-subtle" />
            </motion.div>
            <input
              type="search"
              value={search}
              onChange={(e) => onSearchChange(e.target.value)}
              onFocus={() => setSearchFocused(true)}
              onBlur={() => setSearchFocused(false)}
              placeholder={placeholder}
              aria-label={placeholder}
              aria-expanded={expanded}
              className="spaces-search-input min-h-0 flex-1"
            />
          </motion.div>
        </div>

        <motion.div
          className={cn("view-toggle", expanded && "view-toggle--expanded")}
          role="group"
          aria-label={sortModeLabel}
          initial={false}
          animate={{ height: expanded ? EXPANDED_HEIGHT : COLLAPSED_HEIGHT }}
          transition={{ duration: 0.28, ease: searchEase }}
        >
          {SORT_OPTIONS.map(({ id, icon: Icon, labelKey }) => {
            const active = sortMode === id;
            return (
              <motion.button
                key={id}
                type="button"
                className={cn("view-toggle-btn", active && "view-toggle-btn-active")}
                aria-label={label(labelKey)}
                aria-pressed={active}
                onClick={() => onSortModeChange(id)}
                animate={{
                  opacity: expanded || active ? 1 : 0.55,
                  scale: expanded ? 1 : active ? 1 : 0.92,
                }}
                transition={{ duration: 0.2 }}
              >
                <Icon className="h-4 w-4" />
              </motion.button>
            );
          })}
        </motion.div>

        <motion.div
          className={cn("view-toggle", expanded && "view-toggle--expanded")}
          role="group"
          aria-label={viewModeLabel}
          initial={false}
          animate={{ height: expanded ? EXPANDED_HEIGHT : COLLAPSED_HEIGHT }}
          transition={{ duration: 0.28, ease: searchEase }}
        >
          {VIEW_OPTIONS.map(({ id, icon: Icon, labelKey }) => {
            const active = viewMode === id;
            return (
              <motion.button
                key={id}
                type="button"
                className={cn("view-toggle-btn", active && "view-toggle-btn-active")}
                aria-label={label(labelKey)}
                aria-pressed={active}
                onClick={() => onViewModeChange(id)}
                animate={{
                  opacity: expanded || active ? 1 : 0.55,
                  scale: expanded ? 1 : active ? 1 : 0.92,
                }}
                transition={{ duration: 0.2 }}
              >
                <Icon className="h-4 w-4" />
              </motion.button>
            );
          })}
        </motion.div>
      </div>

      <AnimatePresence>
        {resultText && (
          <motion.p
            key="results"
            initial={{ opacity: 0, y: -4 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -4 }}
            transition={{ duration: 0.2 }}
            className="mt-2 text-sm text-subtle"
          >
            {resultText}
          </motion.p>
        )}
      </AnimatePresence>
    </div>
  );
}

const VIEW_OPTIONS: { id: SpacesViewMode; icon: typeof LayoutGrid; labelKey: string }[] = [
  { id: "grid", icon: LayoutGrid, labelKey: "spaces.viewGrid" },
  { id: "table", icon: Table2, labelKey: "spaces.viewTable" },
  { id: "list", icon: List, labelKey: "spaces.viewList" },
];

interface Space {
  id: string;
  slug: string;
  name: string;
  description?: string;
  is_public: boolean;
}

function matchesQuery(space: Space, query: string) {
  const q = query.trim().toLowerCase();
  if (!q) return true;
  return (
    space.name.toLowerCase().includes(q) ||
    space.slug.toLowerCase().includes(q) ||
    (space.description?.toLowerCase().includes(q) ?? false)
  );
}

function SpaceGridView({ spaces, t }: { spaces: Space[]; t: (key: string) => string }) {
  return (
    <div className="grid auto-rows-fr gap-4 sm:grid-cols-2 lg:grid-cols-3">
      {spaces.map((space, i) => (
        <FadeIn key={space.id} delay={i * 0.04} className="h-full min-h-0">
          <Link
            to={`/spaces/${space.slug}`}
            className="glass-hover flex h-full min-h-[11.5rem] flex-col p-6"
          >
            <FolderOpen className="h-8 w-8 shrink-0 text-primary" />
            <h2 className="mt-4 line-clamp-2 text-lg font-semibold leading-snug text-fg">
              {space.name}
            </h2>
            <p className="mt-2 line-clamp-2 flex-1 text-sm leading-relaxed text-muted">
              {space.description || "\u00a0"}
            </p>
            <div className="mt-4 flex min-h-[1.75rem] items-end">
              {space.is_public && (
                <span className="badge badge-success">{t("common.public")}</span>
              )}
            </div>
          </Link>
        </FadeIn>
      ))}
    </div>
  );
}

function SpaceTableView({ spaces, t }: { spaces: Space[]; t: (key: string) => string }) {
  return (
    <FadeIn>
      <div className="glass overflow-x-auto">
        <table className="w-full min-w-[640px] text-left text-sm">
          <thead>
            <tr className="border-b border-default text-xs uppercase tracking-wide text-subtle">
              <th className="px-4 py-3 font-medium">{t("spaces.tableName")}</th>
              <th className="px-4 py-3 font-medium">{t("spaces.tableSlug")}</th>
              <th className="px-4 py-3 font-medium">{t("spaces.tableDescription")}</th>
              <th className="px-4 py-3 font-medium">{t("spaces.tableAccess")}</th>
            </tr>
          </thead>
          <tbody>
            {spaces.map((space) => (
              <tr key={space.id} className="border-b border-default/60 last:border-0">
                <td className="px-4 py-3">
                  <Link
                    to={`/spaces/${space.slug}`}
                    className="font-medium text-fg transition-colors hover:text-primary"
                  >
                    {space.name}
                  </Link>
                </td>
                <td className="px-4 py-3 text-subtle">
                  <code>/{space.slug}</code>
                </td>
                <td className="max-w-xs px-4 py-3 text-muted">
                  <span className="line-clamp-2">{space.description || "—"}</span>
                </td>
                <td className="px-4 py-3">
                  {space.is_public ? (
                    <span className="badge badge-success">{t("common.public")}</span>
                  ) : (
                    <span className="text-subtle">—</span>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </FadeIn>
  );
}

function SpaceListView({ spaces, t }: { spaces: Space[]; t: (key: string) => string }) {
  return (
    <div className="space-y-2">
      {spaces.map((space, i) => (
        <FadeIn key={space.id} delay={i * 0.03}>
          <Link
            to={`/spaces/${space.slug}`}
            className="glass-hover flex items-start gap-4 p-4 sm:items-center"
          >
            <FolderOpen className="mt-0.5 h-6 w-6 shrink-0 text-primary sm:mt-0" />
            <div className="min-w-0 flex-1">
              <div className="flex flex-wrap items-center gap-x-3 gap-y-1">
                <h2 className="font-semibold text-fg">{space.name}</h2>
                <code className="text-xs text-subtle">/{space.slug}</code>
                {space.is_public && (
                  <span className="badge badge-success">{t("common.public")}</span>
                )}
              </div>
              {space.description && (
                <p className="mt-1 line-clamp-2 text-sm text-muted">{space.description}</p>
              )}
            </div>
          </Link>
        </FadeIn>
      ))}
    </div>
  );
}

export function SpacesPage() {
  const { isAuthenticated, isHydrated, user } = useAuthStore();
  const { t } = useI18n();
  const [search, setSearch] = useState("");
  const [viewMode, setViewMode] = useState<SpacesViewMode>(() => readSpacesView(user?.id));
  const [sortMode, setSortMode] = useState<SpacesSortMode>(() => readSpacesSort(user?.id));
  const isAdmin = user?.roles.some((r) => ["super_admin", "admin"].includes(r)) ?? false;

  useEffect(() => {
    setViewMode(readSpacesView(user?.id));
    setSortMode(readSpacesSort(user?.id));
  }, [user?.id]);

  function changeViewMode(mode: SpacesViewMode) {
    setViewMode(mode);
    writeSpacesView(user?.id, mode);
  }

  function changeSortMode(mode: SpacesSortMode) {
    setSortMode(mode);
    writeSpacesSort(user?.id, mode);
  }

  const { data, isLoading, error } = useQuery({
    queryKey: ["spaces", isAuthenticated],
    queryFn: () =>
      isAuthenticated
        ? optionalAuthApi<{ items: Space[] }>("/api/spaces")
        : getPublicSpaces(),
    enabled: isHydrated,
  });

  const items = data?.items ?? [];
  const filtered = useMemo(() => items.filter((s) => matchesQuery(s, search)), [items, search]);
  const sorted = useMemo(() => sortSpaces(filtered, sortMode), [filtered, sortMode]);
  const showToolbar = items.length > 0;

  if (!isHydrated) return null;

  return (
    <div className="mx-auto max-w-6xl px-4 py-10 sm:px-6" data-tour="spaces-main">
      <div className="mb-8 flex flex-wrap items-center justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold gradient-text">{t("spaces.title")}</h1>
          <p className="mt-1 text-muted">
            {isAuthenticated ? t("spaces.subtitleAuth") : t("spaces.subtitlePublic")}
          </p>
        </div>
        {isAdmin && (
          <Link to="/admin/spaces" className="btn-secondary !py-2 !px-4">
            <Plus className="h-4 w-4" />
            {t("spaces.manage")}
          </Link>
        )}
        {!isAuthenticated && (
          <Link to="/auth" className="btn-secondary !py-2 !px-4">
            {t("common.signIn")}
          </Link>
        )}
      </div>

      {showToolbar && (
        <SpacesToolbar
          search={search}
          onSearchChange={setSearch}
          placeholder={t("spaces.searchPlaceholder")}
          resultText={
            search.trim()
              ? t("spaces.searchResults", { count: filtered.length, total: items.length })
              : undefined
          }
          viewMode={viewMode}
          onViewModeChange={changeViewMode}
          viewModeLabel={t("spaces.viewMode")}
          sortMode={sortMode}
          onSortModeChange={changeSortMode}
          sortModeLabel={t("spaces.sortMode")}
          label={t}
        />
      )}

      {isLoading && (
        <div className="flex justify-center py-20">
          <Loader2 className="h-8 w-8 animate-spin text-primary" />
        </div>
      )}

      {error instanceof ApiError && (
        <div className="glass p-6 text-danger-soft">{error.message}</div>
      )}

      {!isLoading && sorted.length > 0 && viewMode === "grid" && (
        <SpaceGridView spaces={sorted} t={t} />
      )}
      {!isLoading && sorted.length > 0 && viewMode === "table" && (
        <SpaceTableView spaces={sorted} t={t} />
      )}
      {!isLoading && sorted.length > 0 && viewMode === "list" && (
        <SpaceListView spaces={sorted} t={t} />
      )}

      {!isLoading && items.length === 0 && (
        <div className="glass p-12 text-center text-muted">
          {isAuthenticated ? t("spaces.noSpacesAuth") : t("spaces.noSpacesPublic")}
        </div>
      )}

      {!isLoading && items.length > 0 && filtered.length === 0 && (
        <div className="glass p-12 text-center text-muted">
          {t("spaces.noSearchResults", { query: search.trim() })}
        </div>
      )}
    </div>
  );
}
