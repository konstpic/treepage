import { useCallback, useMemo, useState } from "react";
import CodeMirror from "@uiw/react-codemirror";
import { markdown, markdownLanguage } from "@codemirror/lang-markdown";
import { EditorView } from "@codemirror/view";
import type { LinkDoc } from "@/lib/wiki-markdown";
import {
  isEditorAutocompleteEnabled,
  setEditorAutocompleteEnabled,
} from "@/lib/editor-preferences";
import { markdownAutocompleteExtension } from "@/lib/markdown-autocomplete";
import { EditorToolbar } from "@/components/editor-toolbar";
import {
  EditorContextMenu,
  openEditorContextMenu,
  type ContextMenuState,
} from "@/components/editor-context-menu";
import { useI18n } from "@/lib/i18n";

interface MarkdownEditorProps {
  value: string;
  onChange: (value: string) => void;
  documents?: LinkDoc[];
  ariaLabel?: string;
}

export function MarkdownEditor({
  value,
  onChange,
  documents = [],
  ariaLabel,
}: MarkdownEditorProps) {
  const { t } = useI18n();
  const [view, setView] = useState<EditorView | null>(null);
  const [contextMenu, setContextMenu] = useState<ContextMenuState | null>(null);
  const [autocompleteEnabled, setAutocompleteEnabled] = useState(isEditorAutocompleteEnabled);

  const handleAutocompleteToggle = useCallback((enabled: boolean) => {
    setAutocompleteEnabled(enabled);
    setEditorAutocompleteEnabled(enabled);
  }, []);

  const extensions = useMemo(
    () => [
      markdown({ base: markdownLanguage, codeLanguages: [] }),
      EditorView.lineWrapping,
      ...markdownAutocompleteExtension(documents, autocompleteEnabled),
      EditorView.domEventHandlers({
        contextmenu: (event, v) => {
          event.preventDefault();
          setContextMenu(openEditorContextMenu(v, event.clientX, event.clientY));
          return true;
        },
      }),
    ],
    [documents, autocompleteEnabled],
  );

  return (
    <div className="markdown-editor relative">
      <EditorToolbar
        view={view}
        autocompleteEnabled={autocompleteEnabled}
        onAutocompleteToggle={handleAutocompleteToggle}
      />
      <CodeMirror
        value={value}
        height="28rem"
        className="markdown-editor__codemirror"
        basicSetup={{
          lineNumbers: true,
          foldGutter: false,
          highlightActiveLine: true,
          autocompletion: false,
        }}
        extensions={extensions}
        onChange={onChange}
        onCreateEditor={setView}
        aria-label={ariaLabel ?? t("document.contentLabel")}
      />
      <EditorContextMenu view={view} menu={contextMenu} onClose={() => setContextMenu(null)} />
      {autocompleteEnabled && (
        <p className="mt-1 text-xs text-subtle">{t("documentEditor.autocompleteHint")}</p>
      )}
    </div>
  );
}
