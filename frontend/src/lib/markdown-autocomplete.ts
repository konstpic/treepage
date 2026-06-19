import {
  autocompletion,
  completionKeymap,
  type Completion,
  type CompletionContext,
} from "@codemirror/autocomplete";
import { keymap } from "@codemirror/view";
import type { LinkDoc } from "@/lib/wiki-markdown";

const HEADING_LEVELS = ["# ", "## ", "### ", "#### "];
const TAG_SUGGESTIONS = ["tags: ", "tag1, tag2"];
const MERMAID_START = "```mermaid\nflowchart TD\n  A --> B\n```";

function wikiCompletions(documents: LinkDoc[]): Completion[] {
  return documents.flatMap((doc) => [
    {
      label: `[[${doc.slug}|${doc.title}]]`,
      type: "link",
      detail: doc.path,
    },
    {
      label: `[[${doc.path.replace(/\.md$/i, "")}|${doc.title}]]`,
      type: "link",
      detail: doc.path,
    },
  ]);
}

function markdownCompletionSource(documents: LinkDoc[]) {
  const wiki = wikiCompletions(documents);

  return (context: CompletionContext) => {
    const word = context.matchBefore(/\[\[[\w./-]*|#{1,4}\s*[\w-]*|tags:\s*[\w, ]*$/);
    if (!word && !context.explicit) return null;

    const line = context.state.doc.lineAt(context.pos);
    const lineText = line.text.slice(0, context.pos - line.from);

    if (lineText.match(/\[\[[\w./-]*$/)) {
      return { from: word ? word.from : context.pos, options: wiki, validFor: /^[\[\w./|-]*$/ };
    }

    if (lineText.match(/^#{1,4}\s*[\w-]*$/)) {
      return {
        from: line.from,
        options: HEADING_LEVELS.map((h) => ({ label: h.trim(), type: "keyword", apply: h })),
      };
    }

    if (lineText.match(/^tags:\s*[\w, ]*$/i)) {
      return {
        from: word?.from ?? context.pos,
        options: TAG_SUGGESTIONS.map((t) => ({ label: t, type: "property" })),
      };
    }

    if (context.explicit) {
      return {
        from: context.pos,
        options: [
          { label: MERMAID_START, type: "keyword", detail: "Mermaid flowchart" },
          { label: "[[wiki-link]]", type: "link" },
          { label: "tags: ", type: "property" },
          ...HEADING_LEVELS.map((h) => ({ label: h, type: "keyword" })),
        ],
      };
    }

    return null;
  };
}

export function markdownAutocompleteExtension(documents: LinkDoc[], enabled: boolean) {
  if (!enabled) return [];
  return [
    autocompletion({ override: [markdownCompletionSource(documents)] }),
    keymap.of(completionKeymap),
  ];
}
