import { useEffect } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { motion } from "framer-motion";
import { Loader2, ShieldCheck, AlertCircle } from "lucide-react";
import { ApiError, getLoginUrl, loginLocal } from "@/lib/api";
import { useAuthStore } from "@/lib/store";
import { useI18n } from "@/lib/i18n";
import { useAuthFormSplash } from "@/hooks/use-auth-splash";
import { AuthScatterPiece } from "@/components/app-shell";
import { useSplashStore, SCATTER_ANIM_MS } from "@/lib/splash-store";
import { useState } from "react";

const DEV_LOGIN = import.meta.env.VITE_DEV_LOGIN === "true" || import.meta.env.DEV;

const scatterEase: [number, number, number, number] = [0.55, 0, 0.15, 1];

export function AuthPage() {
  const navigate = useNavigate();
  const { t } = useI18n();
  const { startAuthFormSplash } = useAuthFormSplash();
  const scattering = useSplashStore((s) => s.phase === "scatter");
  const setAuth = useAuthStore((s) => s.setAuth);
  const setUser = useAuthStore((s) => s.setUser);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [email, setEmail] = useState("admin@local");
  const [password, setPassword] = useState("admin");

  async function handleOIDCLogin() {
    setLoading(true);
    setError("");
    try {
      window.location.href = await getLoginUrl();
    } catch {
      setError(t("auth.oidcUnavailable"));
      setLoading(false);
    }
  }

  function handleDevLogin(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    startAuthFormSplash(async () => {
      setLoading(true);
      try {
        const data = await loginLocal(email, password);
        setAuth(data.access_token, data.refresh_token);
        if (data.user) setUser(data.user);
        navigate("/spaces", { replace: true });
      } catch (err) {
        setError(err instanceof ApiError ? err.message : t("auth.loginFailed"));
      } finally {
        setLoading(false);
      }
    });
  }

  return (
    <div className="mx-auto flex min-h-[60vh] max-w-md items-center px-4 py-16">
      <motion.div
        className="glass w-full p-8"
        animate={
          scattering
            ? { opacity: 0, scale: 0.94, filter: "blur(6px)" }
            : { opacity: 1, scale: 1, filter: "blur(0px)" }
        }
        transition={{ duration: SCATTER_ANIM_MS / 1000, ease: scatterEase }}
      >
        <AuthScatterPiece exit={{ x: "-75vw", y: "-45vh", rotate: -18, delay: 0 }}>
          <div className="mb-6 flex justify-center">
            <div className="flex h-14 w-14 items-center justify-center rounded-2xl bg-gradient-to-br from-brand-500 to-brand-700 shadow-lg shadow-brand-500/30">
              <ShieldCheck className="h-7 w-7 text-white" />
            </div>
          </div>
        </AuthScatterPiece>

        <AuthScatterPiece exit={{ x: "70vw", y: "-50vh", rotate: 12, delay: 0.04 }}>
          <h1 className="text-center text-2xl font-bold text-fg">{t("auth.title")}</h1>
        </AuthScatterPiece>

        {DEV_LOGIN ? (
          <>
            <AuthScatterPiece exit={{ x: "-55vw", y: "-35vh", rotate: -8, delay: 0.07 }}>
              <p className="mt-2 text-center text-sm text-muted">{t("auth.devHint")}</p>
            </AuthScatterPiece>

            <form onSubmit={handleDevLogin} className="mt-6 space-y-4">
              <AuthScatterPiece exit={{ x: "-95vw", y: "5vh", rotate: -10, delay: 0.1 }}>
                <input
                  className="input-field"
                  type="email"
                  autoComplete="username"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  placeholder={t("auth.email")}
                  required
                  disabled={scattering || loading}
                />
              </AuthScatterPiece>

              <AuthScatterPiece exit={{ x: "95vw", y: "0vh", rotate: 10, delay: 0.14 }}>
                <input
                  className="input-field"
                  type="password"
                  autoComplete="current-password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder={t("auth.password")}
                  required
                  disabled={scattering || loading}
                />
              </AuthScatterPiece>

              {error && (
                <AuthScatterPiece exit={{ x: "0vw", y: "-70vh", rotate: 6, delay: 0.12 }}>
                  <p className="flex items-center gap-2 text-sm text-danger-soft">
                    <AlertCircle className="h-4 w-4 shrink-0" />
                    {error}
                  </p>
                </AuthScatterPiece>
              )}

              <AuthScatterPiece exit={{ x: "0vw", y: "90vh", scale: 0.7, delay: 0.18 }}>
                <button type="submit" className="btn-primary w-full" disabled={loading || scattering}>
                  {loading ? <Loader2 className="h-4 w-4 animate-spin" /> : t("auth.signIn")}
                </button>
              </AuthScatterPiece>
            </form>
          </>
        ) : (
          <>
            <AuthScatterPiece exit={{ x: "-40vw", y: "-30vh", delay: 0.08 }}>
              <p className="mt-2 text-center text-sm text-muted">{t("auth.oidcHint")}</p>
            </AuthScatterPiece>
            <AuthScatterPiece exit={{ x: "0vw", y: "80vh", delay: 0.12 }}>
              <button
                type="button"
                onClick={handleOIDCLogin}
                className="btn-primary mt-8 w-full"
                disabled={loading}
              >
                {loading ? t("auth.redirecting") : t("auth.continueOidc")}
              </button>
            </AuthScatterPiece>
          </>
        )}

        <AuthScatterPiece exit={{ x: "-80vw", y: "75vh", rotate: -14, delay: 0.22 }} className="mt-3">
          <button
            type="button"
            onClick={() => navigate("/")}
            className="btn-ghost w-full"
            disabled={scattering}
          >
            {t("auth.backHome")}
          </button>
        </AuthScatterPiece>
      </motion.div>
    </div>
  );
}

export function AuthCallbackPage() {
  const [params] = useSearchParams();
  const setAuth = useAuthStore((s) => s.setAuth);
  const navigate = useNavigate();

  useEffect(() => {
    const access = params.get("access_token");
    const refresh = params.get("refresh_token");
    if (access && refresh) {
      setAuth(access, refresh);
      navigate("/spaces", { replace: true });
    }
  }, [params, setAuth, navigate]);

  return (
    <div className="flex min-h-[40vh] items-center justify-center">
      <Loader2 className="h-8 w-8 animate-spin text-primary" />
    </div>
  );
}
