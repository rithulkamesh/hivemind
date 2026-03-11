import express from "express";
import path from "path";
import { fileURLToPath } from "url";
import { toNodeHandler } from "better-auth/node";
import { auth } from "./auth.js";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const app = express();
app.use(express.json());

// Set password for OAuth-only users (server-only; session from cookies).
app.post("/auth/set-password", async (req, res) => {
  const newPassword = req.body?.newPassword;
  if (typeof newPassword !== "string" || newPassword.length < 12) {
    res.status(400).json({ error: "newPassword required (min 12 characters)" });
    return;
  }
  try {
    await auth.api.setPassword({
      body: { newPassword },
      headers: req.headers as HeadersInit,
    });
    res.status(200).json({ success: true });
  } catch (err: unknown) {
    const message = err instanceof Error ? err.message : "Failed to set password";
    res.status(400).json({ error: message });
  }
});

app.all("/auth/*", toNodeHandler(auth));
app.use(express.static(path.join(__dirname, "../dist")));
app.get("*", (_req, res) => {
  res.sendFile(path.join(__dirname, "../dist/index.html"));
});

const port = Number(process.env.PORT ?? 3000);
app.listen(port, () => {
  console.log(`Registry web + auth listening on :${port}`);
});
