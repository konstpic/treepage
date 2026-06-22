import { useEffect } from "react";
import { Link, Navigate, Outlet, useLocation, useNavigate } from "react-router-dom";
import {
  BarChart3,
  ClipboardList,
  Database,
  FolderOpen,
  GitBranch,
  Settings,
  Shield,
  Users,
  UsersRound,
} from "lucide-react";
import { useAuthStore } from "@/lib/store";
import { cn } from "@/lib/utils";
import { FadeIn } from "@/components/motion-wrapper";
import { useI18n } from "@/lib/i18n";

function isAdmin(user: { roles: string[] } | null) {
  return user?.roles.some((r) => ["super_admin", "admin"].includes(r)) ?? false;
}

export function AdminLayout() {
  const { isAuthenticated, isHydrated, user } = useAuthStore();
  const navigate = useNavigate();
  const location = useLocation();
  const { t } = useI18n();

  const nav = [
    { href: "/admin/spaces", label: t("admin.nav.spaces"), icon: FolderOpen },
    { href: "/admin/repositories", label: t("admin.nav.repositories"), icon: GitBranch },
    { href: "/admin/users", label: t("admin.nav.users"), icon: Users },
    { href: "/admin/groups", label: t("admin.nav.groups"), icon: UsersRound },
    { href: "/admin/settings", label: t("admin.nav.settings"), icon: Settings },
    { href: "/admin/analytics", label: t("admin.nav.analytics"), icon: BarChart3 },
    { href: "/admin/rag", label: t("admin.nav.rag"), icon: Database },
    { href: "/admin/audit", label: t("admin.nav.audit"), icon: ClipboardList, roles: ["super_admin"] as const },
    { href: "/admin/oidc", label: t("admin.nav.oidc"), icon: Shield, roles: ["super_admin"] as const },
  ];

  useEffect(() => {
    if (isHydrated && !isAuthenticated) navigate("/auth");
    if (isHydrated && user && !isAdmin(user)) navigate("/spaces");
  }, [isHydrated, isAuthenticated, user, navigate]);

  if (!isHydrated || !user || !isAdmin(user)) return null;

  const links = nav.filter((item) => {
    if (!item.roles) return true;
    return item.roles.some((r) => user.roles.includes(r));
  });

  return (
    <div className="mx-auto max-w-6xl px-4 py-10 sm:px-6">
      <FadeIn>
        <div className="mb-8 flex items-center gap-3">
          <Database className="h-8 w-8 text-primary" />
          <div>
            <h1 className="text-3xl font-bold gradient-text">{t("admin.title")}</h1>
            <p className="text-muted">{t("admin.subtitle")}</p>
          </div>
        </div>
      </FadeIn>

      <div className="flex flex-col gap-8 lg:flex-row">
        <aside className="lg:w-56 shrink-0">
          <nav className="glass flex flex-row gap-1 overflow-x-auto p-2 lg:flex-col">
            {links.map((item) => {
              const Icon = item.icon;
              const active = location.pathname === item.href || location.pathname.startsWith(`${item.href}/`);
              return (
                <Link
                  key={item.href}
                  to={item.href}
                  className={cn(
                    "nav-link flex items-center gap-2 whitespace-nowrap px-4 py-2.5",
                    active && "nav-link-active"
                  )}
                >
                  <Icon className="h-4 w-4 shrink-0" />
                  {item.label}
                </Link>
              );
            })}
          </nav>
        </aside>

        <div className="min-w-0 flex-1">
          <Outlet />
        </div>
      </div>
    </div>
  );
}

export function AdminIndexRedirect() {
  return <Navigate to="/admin/spaces" replace />;
}

export function useAdminGuard() {
  const { isAuthenticated, isHydrated, user } = useAuthStore();
  const navigate = useNavigate();

  useEffect(() => {
    if (isHydrated && !isAuthenticated) navigate("/auth");
    if (isHydrated && user && !isAdmin(user)) navigate("/spaces");
  }, [isHydrated, isAuthenticated, user, navigate]);

  return { ready: isHydrated && !!user && isAdmin(user), user };
}
