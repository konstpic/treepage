import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { BrowserRouter } from "react-router-dom";
import { Providers } from "@/components/providers";
import { AppBootstrap } from "@/components/app-bootstrap";
import { AppRoutes } from "@/App";
import "@/styles/index.css";

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <BrowserRouter>
      <Providers>
        <AppBootstrap>
          <div className="flex min-h-dvh flex-col">
            <AppRoutes />
          </div>
        </AppBootstrap>
      </Providers>
    </BrowserRouter>
  </StrictMode>
);
