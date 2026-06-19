import { useEffect, useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { ChevronRight, FileText, Folder, FolderOpen } from "lucide-react";
import {
  ancestorFolderPaths,
  buildDocTree,
  countDocs,
  type DocItem,
  type DocTreeNode,
} from "@/lib/doc-tree";
import { cn } from "@/lib/utils";
import { useI18n } from "@/lib/i18n";

interface DocumentTreeProps {
  spaceSlug: string;
  documents: DocItem[];
  activeSlug?: string;
  activePath?: string;
  className?: string;
}

function TreeNode({
  node,
  spaceSlug,
  depth,
  activeSlug,
  expanded,
  onToggle,
}: {
  node: DocTreeNode;
  spaceSlug: string;
  depth: number;
  activeSlug?: string;
  expanded: Set<string>;
  onToggle: (folderPath: string) => void;
}) {
  if (node.type === "file") {
    const active = node.doc.slug === activeSlug;
    return (
      <Link
        to={`/spaces/${spaceSlug}/docs/${node.doc.slug}`}
        className={cn(
          "flex items-center gap-2 rounded-lg py-1.5 pr-2 text-sm transition-colors",
          active
            ? "surface-active"
            : "text-muted hover:bg-surface-hover hover:text-fg-secondary"
        )}
        style={{ paddingLeft: `${depth * 12 + 8}px` }}
        title={node.doc.path}
      >
        <FileText className="h-3.5 w-3.5 shrink-0 opacity-70" />
        <span className="truncate">{node.doc.title}</span>
      </Link>
    );
  }

  const open = expanded.has(node.path);
  const Icon = open ? FolderOpen : Folder;

  return (
    <div>
      <button
        type="button"
        onClick={() => onToggle(node.path)}
        className="flex w-full items-center gap-1.5 rounded-lg py-1.5 pr-2 text-left text-sm font-medium text-fg-secondary transition-colors hover:bg-surface-hover"
        style={{ paddingLeft: `${depth * 12 + 4}px` }}
      >
        <ChevronRight
          className={cn("h-3.5 w-3.5 shrink-0 text-subtle transition-transform", open && "rotate-90")}
        />
        <Icon className="h-3.5 w-3.5 shrink-0 text-primary/80" />
        <span className="truncate">{node.name}</span>
      </button>
      {open && (
        <div>
          {node.children.map((child) => (
            <TreeNode
              key={child.type === "folder" ? `f:${child.path}` : `d:${child.doc.id}`}
              node={child}
              spaceSlug={spaceSlug}
              depth={depth + 1}
              activeSlug={activeSlug}
              expanded={expanded}
              onToggle={onToggle}
            />
          ))}
        </div>
      )}
    </div>
  );
}

export function DocumentTree({
  spaceSlug,
  documents,
  activeSlug,
  activePath,
  className,
}: DocumentTreeProps) {
  const { t, pagesCount } = useI18n();
  const tree = useMemo(() => buildDocTree(documents), [documents]);
  const total = useMemo(() => countDocs(tree), [tree]);

  const [expanded, setExpanded] = useState<Set<string>>(() => new Set());

  useEffect(() => {
    const next = new Set<string>();
    // Expand top-level folders by default
    for (const node of tree) {
      if (node.type === "folder") next.add(node.path);
    }
    // Expand path to active document
    if (activePath) {
      for (const p of ancestorFolderPaths(activePath)) next.add(p);
    }
    setExpanded(next);
  }, [tree, activePath]);

  function toggle(folderPath: string) {
    setExpanded((prev) => {
      const next = new Set(prev);
      if (next.has(folderPath)) next.delete(folderPath);
      else next.add(folderPath);
      return next;
    });
  }

  if (documents.length === 0) {
    return (
      <p className={cn("text-sm text-subtle", className)}>{t("space.noDocuments")}</p>
    );
  }

  return (
    <nav className={cn("space-y-0.5", className)} aria-label={t("space.docTreeAria")}>
      <p className="mb-3 px-2 text-xs font-medium uppercase tracking-wide text-subtle">
        {pagesCount(total)}
      </p>
      {tree.map((node) => (
        <TreeNode
          key={node.type === "folder" ? `f:${node.path}` : `d:${node.doc.id}`}
          node={node}
          spaceSlug={spaceSlug}
          depth={0}
          activeSlug={activeSlug}
          expanded={expanded}
          onToggle={toggle}
        />
      ))}
    </nav>
  );
}
