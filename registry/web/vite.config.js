var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    function adopt(value) { return value instanceof P ? value : new P(function (resolve) { resolve(value); }); }
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
var __generator = (this && this.__generator) || function (thisArg, body) {
    var _ = { label: 0, sent: function() { if (t[0] & 1) throw t[1]; return t[1]; }, trys: [], ops: [] }, f, y, t, g = Object.create((typeof Iterator === "function" ? Iterator : Object).prototype);
    return g.next = verb(0), g["throw"] = verb(1), g["return"] = verb(2), typeof Symbol === "function" && (g[Symbol.iterator] = function() { return this; }), g;
    function verb(n) { return function (v) { return step([n, v]); }; }
    function step(op) {
        if (f) throw new TypeError("Generator is already executing.");
        while (g && (g = 0, op[0] && (_ = 0)), _) try {
            if (f = 1, y && (t = op[0] & 2 ? y["return"] : op[0] ? y["throw"] || ((t = y["return"]) && t.call(y), 0) : y.next) && !(t = t.call(y, op[1])).done) return t;
            if (y = 0, t) op = [op[0] & 2, t.value];
            switch (op[0]) {
                case 0: case 1: t = op; break;
                case 4: _.label++; return { value: op[1], done: false };
                case 5: _.label++; y = op[1]; op = [0]; continue;
                case 7: op = _.ops.pop(); _.trys.pop(); continue;
                default:
                    if (!(t = _.trys, t = t.length > 0 && t[t.length - 1]) && (op[0] === 6 || op[0] === 2)) { _ = 0; continue; }
                    if (op[0] === 3 && (!t || (op[1] > t[0] && op[1] < t[3]))) { _.label = op[1]; break; }
                    if (op[0] === 6 && _.label < t[1]) { _.label = t[1]; t = op; break; }
                    if (t && _.label < t[2]) { _.label = t[2]; _.ops.push(op); break; }
                    if (t[2]) _.ops.pop();
                    _.trys.pop(); continue;
            }
            op = body.call(thisArg, _);
        } catch (e) { op = [6, e]; y = 0; } finally { f = t = 0; }
        if (op[0] & 5) throw op[1]; return { value: op[0] ? op[1] : void 0, done: true };
    }
};
var _a, _b;
import path from "path";
import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";
import { toNodeHandler } from "better-auth/node";
export default defineConfig({
    plugins: [
        react(),
        {
            name: "better-auth",
            configureServer: function (server) {
                var _this = this;
                // Handle /auth/token (GET) to return JWT for Go API
                server.middlewares.use("/auth/token", function (req, res, next) { return __awaiter(_this, void 0, void 0, function () {
                    var auth, session, token, scopes, payload, result, err_1, e_1;
                    return __generator(this, function (_a) {
                        switch (_a.label) {
                            case 0:
                                if (req.method !== "GET")
                                    return [2 /*return*/, next()];
                                _a.label = 1;
                            case 1:
                                _a.trys.push([1, 10, , 11]);
                                console.log("[Vite] Handling /auth/token request");
                                return [4 /*yield*/, import("./server/auth")];
                            case 2:
                                auth = (_a.sent()).auth;
                                return [4 /*yield*/, auth.api.getSession({ headers: req.headers })];
                            case 3:
                                session = _a.sent();
                                if (!session) {
                                    console.log("[Vite] No session found");
                                    res.writeHead(401, { "Content-Type": "application/json" });
                                    res.end(JSON.stringify({ error: "Unauthorized" }));
                                    return [2 /*return*/];
                                }
                                console.log("[Vite] Session found for user:", session.user.email);
                                token = session.session.token;
                                if (!auth.api.signJWT) return [3 /*break*/, 8];
                                _a.label = 4;
                            case 4:
                                _a.trys.push([4, 6, , 7]);
                                scopes = [];
                                // @ts-ignore
                                if (session.user.role === "admin") {
                                    scopes.push("admin");
                                }
                                payload = {
                                    sub: session.user.id,
                                    email: session.user.email,
                                    name: session.user.name,
                                    scopes: scopes,
                                    // Standard claims
                                    iat: Math.floor(Date.now() / 1000),
                                    exp: Math.floor(Date.now() / 1000) + (60 * 60), // 1 hour
                                };
                                console.log("[Vite] Signing JWT with payload:", JSON.stringify(payload));
                                return [4 /*yield*/, auth.api.signJWT({ body: { payload: payload } })];
                            case 5:
                                result = _a.sent();
                                // @ts-ignore
                                token = result.token;
                                console.log("[Vite] JWT signed successfully");
                                return [3 /*break*/, 7];
                            case 6:
                                err_1 = _a.sent();
                                console.error("[Vite] Error signing JWT:", err_1);
                                return [3 /*break*/, 7];
                            case 7: return [3 /*break*/, 9];
                            case 8:
                                console.log("[Vite] auth.api.signJWT method NOT found");
                                _a.label = 9;
                            case 9:
                                res.writeHead(200, { "Content-Type": "application/json" });
                                res.end(JSON.stringify({ token: token }));
                                return [3 /*break*/, 11];
                            case 10:
                                e_1 = _a.sent();
                                console.error("[Vite] Error in /auth/token:", e_1);
                                next(e_1);
                                return [3 /*break*/, 11];
                            case 11: return [2 /*return*/];
                        }
                    });
                }); });
                // Handle set-password first (must run before catch-all /auth handler).
                server.middlewares.use("/auth/set-password", function (req, res, next) {
                    if (req.method !== "POST")
                        return next();
                    // Log that we hit this
                    console.log("Handling /auth/set-password");
                    var body = "";
                    req.on("data", function (chunk) { body += chunk; });
                    req.on("end", function () { return __awaiter(_this, void 0, void 0, function () {
                        var data, auth, err_2;
                        return __generator(this, function (_a) {
                            switch (_a.label) {
                                case 0:
                                    _a.trys.push([0, 3, , 4]);
                                    if (!body) {
                                        // Try to read from req.body if parsed?
                                        // But in vite dev server, it's raw.
                                    }
                                    data = JSON.parse(body || "{}");
                                    return [4 /*yield*/, import("./server/auth")];
                                case 1:
                                    auth = (_a.sent()).auth;
                                    return [4 /*yield*/, auth.api.setPassword({ body: { newPassword: data.newPassword }, headers: req.headers })];
                                case 2:
                                    _a.sent();
                                    res.writeHead(200, { "Content-Type": "application/json" });
                                    res.end(JSON.stringify({ success: true }));
                                    return [3 /*break*/, 4];
                                case 3:
                                    err_2 = _a.sent();
                                    console.error("set-password error", err_2);
                                    res.writeHead(400, { "Content-Type": "application/json" });
                                    res.end(JSON.stringify({ error: err_2 instanceof Error ? err_2.message : "Failed to set password" }));
                                    return [3 /*break*/, 4];
                                case 4: return [2 /*return*/];
                            }
                        });
                    }); });
                });
                server.middlewares.use(function (req, res, next) { return __awaiter(_this, void 0, void 0, function () {
                    var auth, handler;
                    var _a;
                    return __generator(this, function (_b) {
                        switch (_b.label) {
                            case 0:
                                if (!((_a = req.url) === null || _a === void 0 ? void 0 : _a.startsWith("/auth")))
                                    return [2 /*return*/, next()];
                                return [4 /*yield*/, import("./server/auth")];
                            case 1:
                                auth = (_b.sent()).auth;
                                handler = toNodeHandler(auth);
                                handler(req, res);
                                return [2 /*return*/];
                        }
                    });
                }); });
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
                target: (_a = process.env.VITE_PROXY_TARGET) !== null && _a !== void 0 ? _a : "http://localhost:8080",
                changeOrigin: true,
                secure: false,
                timeout: 30000,
                ws: true,
            },
            "/simple": {
                target: (_b = process.env.VITE_PROXY_TARGET) !== null && _b !== void 0 ? _b : "http://localhost:8080",
                changeOrigin: true,
                secure: false,
                timeout: 30000,
                ws: true,
            },
        },
    },
});
