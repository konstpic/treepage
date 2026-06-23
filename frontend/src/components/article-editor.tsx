import { useCallback, useRef } from "react";
import {
  Bold,
  Heading2,
  Italic,
  Link as LinkIcon,
  List,
  ListOrdered,
  Quote,
} from "lucide-react";
import { useI18n } from "@/lib/i18n";

interface ArticleEditorProps {
  value: string;
  onChange: (value: string) => void;
  ariaLabel?: string;
}

function wrapSelection(text: string, start: number, end: number, before: string, after = before) {
  const selected = text.slice(start, end);
  const wrapped = before + (selected || "text") + after;
  const next = text.slice(0, start) + wrapped + text.slice(end);
  const cursor = start + before.length + (selected || "text").length;
  return { next, cursor: selected ? end + before.length + after.length : cursor };
}

function prefixLines(text: string, start: number, end: number, prefix: string) {
  const blockStart = text.lastIndexOf("\n", start - 1) + 1;
  const blockEnd = text.indexOf("\n", end);
  const sliceEnd = blockEnd === -1 ? text.length : blockEnd;
  const block = text.slice(blockStart, sliceEnd);
  const lines = block.split("\n").map((line) => (line.startsWith(prefix) ? line : prefix + line));
  const next = text.slice(0, blockStart) + lines.join("\n") + text.slice(sliceEnd);
  return { next, cursor: blockStart + lines.join("\n").length };
}

export function ArticleEditor({ value, onChange, ariaLabel }: ArticleEditorProps) {
  const { t } = useI18n();
  const ref = useRef<HTMLTextAreaElement>(null);

  const apply = useCallback(
    (action: "bold" | "italic" | "h2" | "quote" | "ul" | "ol" | "link") => {
      const el = ref.current;
      if (!el) return;
      const start = el.selectionStart;
      const end = el.selectionEnd;
      let result = { next: value, cursor: end };

      switch (action) {
        case "bold":
          result = wrapSelection(value, start, end, "**");
          break;
        case "italic":
          result = wrapSelection(value, start, end, "*");
          break;
        case "h2":
          result = prefixLines(value, start, end, "## ");
          break;
        case "quote":
          result = prefixLines(value, start, end, "> ");
          break;
        case "ul":
          result = prefixLines(value, start, end, "- ");
          break;
        case "ol":
          result = prefixLines(value, start, end, "1. ");
          break;
        case "link": {
          const selected = value.slice(start, end) || "link text";
          const url = window.prompt("URL", "https://");
          if (!url) return;
          result = {
            next: value.slice(0, start) + `[${selected}](${url})` + value.slice(end),
            cursor: start + `[${selected}](${url})`.length,
          };
          break;
        }
      }

      onChange(result.next);
      requestAnimationFrame(() => {
        el.setSelectionRange(result.cursor, result.cursor);
        el.focus();
      });
    },
    [value, onChange],
  );

  const buttons = [
    { action: "bold" as const, icon: Bold, label: t("documentEditor.bold") },
    { action: "italic" as const, icon: Italic, label: t("documentEditor.italic") },
    { action: "h2" as const, icon: Heading2, label: t("documentEditor.heading2") },
    { action: "quote" as const, icon: Quote, label: t("documentEditor.quote") },
    { action: "ul" as const, icon: List, label: t("documentEditor.bulletList") },
    { action: "ol" as const, icon: ListOrdered, label: t("documentEditor.orderedList") },
    { action: "link" as const, icon: LinkIcon, label: t("documentEditor.link") },
  ];

  return (
    <div className="article-editor rounded-xl border border-default bg-surface">
      <div className="flex flex-wrap gap-1 border-b border-default p-2">
        {buttons.map(({ action, icon: Icon, label }) => (
          <button
            key={action}
            type="button"
            className="btn-ghost !px-2 !py-1"
            title={label}
            aria-label={label}
            onClick={() => apply(action)}
          >
            <Icon className="h-4 w-4" />
          </button>
        ))}
      </div>
      <textarea
        ref={ref}
        className="min-h-[28rem] w-full resize-y border-0 bg-transparent px-4 py-3 text-base leading-relaxed text-fg outline-none"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        aria-label={ariaLabel ?? t("document.contentLabel")}
      />
      <p className="border-t border-default px-4 py-2 text-xs text-subtle">{t("documentEditor.formatArticleHint")}</p>
    </div>
  );
}
