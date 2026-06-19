import { Routes, Route } from "react-router-dom";
import { Navigation } from "@/components/navigation";
import { Footer } from "@/components/footer";
import { AppShell } from "@/components/app-shell";
import { HomePage } from "@/pages/home";
import { AuthPage, AuthCallbackPage } from "@/pages/auth";
import { SpacesPage } from "@/pages/spaces";
import { SpaceIndexPage } from "@/pages/space";
import { SpaceBooksPage } from "@/pages/books";
import { BookReaderPage } from "@/pages/book";
import { DocumentPage } from "@/pages/document";
import { SpaceDocLayout } from "@/components/space-doc-layout";
import { SearchPage } from "@/pages/search";
import {
  AdminLayout,
  AdminIndexRedirect,
  AdminSpacesPage,
  AdminRepositoriesPage,
  AdminUsersPage,
  AdminGroupsPage,
  AdminSettingsPage,
  AdminOIDCPage,
  AdminAuditPage,
} from "@/pages/admin";

export function AppRoutes() {
  return (
    <AppShell
      header={<Navigation />}
      main={
        <Routes>
          <Route path="/" element={<HomePage />} />
          <Route path="/auth" element={<AuthPage />} />
          <Route path="/auth/callback" element={<AuthCallbackPage />} />
          <Route path="/spaces" element={<SpacesPage />} />
          <Route path="/spaces/:slug" element={<SpaceDocLayout />}>
            <Route index element={<SpaceIndexPage />} />
            <Route path="books" element={<SpaceBooksPage />} />
            <Route path="books/:bookSlug" element={<BookReaderPage />} />
            <Route path="docs/:docSlug" element={<DocumentPage />} />
          </Route>
          <Route path="/search" element={<SearchPage />} />
          <Route path="/admin" element={<AdminLayout />}>
            <Route index element={<AdminIndexRedirect />} />
            <Route path="spaces" element={<AdminSpacesPage />} />
            <Route path="repositories" element={<AdminRepositoriesPage />} />
            <Route path="users" element={<AdminUsersPage />} />
            <Route path="groups" element={<AdminGroupsPage />} />
            <Route path="settings" element={<AdminSettingsPage />} />
            <Route path="oidc" element={<AdminOIDCPage />} />
            <Route path="audit" element={<AdminAuditPage />} />
          </Route>
        </Routes>
      }
      footer={<Footer />}
    />
  );
}
