import { useEffect, useRef } from "react";
import type { EditorView } from "@codemirror/view";
import { useI18n } from "@/lib/i18n";
import {
  getSelection,
  insertBlock,
  insertLinePrefix,
  MARKDOWN_SNIPPETS,
  wrapSelection,
} from "@/lib/markdown-editor-helpers";

export interface ContextMenuState {
  x: number;
  y: number;
  from: number;
  to: number;
}

interface EditorContextMenuProps {
  view: EditorView | null;
  menu: ContextMenuState | null;
  onClose: () => void;
}

export function EditorContextMenu({ view, menu, onClose }: EditorContextMenuProps) {
  const { t } = useI18n();
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!menu) return;
    function onPointerDown(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node)) onClose();
    }
    function onKey(e: KeyboardEvent) {
      if (e.key === "Escape") onClose();
    }
    window.addEventListener("mousedown", onPointerDown);
    window.addEventListener("keydown", onKey);
    return () => {
      window.removeEventListener("mousedown", onPointerDown);
      window.removeEventListener("keydown", onKey);
    };
  }, [menu, onClose]);

  if (!menu || !view) return null;

  const { text } = getSelection(view);

  const items = [
    {
      label: t("documentEditor.bold"),
      action: () => wrapSelection(view, "**", "**"),
      disabled: false,
    },
    {
      label: t("documentEditor.italic"),
      action: () => wrapSelection(view, "*", "*"),
      disabled: false,
    },
    {
      label: t("documentEditor.link"),
      action: () => wrapSelection(view, "[", "](url)"),
      disabled: false,
    },
    {
      label: t("documentEditor.quote"),
      action: () => insertLinePrefix(view, "> "),
      disabled: false,
    },
    {
      label: t("documentEditor.wikiLink"),
      action: () =>
        text
          ? wrapSelection(view, "[[", "]]")
          : wrapSelection(view, "[[", "|link text]]"),
      disabled: false,
    },
    {
      label: t("documentEditor.insertTable"),
      action: () => insertBlock(view, MARKDOWN_SNIPPETS.table),
      disabled: false,
    },
    {
      label: t("documentEditor.insertMermaid"),
      action: () => insertBlock(view, MARKDOWN_SNIPPETS.mermaidFlow),
      disabled: false,
    },
    {
      label: t("documentEditor.insertTags"),
      action: () => insertBlock(view, MARKDOWN_SNIPPETS.tags),
      disabled: false,
    },
  ];

  return (
    <div
      ref={ref}
      className="editor-context-menu"
      style={{ top: menu.y, left: menu.x }}
      role="menu"
    >
      {items.map((item) => (
        <button
          key={item.label}
          type="button"
          role="menuitem"
          className="editor-context-menu__item"
          disabled={item.disabled}
          onClick={() => {
            item.action();
            onClose();
          }}
        >
          {item.label}
        </button>
      ))}
    </div>
  );
}

export function openEditorContextMenu(
  view: EditorView,
  clientX: number,
  clientY: number,
): ContextMenuState {
  const { from, to } = view.state.selection.main;
  return { x: clientX, y: clientY, from, to };
}
