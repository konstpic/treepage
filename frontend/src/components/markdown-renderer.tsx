import { useMemo } from "react";
import ReactMarkdown from "react-markdown";
import { Link } from "react-router-dom";
import remarkGfm from "remark-gfm";
import rehypeHighlight from "rehype-highlight";
import {
  buildLinkIndex,
  buildRelationsSection,
  prepareMarkdown,
  type LinkDoc,
  type ParsedDocMeta,
} from "@/lib/wiki-markdown";
import { MermaidDiagram } from "@/components/mermaid-diagram";

function DocProperties({ meta }: { meta: ParsedDocMeta }) {
  const entries = Object.entries(meta.properties);
  if (!meta.status && entries.length === 0) return null;

  return (
    <div className="mb-6 flex flex-wrap gap-2 border-b border-default pb-4">
      {meta.status && (
        <span className="badge badge-success">status: {meta.status}</span>
      )}
      {entries.map(([k, v]) => {
        if (k.toLowerCase() === "status") return null;
        const short = v.length > 48 ? `${v.slice(0, 48)}…` : v;
        const isUrl = /^https?:\/\//i.test(v);
        return (
          <span key={k} className="badge badge-neutral" title={`${k}: ${v}`}>
            {k}: {isUrl ? (
              <a href={v} target="_blank" rel="noreferrer" className="text-primary hover:underline">
                link
              </a>
            ) : (
              short
            )}
          </span>
        );
      })}
    </div>
  );
}

interface MarkdownRendererProps {
  content: string;
  spaceSlug?: string;
  documents?: LinkDoc[];
  docPath?: string;
}

export function MarkdownRenderer({
  content,
  spaceSlug,
  documents = [],
  docPath,
}: MarkdownRendererProps) {
  const prepared = useMemo(() => {
    if (!spaceSlug || documents.length === 0) {
      return { body: content, meta: { properties: {}, relations: [] } as ParsedDocMeta };
    }
    const index = buildLinkIndex(documents);
    const result = prepareMarkdown(content, { spaceSlug, documents, docPath });
    const relationsBlock = buildRelationsSection(result.meta.relations, index, spaceSlug);
    return {
      body: result.body + relationsBlock,
      meta: result.meta,
    };
  }, [content, spaceSlug, documents, docPath]);

  return (
    <div className="prose-docs">
      <DocProperties meta={prepared.meta} />
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        rehypePlugins={[rehypeHighlight]}
        components={{
          a({ href, children, ...props }) {
            if (href?.startsWith("/spaces/")) {
              return (
                <Link to={href} className="text-primary hover:text-primary-hover hover:underline" {...props}>
                  {children}
                </Link>
              );
            }
            return (
              <a
                href={href}
                target={href?.startsWith("http") ? "_blank" : undefined}
                rel={href?.startsWith("http") ? "noreferrer" : undefined}
                className="text-primary hover:text-primary-hover hover:underline"
                {...props}
              >
                {children}
              </a>
            );
          },
          code({ className, children, ...props }) {
            const match = /language-(\w+)/.exec(className || "");
            const lang = match?.[1];
            const text = String(children).replace(/\n$/, "");
            if (lang === "mermaid") {
              return <MermaidDiagram chart={text} />;
            }
            return (
              <code className={className} {...props}>
                {children}
              </code>
            );
          },
          pre({ children, node, ...props }) {
            const codeEl = node?.children?.[0];
            if (codeEl?.type === "element") {
              const rawClass = codeEl.properties?.className;
              const classStr = Array.isArray(rawClass) ? rawClass.join(" ") : String(rawClass ?? "");
              if (classStr.includes("language-mermaid")) {
                return <>{children}</>;
              }
            }
            return <pre {...props}>{children}</pre>;
          },
        }}
      >
        {prepared.body}
      </ReactMarkdown>
    </div>
  );
}
