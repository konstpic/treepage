const AUTocomplete_KEY = "treepage_editor_autocomplete";

export function isEditorAutocompleteEnabled(): boolean {
  if (typeof localStorage === "undefined") return true;
  const v = localStorage.getItem(AUTocomplete_KEY);
  return v === null ? true : v === "true";
}

export function setEditorAutocompleteEnabled(enabled: boolean): void {
  localStorage.setItem(AUTocomplete_KEY, String(enabled));
}
