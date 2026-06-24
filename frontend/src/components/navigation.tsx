import { Link, useLocation, useNavigate } from "react-router-dom";
import { useState } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { Menu, X, LogOut, User, ChevronRight, Search } from "lucide-react";
import { NotificationsBell } from "@/components/notifications-bell";
import { ProductTourTrigger } from "@/components/product-tour";
import { TreePageLogo } from "@/components/treepage-logo";
import { useAuthStore, useBrandingStore } from "@/lib/store";
import { useI18n } from "@/lib/i18n";
import { cn } from "@/lib/utils";

export function Navigation() {
  const location = useLocation();
  const navigate = useNavigate();
  const { t } = useI18n();
  const { isAuthenticated, isHydrated, logout, user } = useAuthStore();
  const projectName = useBrandingStore((s) => s.projectName);
  const [mobileOpen, setMobileOpen] = useState(false);

  const navLinks = [
    { href: "/spaces", label: t("nav.spaces") },
    { href: "/search", label: t("nav.search") },
    { href: "/me", label: t("nav.myPages"), auth: true },
    { href: "/admin", label: t("nav.admin"), auth: true, roles: ["super_admin", "admin"] as const },
  ];

  const visibleLinks = navLinks.filter((l) => {
    if (l.auth && !isAuthenticated) return false;
    if (l.roles && user) {
      const hasRole = l.roles.some((r) => user.roles.includes(r) || user.roles.includes("super_admin"));
      if (!hasRole) return false;
    }
    return true;
  });

  function handleLogout() {
    logout();
    navigate("/");
  }

  return (
    <header className="nav-header">
      <div className="mx-auto flex w-full max-w-[min(100%,100rem)] items-center justify-between px-4 py-3 sm:px-6 lg:px-8 2xl:px-10">
        <Link to="/" className="group flex items-center gap-2.5">
          <div className="logo-icon flex h-8 w-8 items-center justify-center rounded-lg transition-shadow group-hover:shadow-lg">
            <TreePageLogo size={22} animate={false} variant="onPrimary" />
          </div>
          <span className="text-lg font-bold gradient-text">{projectName}</span>
        </Link>

        <nav className="hidden items-center gap-1 sm:flex">
          {visibleLinks.map((link) => (
            <Link
              key={link.href}
              to={link.href}
              data-tour={`nav-${link.href.replace("/", "") || "home"}`}
              className={cn(
                "nav-link",
                location.pathname.startsWith(link.href) && "nav-link-active"
              )}
            >
              {link.label}
            </Link>
          ))}
        </nav>

        <div className="hidden items-center gap-2 sm:flex">
          <Link to="/search" className="btn-ghost" data-tour="nav-search">
            <Search className="h-4 w-4" />
          </Link>
          {isHydrated && isAuthenticated ? (
            <>
              <NotificationsBell />
              <ProductTourTrigger />
              <Link to="/me" className="btn-ghost" data-tour="nav-me">
                <User className="h-4 w-4" />
                {user?.display_name || t("common.account")}
              </Link>
              <button onClick={handleLogout} className="btn-ghost text-subtle hover:text-danger-soft">
                <LogOut className="h-4 w-4" />
              </button>
            </>
          ) : isHydrated ? (
            <Link to="/auth" className="btn-primary !py-2 !px-5 !text-sm">
              {t("nav.signIn")}
              <ChevronRight className="h-4 w-4" />
            </Link>
          ) : null}
        </div>

        <button
          className="nav-link rounded-lg p-2 sm:hidden"
          onClick={() => setMobileOpen(!mobileOpen)}
          aria-label={t("common.menu")}
        >
          {mobileOpen ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
        </button>
      </div>

      <AnimatePresence>
        {mobileOpen && (
          <motion.div
            initial={{ height: 0, opacity: 0 }}
            animate={{ height: "auto", opacity: 1 }}
            exit={{ height: 0, opacity: 0 }}
            className="overflow-hidden border-t border-default sm:hidden"
          >
            <nav className="flex flex-col gap-1 p-4">
              {visibleLinks.map((link) => (
                <Link
                  key={link.href}
                  to={link.href}
                  onClick={() => setMobileOpen(false)}
                  className={cn(
                    "nav-link px-4 py-3",
                    location.pathname.startsWith(link.href) && "nav-link-active"
                  )}
                >
                  {link.label}
                </Link>
              ))}
            </nav>
          </motion.div>
        )}
      </AnimatePresence>
    </header>
  );
}
