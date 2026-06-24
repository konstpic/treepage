import { useCallback, useEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";
import { Maximize2, X, ZoomIn, ZoomOut } from "lucide-react";
import mermaid from "mermaid";
import { useI18n } from "@/lib/i18n";

function isLightTheme() {
  const t = document.documentElement.dataset.theme;
  return t === "fox_white" || t === "light";
}

function initMermaid() {
  const light = isLightTheme();
  mermaid.initialize({
    startOnLoad: false,
    securityLevel: "loose",
    theme: light ? "neutral" : "dark",
    themeVariables: light
      ? {
          primaryTextColor: "#0f172a",
          secondaryTextColor: "#1e293b",
          tertiaryTextColor: "#0f172a",
          textColor: "#0f172a",
          primaryColor: "#f1f5f9",
          primaryBorderColor: "#64748b",
          lineColor: "#475569",
          mainBkg: "#e2e8f0",
          nodeBorder: "#64748b",
          clusterBkg: "#f8fafc",
          titleColor: "#0f172a",
          edgeLabelBackground: "#f8fafc",
        }
      : {
          primaryTextColor: "#f8fafc",
          secondaryTextColor: "#f1f5f9",
          tertiaryTextColor: "#f8fafc",
          textColor: "#f8fafc",
          primaryColor: "#334155",
          primaryBorderColor: "#94a3b8",
          lineColor: "#cbd5e1",
          mainBkg: "#475569",
          nodeBorder: "#94a3b8",
          clusterBkg: "#1e293b",
          titleColor: "#f8fafc",
          edgeLabelBackground: "#1e293b",
        },
  });
}

interface MermaidDiagramProps {
  chart: string;
}

export function MermaidDiagram({ chart }: MermaidDiagramProps) {
  const { t } = useI18n();
  const ref = useRef<HTMLDivElement>(null);
  const modalRef = useRef<HTMLDivElement>(null);
  const [error, setError] = useState("");
  const [svg, setSvg] = useState("");
  const [expanded, setExpanded] = useState(false);
  const [zoom, setZoom] = useState(1);

  const renderChart = useCallback(async () => {
    initMermaid();
    const id = `mermaid-${Math.random().toString(36).slice(2)}`;
    try {
      const { svg: rendered } = await mermaid.render(id, chart);
      setSvg(rendered);
      setError("");
    } catch (e) {
      setSvg("");
      setError(e instanceof Error ? e.message : "Syntax error in diagram");
    }
  }, [chart]);

  useEffect(() => {
    renderChart();
  }, [renderChart]);

  useEffect(() => {
    if (ref.current) {
      ref.current.innerHTML = svg;
    }
  }, [svg]);

  useEffect(() => {
    if (expanded && modalRef.current) {
      modalRef.current.innerHTML = svg;
    }
  }, [expanded, svg]);

  useEffect(() => {
    if (!expanded) setZoom(1);
  }, [expanded]);

  useEffect(() => {
    if (!expanded) return;
    const prev = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") setExpanded(false);
    };
    window.addEventListener("keydown", onKey);
    return () => {
      document.body.style.overflow = prev;
      window.removeEventListener("keydown", onKey);
    };
  }, [expanded]);

  const modal =
    expanded && svg ? (
      <div
        className="fixed inset-0 z-[200] flex flex-col bg-black/75 p-4 backdrop-blur-sm sm:p-6"
        role="dialog"
        aria-modal="true"
        aria-label={t("mermaid.diagram")}
        onClick={() => setExpanded(false)}
      >
        <div
          className="mx-auto flex h-full w-full max-w-[min(100%,100rem)] flex-col overflow-hidden rounded-2xl border border-default bg-surface shadow-2xl"
          onClick={(e) => e.stopPropagation()}
        >
          <div className="flex shrink-0 items-center justify-between gap-2 border-b border-default px-4 py-3">
            <span className="text-sm font-medium text-fg">{t("mermaid.diagram")}</span>
            <div className="flex items-center gap-1">
              <button
                type="button"
                className="btn-ghost !py-1 !px-2"
                onClick={() => setZoom((z) => Math.max(0.5, z - 0.25))}
                aria-label={t("mermaid.zoomOut")}
              >
                <ZoomOut className="h-4 w-4" />
              </button>
              <span className="min-w-[3rem] text-center text-xs text-subtle">{Math.round(zoom * 100)}%</span>
              <button
                type="button"
                className="btn-ghost !py-1 !px-2"
                onClick={() => setZoom((z) => Math.min(3, z + 0.25))}
                aria-label={t("mermaid.zoomIn")}
              >
                <ZoomIn className="h-4 w-4" />
              </button>
              <button
                type="button"
                className="btn-ghost !py-1 !px-2"
                onClick={() => setExpanded(false)}
                aria-label={t("mermaid.close")}
              >
                <X className="h-4 w-4" />
              </button>
            </div>
          </div>
          <div className="min-h-0 flex-1 overflow-auto p-4 sm:p-6">
            <div
              ref={modalRef}
              className="mermaid-canvas mermaid-canvas-expanded origin-top-left"
              style={{ transform: `scale(${zoom})` }}
            />
          </div>
        </div>
      </div>
    ) : null;

  return (
    <>
      <div className="mermaid-wrap my-6 not-prose">
        <div className="mb-2 flex items-center justify-end gap-2">
          {svg && (
            <button
              type="button"
              className="btn-ghost !py-1 !px-2 text-xs"
              onClick={() => setExpanded(true)}
              title={t("mermaid.expand")}
            >
              <Maximize2 className="h-4 w-4" />
              {t("mermaid.expand")}
            </button>
          )}
        </div>
        {error ? (
          <div className="rounded-xl border border-danger/30 bg-surface-muted p-4 text-sm">
            <p className="text-danger-soft">
              {t("mermaid.renderError")}: {error}
            </p>
            <details className="mt-3">
              <summary className="cursor-pointer text-subtle">{t("mermaid.sourceCode")}</summary>
              <pre className="mt-2 overflow-x-auto rounded-lg bg-surface p-3 text-xs text-muted">{chart}</pre>
            </details>
          </div>
        ) : (
          <div ref={ref} className="mermaid-canvas overflow-x-auto rounded-xl border border-default bg-surface-muted p-4" />
        )}
      </div>

      {modal && createPortal(modal, document.body)}
    </>
  );
}
