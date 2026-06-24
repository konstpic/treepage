export type SpaceDocChrome = {
  documentId: string | null;
  showComments: boolean;
};

export type SpaceDocOutletContext = {
  setDocChrome: (chrome: SpaceDocChrome) => void;
};

export const defaultSpaceDocChrome: SpaceDocChrome = {
  documentId: null,
  showComments: false,
};
