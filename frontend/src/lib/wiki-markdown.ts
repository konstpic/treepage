import type { DocItem } from "@/lib/doc-tree";

export type LinkDoc = Pick<DocItem, "id" | "slug" | "title" | "path">;

/** Matches backend syncer slugify. */
export function slugifyPath(path: string): string {
  const s = path
    .replace(/\.md$/i, "")
    .replace(/\\/g, "/")
    .toLowerCase()
    .trim()
    .replace(/\//g, "-");
  return s.replace(/[^a-z0-9]+/g, "-").replace(/^-+|-+$/g, "");
}

export interface LinkIndex {
  bySlug: Map<string, LinkDoc>;
  /** wiki ref → doc, e.g. api.callapi, api/callapi */
  byRef: Map<string, LinkDoc>;
  documents: LinkDoc[];
}

/** IcePanel refs (comagic.comagic-server) often differ from file names (comagic.widget.comagic-server.md). */
function addIcePanelAliases(path: string, byRef: Map<string, LinkDoc>, doc: LinkDoc) {
  const parts = path.replace(/\\/g, "/").toLowerCase().split("/").filter(Boolean);
  if (parts.length === 0) return;

  const filename = parts[parts.length - 1].replace(/\.md$/i, "");
  const product = parts[0];
  const isEntityFile = filename.includes(".");

  if (isEntityFile) {
    byRef.set(filename, doc);
  }

  const widgetIdx = parts.indexOf("widget");
  if (widgetIdx >= 1 && widgetIdx + 1 < parts.length - 1 && isEntityFile) {
    const service = parts[widgetIdx + 1];
    byRef.set(`${product}.${service}`, doc);
    byRef.set(`${product}.widget-${service}`, doc);
  }

  if (parts.length >= 3 && isEntityFile && filename.startsWith(`${product}.`)) {
    const serviceDir = parts[parts.length - 2];
    if (!serviceDir.includes(".")) {
      byRef.set(`${product}.${serviceDir}`, doc);
    }
  }
}

export function buildLinkIndex(documents: LinkDoc[]): LinkIndex {
  const bySlug = new Map<string, LinkDoc>();
  const byRef = new Map<string, LinkDoc>();

  for (const doc of documents) {
    bySlug.set(doc.slug, doc);
    const pathKey = doc.path.replace(/\.md$/i, "").replace(/\\/g, "/").toLowerCase();
    const refs = new Set<string>([
      pathKey,
      pathKey.replace(/\//g, "."),
      pathKey.replace(/\//g, "-"),
      doc.slug,
      pathKey.split("/").pop() || "",
    ]);
    for (const ref of refs) {
      if (ref) byRef.set(ref, doc);
    }
    addIcePanelAliases(doc.path, byRef, doc);
  }

  return { bySlug, byRef, documents };
}

export function resolveWikiTarget(target: string, index: LinkIndex): LinkDoc | undefined {
  const raw = target.trim().replace(/^\//, "").replace(/\.md$/i, "").toLowerCase();
  const candidates = [
    raw,
    raw.replace(/\./g, "/"),
    raw.replace(/\//g, "."),
    raw.replace(/-/g, "."),
    slugifyPath(raw),
    slugifyPath(raw.replace(/\./g, "/")),
  ];
  for (const key of candidates) {
    const doc = index.byRef.get(key) ?? index.bySlug.get(key);
    if (doc) return doc;
  }

  // Last segment match: comagic.comagic-server → *comagic-server entity file under comagic/
  const segments = raw.split(".");
  if (segments.length >= 2) {
    const product = segments[0];
    const tail = segments[segments.length - 1];
    for (const doc of index.documents) {
      const parts = doc.path.replace(/\\/g, "/").toLowerCase().split("/");
      const fn = parts[parts.length - 1]?.replace(/\.md$/i, "") || "";
      if (!parts[0]?.startsWith(product)) continue;
      if (fn.endsWith(`.${tail}`) || fn === tail) return doc;
      const widgetIdx = parts.indexOf("widget");
      if (widgetIdx >= 1 && parts[widgetIdx + 1] === tail) return doc;
    }
  }

  return undefined;
}

export interface ParsedDocMeta {
  status?: string;
  properties: Record<string, string>;
  relations: string[];
}

export interface PreparedMarkdown {
  body: string;
  meta: ParsedDocMeta;
}

const WIKI_LINK_RE = /\[\[([^\]|]+)(?:\|([^\]]+))?\]\]/g;
const WIKI_LINK_TEST = /\[\[([^\]|]+)(?:\|([^\]]+))?\]\]/;
const FRONTMATTER_RE = /^---\r?\n([\s\S]*?)\r?\n---\r?\n?/;

function parseSimpleFrontmatter(block: string): ParsedDocMeta {
  const meta: ParsedDocMeta = { properties: {}, relations: [] };
  let inHasPart = false;

  for (const line of block.split("\n")) {
    const trimmed = line.trim();
    if (!trimmed) {
      inHasPart = false;
      continue;
    }

    const wikiInQuotes = trimmed.match(/^"(\[\[[^\]]+\]\])"$/);
    if (wikiInQuotes || (inHasPart && WIKI_LINK_TEST.test(trimmed))) {
      inHasPart = true;
      const wikis = trimmed.match(WIKI_LINK_RE);
      if (wikis) {
        for (const m of trimmed.matchAll(WIKI_LINK_RE)) meta.relations.push(m[1].trim());
      }
      continue;
    }

    if (/^haspart\s*:/i.test(trimmed)) {
      inHasPart = true;
      continue;
    }

    const kv = trimmed.match(/^([\w.-]+)\s*:\s*(.*)$/);
    if (kv) {
      inHasPart = false;
      const key = kv[1].toLowerCase();
      const val = kv[2].trim();
      if (key === "status") meta.status = val;
      else meta.properties[kv[1]] = val;
      continue;
    }
  }

  return meta;
}

function stripLeadingMetaBlock(content: string): { rest: string; meta: ParsedDocMeta } {
  const lines = content.split("\n");
  const meta: ParsedDocMeta = { properties: {}, relations: [] };
  let i = 0;

  while (i < lines.length) {
    const line = lines[i];
    const trimmed = line.trim();

    if (!trimmed) {
      i++;
      continue;
    }
    if (trimmed.startsWith("#")) break;

    const isMetaLine =
      /^[\w.-]+\s*:/.test(trimmed) ||
      /^"(\[\[[^\]]+\]\])"$/.test(trimmed) ||
      /^-\s*"(\[\[[^\]]+\]\])"$/.test(trimmed);

    if (!isMetaLine) break;

    const slice = lines.slice(i).join("\n");
    const parsed = parseSimpleFrontmatter(slice);
    Object.assign(meta.properties, parsed.properties);
    if (parsed.status) meta.status = parsed.status;
    meta.relations.push(...parsed.relations);

    while (i < lines.length) {
      const t = lines[i].trim();
      if (!t) {
        i++;
        continue;
      }
      if (t.startsWith("#")) break;
      if (
        /^[\w.-]+\s*:/.test(t) ||
        /^"(\[\[[^\]]+\]\])"$/.test(t) ||
        /^-\s*"(\[\[[^\]]+\]\])"$/.test(t)
      ) {
        i++;
        continue;
      }
      break;
    }
    break;
  }

  return { rest: lines.slice(i).join("\n").trimStart(), meta };
}

function wikiToMarkdown(
  target: string,
  label: string | undefined,
  index: LinkIndex,
  spaceSlug: string
): string {
  const doc = resolveWikiTarget(target, index);
  const text = label?.trim() || target.trim();
  if (doc) {
    return `[${text}](/spaces/${spaceSlug}/docs/${doc.slug})`;
  }
  return `[${text}](#wiki-${slugifyPath(target)})`;
}

function replaceWikiLinks(
  text: string,
  index: LinkIndex,
  spaceSlug: string
): string {
  return text.replace(WIKI_LINK_RE, (_, target: string, label?: string) =>
    wikiToMarkdown(target, label, index, spaceSlug)
  );
}

function replaceMarkdownFileLinks(
  text: string,
  index: LinkIndex,
  spaceSlug: string,
  currentPath: string
): string {
  return text.replace(/\[([^\]]+)\]\(([^)]+)\)/g, (full, linkText: string, href: string) => {
    if (/^(https?:|mailto:|#)/i.test(href)) return full;
    let resolved = href.split("#")[0].trim();
    if (resolved.startsWith("./")) resolved = resolved.slice(2);
    if (!resolved.endsWith(".md") && !resolved.includes(".")) resolved += ".md";

    const baseDir = currentPath.replace(/\\/g, "/").split("/").slice(0, -1);
    const parts = resolved.replace(/^\//, "").split("/");
    const stack = [...baseDir];
    for (const p of parts) {
      if (p === "..") stack.pop();
      else if (p !== ".") stack.push(p);
    }
    const pathKey = stack.join("/").replace(/\.md$/i, "").toLowerCase();
    const doc =
      index.byRef.get(pathKey) ??
      index.byRef.get(pathKey.replace(/\//g, ".")) ??
      index.bySlug.get(slugifyPath(pathKey));

    if (doc) return `[${linkText}](/spaces/${spaceSlug}/docs/${doc.slug})`;
    return full;
  });
}

export function prepareMarkdown(
  raw: string,
  opts: { spaceSlug: string; documents: LinkDoc[]; docPath?: string }
): PreparedMarkdown {
  const index = buildLinkIndex(opts.documents);
  let content = raw.replace(/\r\n/g, "\n");
  let meta: ParsedDocMeta = { properties: {}, relations: [] };

  const fm = content.match(FRONTMATTER_RE);
  if (fm) {
    const parsed = parseSimpleFrontmatter(fm[1]);
    meta = {
      status: parsed.status,
      properties: { ...parsed.properties },
      relations: [...parsed.relations],
    };
    content = content.slice(fm[0].length);
  }

  const stripped = stripLeadingMetaBlock(content);
  content = stripped.rest;
  meta.status = meta.status || stripped.meta.status;
  meta.properties = { ...stripped.meta.properties, ...meta.properties };
  meta.relations = [...new Set([...meta.relations, ...stripped.meta.relations])];

  content = replaceWikiLinks(content, index, opts.spaceSlug);
  if (opts.docPath) {
    content = replaceMarkdownFileLinks(content, index, opts.spaceSlug, opts.docPath);
  }

  return { body: content.trim(), meta };
}

export function buildRelationsSection(
  relations: string[],
  index: LinkIndex,
  spaceSlug: string
): string {
  const items = relations
    .map((target) => {
      const doc = resolveWikiTarget(target, index);
      if (doc) return `- [${doc.title}](/spaces/${spaceSlug}/docs/${doc.slug})`;
      return `- ${target} *(link not found)*`;
    })
    .join("\n");
  return items ? `\n\n## Related\n\n${items}` : "";
}
