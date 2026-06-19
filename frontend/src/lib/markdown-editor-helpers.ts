import type { EditorView } from "@codemirror/view";

export interface SelectionRange {
  from: number;
  to: number;
  text: string;
}

export function getSelection(view: EditorView): SelectionRange {
  const { from, to } = view.state.selection.main;
  return { from, to, text: view.state.sliceDoc(from, to) };
}

export function replaceRange(view: EditorView, from: number, to: number, insert: string) {
  view.dispatch({
    changes: { from, to, insert },
    selection: { anchor: from + insert.length },
  });
  view.focus();
}

export function wrapSelection(view: EditorView, before: string, after: string) {
  const { from, to, text } = getSelection(view);
  if (from === to) {
    replaceRange(view, from, to, `${before}text${after}`);
    view.dispatch({
      selection: { anchor: from + before.length, head: from + before.length + 4 },
    });
    return;
  }
  replaceRange(view, from, to, `${before}${text}${after}`);
}

export function insertLinePrefix(view: EditorView, prefix: string) {
  const { from, to } = view.state.selection.main;
  const doc = view.state.doc;
  const startLine = doc.lineAt(from).number;
  const endLine = doc.lineAt(to).number;
  const changes: { from: number; insert: string }[] = [];
  for (let n = startLine; n <= endLine; n++) {
    const line = doc.line(n);
    changes.push({ from: line.from, insert: prefix });
  }
  view.dispatch({ changes });
  view.focus();
}

export function insertBlock(view: EditorView, block: string) {
  const { from, to } = getSelection(view);
  const doc = view.state.doc;
  const line = doc.lineAt(from);
  const needsLeadingNewline = line.from > 0 && doc.sliceString(line.from - 1, line.from) !== "\n";
  const insert = `${needsLeadingNewline ? "\n" : ""}${block}\n`;
  replaceRange(view, to, to, insert);
}

export const MARKDOWN_SNIPPETS = {
  table: `| Column 1 | Column 2 |
| -------- | -------- |
| Cell 1   | Cell 2   |
`,
  mermaidFlow: `\`\`\`mermaid
flowchart TD
  A[Start] --> B{Decision}
  B -->|Yes| C[Action]
  B -->|No| D[End]
\`\`\`
`,
  tags: `tags: tag1, tag2

`,
  wikiLink: "[[page-slug|Link text]]",
  codeFence: `\`\`\`
code
\`\`\`
`,
} as const;
