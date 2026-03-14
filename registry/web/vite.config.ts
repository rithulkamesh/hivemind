import path from "path";
import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";
import { toNodeHandler } from "better-auth/node";

export default defineConfig({
  plugins: [
    react(),
    {
      name: "better-auth",
      configureServer(server) {
        // Handle /auth/token (GET) to return JWT for Go API
        server.middlewares.use("/auth/token", async (req, res, next) => {
          if (req.method !== "GET") return next();
          try {
            console.log("[Vite] Handling /auth/token request");
            // Import auth dynamically to avoid build-time issues
            const { auth } = await import("./server/auth");
            const session = await auth.api.getSession({ headers: req.headers as HeadersInit });
            
            if (!session) {
              console.log("[Vite] No session found");
              res.writeHead(401, { "Content-Type": "application/json" });
              res.end(JSON.stringify({ error: "Unauthorized" }));
              return;
            }

            console.log("[Vite] Session found for user:", session.user.email);

            // Generate a JWT compatible with the backend JWKS verifier
            // The Go backend expects: sub (user ID), email, name, scopes
            // We use the `jwt` plugin's `signJWT` method if available.

            let token = session.session.token; // Fallback

            // @ts-ignore
            if (auth.api.signJWT) {
                 try {
                     const scopes = [];
                     // @ts-ignore
                     if (session.user.role === "admin") {
                         scopes.push("admin");
                     }
                     
                     // Construct payload matching Go's JWKSClaims
                     const payload = {
                         sub: session.user.id,
                         email: session.user.email,
                         name: session.user.name,
                         scopes: scopes,
                         // Standard claims
                         iat: Math.floor(Date.now() / 1000),
                         exp: Math.floor(Date.now() / 1000) + (60 * 60), // 1 hour
                     };

                     console.log("[Vite] Signing JWT with payload:", JSON.stringify(payload));
                     // @ts-ignore
                     const result = await auth.api.signJWT({ body: { payload } });
                     // @ts-ignore
                     token = result.token;
                     
                     console.log("[Vite] JWT signed successfully");
                 } catch (err) {
                     console.error("[Vite] Error signing JWT:", err);
                     // Fallback to session token (likely won't work for backend but keeps flow alive)
                 }
            } else {
                 console.log("[Vite] auth.api.signJWT method NOT found");
            }
            
            res.writeHead(200, { "Content-Type": "application/json" });
            res.end(JSON.stringify({ token }));
          } catch (e) {
            console.error("[Vite] Error in /auth/token:", e);
             next(e);
          }
        });

        // Handle set-password first (must run before catch-all /auth handler).
        server.middlewares.use("/auth/set-password", (req, res, next) => {
          if (req.method !== "POST") return next();
          // Log that we hit this
          console.log("Handling /auth/set-password");
          let body = "";
          req.on("data", (chunk) => { body += chunk; });
          req.on("end", async () => {
            try {
              if (!body) {
                 // Try to read from req.body if parsed?
                 // But in vite dev server, it's raw.
              }
              const data = JSON.parse(body || "{}");
              // ...
              const { auth } = await import("./server/auth");
              await auth.api.setPassword({ body: { newPassword: data.newPassword }, headers: req.headers as HeadersInit });
              res.writeHead(200, { "Content-Type": "application/json" });
              res.end(JSON.stringify({ success: true }));
            } catch (err) {
              console.error("set-password error", err);
              res.writeHead(400, { "Content-Type": "application/json" });
              res.end(JSON.stringify({ error: err instanceof Error ? err.message : "Failed to set password" }));
            }
          });
        });
        server.middlewares.use(async (req, res, next) => {
          if (!req.url?.startsWith("/auth")) return next();
          const { auth } = await import("./server/auth");
          const handler = toNodeHandler(auth);
          handler(req, res);
        });
      },
    },
  ],
  resolve: {
    alias: { "@": path.resolve(__dirname, "./src") },
  },
  base: "/",
  build: {
    outDir: "dist",
    target: "es2020",
  },
  server: {
    host: "0.0.0.0",
    port: 3000,
    strictPort: true,
    proxy: {
      "/api": {
        target: process.env.VITE_PROXY_TARGET ?? "http://localhost:8080",
        changeOrigin: true,
        secure: false,
        timeout: 30000,
        ws: true,
      },
      "/simple": {
        target: process.env.VITE_PROXY_TARGET ?? "http://localhost:8080",
        changeOrigin: true,
        secure: false,
        timeout: 30000,
        ws: true,
      },
    },
  },
});
