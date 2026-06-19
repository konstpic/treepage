export interface DocItem {
  id: string;
  slug: string;
  title: string;
  path: string;
  updated_at: string;
}

export interface DocFolderNode {
  type: "folder";
  name: string;
  /** Full folder path from repo root, e.g. architecture/services */
  path: string;
  children: DocTreeNode[];
}

export interface DocFileNode {
  type: "file";
  doc: DocItem;
}

export type DocTreeNode = DocFolderNode | DocFileNode;

function sortNodes(nodes: DocTreeNode[]): DocTreeNode[] {
  return [...nodes].sort((a, b) => {
    if (a.type !== b.type) return a.type === "folder" ? -1 : 1;
    const nameA = a.type === "folder" ? a.name : a.doc.title;
    const nameB = b.type === "folder" ? b.name : b.doc.title;
    return nameA.localeCompare(nameB, undefined, { sensitivity: "base" });
  });
}

export function buildDocTree(docs: DocItem[]): DocTreeNode[] {
  const root: DocFolderNode = { type: "folder", name: "", path: "", children: [] };

  for (const doc of docs) {
    const normalized = doc.path.replace(/\\/g, "/");
    const parts = normalized.split("/").filter(Boolean);
    if (parts.length === 0) continue;

    let current = root;
    for (let i = 0; i < parts.length - 1; i++) {
      const segment = parts[i];
      const folderPath = parts.slice(0, i + 1).join("/");
      let folder = current.children.find(
        (n): n is DocFolderNode => n.type === "folder" && n.name === segment
      );
      if (!folder) {
        folder = { type: "folder", name: segment, path: folderPath, children: [] };
        current.children.push(folder);
      }
      current = folder;
    }

    current.children.push({ type: "file", doc });
  }

  function finalize(node: DocFolderNode): DocTreeNode[] {
    node.children = sortNodes(node.children.map((child) => {
      if (child.type === "folder") {
        return { ...child, children: finalize(child) };
      }
      return child;
    }));
    return node.children;
  }

  return finalize(root);
}

/** Folder paths that contain the given document path (for auto-expand). */
export function ancestorFolderPaths(docPath: string): string[] {
  const parts = docPath.replace(/\\/g, "/").split("/").filter(Boolean);
  if (parts.length <= 1) return [];
  const paths: string[] = [];
  for (let i = 1; i < parts.length; i++) {
    paths.push(parts.slice(0, i).join("/"));
  }
  return paths;
}

export function countDocs(nodes: DocTreeNode[]): number {
  let n = 0;
  for (const node of nodes) {
    if (node.type === "file") n++;
    else n += countDocs(node.children);
  }
  return n;
}
