import { useEffect, useLayoutEffect, useMemo, useState } from "react";
import { GraduationCap, X } from "lucide-react";
import { useLocation, useNavigate } from "react-router-dom";
import { useOnboardingStore } from "@/lib/onboarding-store";
import { useAuthStore } from "@/lib/store";
import { useSplashStore } from "@/lib/splash-store";
import { useI18n } from "@/lib/i18n";

type TourStepTitleKey = keyof typeof import("@/lib/i18n/en").en.tour.steps;

type TourStep = {
  id: string;
  selector: string;
  titleKey: TourStepTitleKey;
  route?: string;
  adminOnly?: boolean;
};

/** Default doc for tree/comments steps (Welcome space). */
const TOUR_DOC_ROUTE = "/spaces/welcome/docs/welcome";

const TOUR_STEPS: TourStep[] = [
  { id: "spaces-nav", selector: '[data-tour="nav-spaces"]', titleKey: "navSpaces", route: "/spaces" },
  { id: "spaces-page", selector: '[data-tour="spaces-main"]', titleKey: "spacesSection", route: "/spaces" },
  { id: "search-nav", selector: '[data-tour="nav-search"]', titleKey: "navSearch", route: "/search" },
  { id: "search-page", selector: '[data-tour="search-main"]', titleKey: "searchSection", route: "/search" },
  { id: "me-nav", selector: '[data-tour="nav-me"]', titleKey: "navAccount", route: "/me" },
  { id: "me-page", selector: '[data-tour="me-main"]', titleKey: "meSection", route: "/me" },
  {
    id: "admin-nav",
    selector: '[data-tour="nav-admin"]',
    titleKey: "navAdmin",
    route: "/admin/spaces",
    adminOnly: true,
  },
  {
    id: "admin-page",
    selector: '[data-tour="admin-nav"]',
    titleKey: "adminSection",
    route: "/admin/spaces",
    adminOnly: true,
  },
  { id: "doc-tree", selector: '[data-tour="doc-tree"]', titleKey: "docTree", route: TOUR_DOC_ROUTE },
  { id: "doc-comments", selector: '[data-tour="doc-comments"]', titleKey: "docComments", route: TOUR_DOC_ROUTE },
];

function routeMatches(pathname: string, route: string): boolean {
  return pathname === route || pathname.startsWith(`${route}/`);
}

function spotlightClipPath(spotlight: { top: number; left: number; width: number; height: number }): string {
  const { top, left, width, height } = spotlight;
  const x1 = Math.round(left);
  const y1 = Math.round(top);
  const x2 = Math.round(left + width);
  const y2 = Math.round(top + height);
  return `polygon(evenodd, 0 0, 100vw 0, 100vw 100vh, 0 100vh, 0 0, ${x1}px ${y1}px, ${x2}px ${y1}px, ${x2}px ${y2}px, ${x1}px ${y2}px, ${x1}px ${y1}px)`;
}

function useTourSteps(): TourStep[] {
  const { user } = useAuthStore();
  const isAdmin = user?.roles.some((r) => ["super_admin", "admin"].includes(r)) ?? false;
  return useMemo(() => TOUR_STEPS.filter((s) => !s.adminOnly || isAdmin), [isAdmin]);
}

export function ProductTourTrigger() {
  const { t } = useI18n();
  const restart = useOnboardingStore((s) => s.restart);
  const { isAuthenticated } = useAuthStore();

  if (!isAuthenticated) return null;

  return (
    <button
      type="button"
      className="btn-ghost !px-2"
      onClick={restart}
      title={t("tour.restart")}
      aria-label={t("tour.restart")}
      data-tour="tour-restart"
    >
      <GraduationCap className="h-4 w-4" />
    </button>
  );
}

