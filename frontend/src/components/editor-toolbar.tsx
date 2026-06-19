import type { EditorView } from "@codemirror/view";
import {
  Bold,
  Code,
  GitBranch,
  Heading1,
  Heading2,
  Italic,
  Link,
  List,
  ListOrdered,
  Minus,
  Quote,
  Sparkles,
  Strikethrough,
  Table,
  Tag,
  Underline,
  Workflow,
} from "lucide-react";
import { useI18n } from "@/lib/i18n";
import {
  insertBlock,
  insertLinePrefix,
  MARKDOWN_SNIPPETS,
  wrapSelection,
} from "@/lib/markdown-editor-helpers";

interface EditorToolbarProps {
  view: EditorView | null;
  autocompleteEnabled: boolean;
  onAutocompleteToggle: (enabled: boolean) => void;
}

type ToolAction = {
  id: string;
  icon: React.ReactNode;
  label: string;
  action: (view: EditorView) => void;
};

export function EditorToolbar({
  view,
  autocompleteEnabled,
  onAutocompleteToggle,
}: EditorToolbarProps) {
  const { t } = useI18n();

  const tools: ToolAction[] = [
    {
      id: "bold",
      icon: <Bold className="h-4 w-4" />,
      label: t("documentEditor.bold"),
      action: (v) => wrapSelection(v, "**", "**"),
    },
    {
      id: "italic",
      icon: <Italic className="h-4 w-4" />,
      label: t("documentEditor.italic"),
      action: (v) => wrapSelection(v, "*", "*"),
    },
    {
      id: "underline",
      icon: <Underline className="h-4 w-4" />,
      label: t("documentEditor.underline"),
      action: (v) => wrapSelection(v, "<u>", "</u>"),
    },
    {
      id: "strike",
      icon: <Strikethrough className="h-4 w-4" />,
      label: t("documentEditor.strike"),
      action: (v) => wrapSelection(v, "~~", "~~"),
    },
    {
      id: "h1",
      icon: <Heading1 className="h-4 w-4" />,
      label: t("documentEditor.heading1"),
      action: (v) => insertLinePrefix(v, "# "),
    },
    {
      id: "h2",
      icon: <Heading2 className="h-4 w-4" />,
      label: t("documentEditor.heading2"),
      action: (v) => insertLinePrefix(v, "## "),
    },
    {
      id: "quote",
      icon: <Quote className="h-4 w-4" />,
      label: t("documentEditor.quote"),
      action: (v) => insertLinePrefix(v, "> "),
    },
    {
      id: "ul",
      icon: <List className="h-4 w-4" />,
      label: t("documentEditor.bulletList"),
      action: (v) => insertLinePrefix(v, "- "),
    },
    {
      id: "ol",
      icon: <ListOrdered className="h-4 w-4" />,
      label: t("documentEditor.orderedList"),
      action: (v) => insertLinePrefix(v, "1. "),
    },
    {
      id: "code",
      icon: <Code className="h-4 w-4" />,
      label: t("documentEditor.code"),
      action: (v) => wrapSelection(v, "`", "`"),
    },
    {
      id: "codeblock",
      icon: <Minus className="h-4 w-4" />,
      label: t("documentEditor.codeBlock"),
      action: (v) => insertBlock(v, MARKDOWN_SNIPPETS.codeFence),
    },
    {
      id: "link",
      icon: <Link className="h-4 w-4" />,
      label: t("documentEditor.link"),
      action: (v) => wrapSelection(v, "[", "](url)"),
    },
    {
      id: "table",
      icon: <Table className="h-4 w-4" />,
      label: t("documentEditor.table"),
      action: (v) => insertBlock(v, MARKDOWN_SNIPPETS.table),
    },
    {
      id: "tags",
      icon: <Tag className="h-4 w-4" />,
      label: t("documentEditor.tags"),
      action: (v) => insertBlock(v, MARKDOWN_SNIPPETS.tags),
    },
    {
      id: "mermaid",
      icon: <Workflow className="h-4 w-4" />,
      label: t("documentEditor.mermaid"),
      action: (v) => insertBlock(v, MARKDOWN_SNIPPETS.mermaidFlow),
    },
    {
      id: "wiki",
      icon: <GitBranch className="h-4 w-4" />,
      label: t("documentEditor.wikiLink"),
      action: (v) => wrapSelection(v, "[[", "]]"),
    },
  ];

  return (
    <div className="editor-toolbar flex flex-wrap items-center gap-0.5 rounded-t-xl border border-b-0 border-default bg-surface-muted px-2 py-1.5">
      {tools.map((tool) => (
        <button
          key={tool.id}
          type="button"
          className="editor-toolbar__btn"
          title={tool.label}
          aria-label={tool.label}
          disabled={!view}
          onMouseDown={(e) => e.preventDefault()}
          onClick={() => view && tool.action(view)}
        >
          {tool.icon}
        </button>
      ))}
      <div className="ml-auto flex items-center gap-2 pl-2">
        <label className="flex cursor-pointer items-center gap-1.5 text-xs text-muted">
          <Sparkles className="h-3.5 w-3.5" />
          <span>{t("documentEditor.autocomplete")}</span>
          <input
            type="checkbox"
            className="h-3.5 w-3.5 rounded border-default accent-primary"
            checked={autocompleteEnabled}
            onChange={(e) => onAutocompleteToggle(e.target.checked)}
          />
        </label>
      </div>
    </div>
  );
}
