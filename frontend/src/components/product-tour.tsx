import { useEffect, useLayoutEffect, useMemo, useState } from "react";
import { GraduationCap, X } from "lucide-react";
import { useLocation } from "react-router-dom";
import { useOnboardingStore } from "@/lib/onboarding-store";
import { useAuthStore } from "@/lib/store";
import { useI18n } from "@/lib/i18n";

type TourStep = {
  id: string;
  selector: string;
  titleKey: keyof typeof import("@/lib/i18n/en").en.tour.steps;
};

const MAIN_STEPS: TourStep[] = [
  { id: "spaces", selector: '[data-tour="nav-spaces"]', titleKey: "navSpaces" },
  { id: "search", selector: '[data-tour="nav-search"]', titleKey: "navSearch" },
  { id: "account", selector: '[data-tour="nav-me"]', titleKey: "navAccount" },
  { id: "admin", selector: '[data-tour="nav-admin"]', titleKey: "navAdmin" },
];

const DOC_STEPS: TourStep[] = [
  { id: "tree", selector: '[data-tour="doc-tree"]', titleKey: "docTree" },
  { id: "comments", selector: '[data-tour="doc-comments"]', titleKey: "docComments" },
];

function useTourSteps(): TourStep[] {
  const location = useLocation();
  const { user } = useAuthStore();
  const isAdmin = user?.roles.some((r) => ["super_admin", "admin"].includes(r)) ?? false;

  return useMemo(() => {
    const steps = MAIN_STEPS.filter((s) => s.id !== "admin" || isAdmin);
    if (location.pathname.includes("/spaces/") && location.pathname.includes("/docs/")) {
      return [...steps, ...DOC_STEPS];
    }
    return steps;
  }, [location.pathname, isAdmin]);
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
  const { isAuthenticated } = useAuthStore();
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

  useEffect(() => {
    if (!isAuthenticated || !shouldAutoStart()) return;
    const timer = window.setTimeout(() => start(), 800);
    return () => window.clearTimeout(timer);
  }, [isAuthenticated, shouldAutoStart, start]);

  useLayoutEffect(() => {
    if (!active || !current) {
      setRect(null);
      return;
    }
    const el = document.querySelector(current.selector);
    if (!el) {
      setRect(null);
      return;
    }
    const update = () => setRect(el.getBoundingClientRect());
    update();
    el.scrollIntoView({ block: "nearest", behavior: "smooth" });
    window.addEventListener("resize", update);
    window.addEventListener("scroll", update, true);
    return () => {
      window.removeEventListener("resize", update);
      window.removeEventListener("scroll", update, true);
    };
  }, [active, current, step]);

  if (!active || !current) return null;

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
      <div className="product-tour__backdrop" onClick={skip} aria-hidden />
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
      <div className="product-tour__card" style={cardStyle}>
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