export function ProductTourOverlay() {
  const { t } = useI18n();
  const navigate = useNavigate();
  const location = useLocation();
  const { isAuthenticated } = useAuthStore();
  const splashIdle = useSplashStore((s) => s.phase === "idle");
  const active = useOnboardingStore((s) => s.active);
  const step = useOnboardingStore((s) => s.step);
  const next = useOnboardingStore((s) => s.next);
  const back = useOnboardingStore((s) => s.back);
  const skip = useOnboardingStore((s) => s.skip);
  const finish = useOnboardingStore((s) => s.finish);
  const start = useOnboardingStore((s) => s.start);
  const shouldAutoStart = useOnboardingStore((s) => s.shouldAutoStart);

  const steps = useTourSteps();
  const current = steps[step];
  const [rect, setRect] = useState<DOMRect | null>(null);
  const [navReady, setNavReady] = useState(false);

  useEffect(() => {
    if (!isAuthenticated || !shouldAutoStart() || !splashIdle) return;
    const timer = window.setTimeout(() => start(), 500);
    return () => window.clearTimeout(timer);
  }, [isAuthenticated, shouldAutoStart, start, splashIdle]);

  useEffect(() => {
    if (!active || !current?.route) {
      setNavReady(true);
      return;
    }
    setNavReady(false);
    if (!routeMatches(location.pathname, current.route)) {
      navigate(current.route);
      return;
    }
    const timer = window.setTimeout(() => setNavReady(true), 120);
    return () => window.clearTimeout(timer);
  }, [active, current, location.pathname, navigate]);

  useLayoutEffect(() => {
    if (!active || !current || !navReady) {
      setRect(null);
      return;
    }

    let cancelled = false;
    let attempts = 0;

    const measure = () => {
      if (cancelled) return;
      const el = document.querySelector(current.selector);
      if (el) {
        setRect(el.getBoundingClientRect());
        el.scrollIntoView({ block: "nearest", behavior: "smooth" });
        return;
      }
      if (attempts++ < 30) {
        window.requestAnimationFrame(measure);
      } else {
        setRect(null);
      }
    };

    measure();
    const onResize = () => {
      const el = document.querySelector(current.selector);
      if (el) setRect(el.getBoundingClientRect());
    };
    window.addEventListener("resize", onResize);
    window.addEventListener("scroll", onResize, true);
    return () => {
      cancelled = true;
      window.removeEventListener("resize", onResize);
      window.removeEventListener("scroll", onResize, true);
    };
  }, [active, current, step, navReady, location.pathname]);

  if (!active || !current || !splashIdle) return null;

  const stepText = t(`tour.steps.${current.titleKey}.title`);
  const bodyText = t(`tour.steps.${current.titleKey}.body`);
  const isLast = step >= steps.length - 1;

  const pad = 8;
  const spotlight = rect
    ? {
        top: rect.top - pad,
        left: rect.left - pad,
        width: rect.width + pad * 2,
        height: rect.height + pad * 2,
      }
    : null;

  const cardStyle: React.CSSProperties = spotlight
    ? {
        top: Math.min(spotlight.top + spotlight.height + 12, window.innerHeight - 180),
        left: Math.min(Math.max(16, spotlight.left), window.innerWidth - 336),
      }
    : { top: "50%", left: "50%", transform: "translate(-50%, -50%)" };

  return (
    <div className="product-tour" role="dialog" aria-modal="true" aria-label={stepText}>
      <div
        className="product-tour__backdrop"
        style={spotlight ? { clipPath: spotlightClipPath(spotlight) } : undefined}
        onClick={skip}
        aria-hidden
      />
      {spotlight && (
        <div
          className="product-tour__spotlight"
          style={{
            top: spotlight.top,
            left: spotlight.left,
            width: spotlight.width,
            height: spotlight.height,
          }}
        />
      )}
      <div className="product-tour__card" style={cardStyle} onClick={(e) => e.stopPropagation()}>
        <div className="flex items-start justify-between gap-3">
          <div>
            <p className="text-xs text-subtle">
              {t("tour.stepOf", { current: step + 1, total: steps.length })}
            </p>
            <h3 className="mt-1 text-base font-semibold text-fg">{stepText}</h3>
          </div>
          <button type="button" className="btn-ghost !p-1" onClick={skip} aria-label={t("tour.skip")}>
            <X className="h-4 w-4" />
          </button>
        </div>
        <p className="mt-2 text-sm text-muted">{bodyText}</p>
        <div className="mt-4 flex flex-wrap justify-between gap-2">
          <button type="button" className="btn-ghost !text-sm" onClick={skip}>
            {t("tour.skip")}
          </button>
          <div className="flex gap-2">
            {step > 0 && (
              <button type="button" className="btn-secondary !text-sm" onClick={back}>
                {t("tour.back")}
              </button>
            )}
            <button
              type="button"
              className="btn-primary !text-sm"
              onClick={() => (isLast ? finish() : next())}
            >
              {isLast ? t("tour.finish") : t("tour.next")}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
