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
var _a, _b, _c, _d, _e, _f, _g;
import { betterAuth } from "better-auth";
import { jwt, twoFactor, oneTap, username, admin, organization, haveIBeenPwned, lastLoginMethod, bearer } from "better-auth/plugins";
import { passkey } from "@better-auth/passkey";
import { apiKey } from "@better-auth/api-key";
import { Pool } from "pg";
var baseURL = (_b = (_a = process.env.BETTER_AUTH_URL) !== null && _a !== void 0 ? _a : process.env.FRONTEND_URL) !== null && _b !== void 0 ? _b : "http://localhost:3000";
export var auth = betterAuth({
    basePath: "/auth",
    baseURL: baseURL,
    appName: "Hivemind Registry",
    database: new Pool({
        connectionString: process.env.DATABASE_URL,
    }),
    emailAndPassword: {
        enabled: true,
        requireEmailVerification: true,
    },
    emailVerification: {
        sendVerificationEmail: function (_a) { return __awaiter(void 0, [_a], void 0, function (_b) {
            var internalUrl, secret;
            var _c, _d;
            var user = _b.user, url = _b.url;
            return __generator(this, function (_e) {
                switch (_e.label) {
                    case 0:
                        internalUrl = (_c = process.env.GO_API_INTERNAL_URL) !== null && _c !== void 0 ? _c : "http://localhost:8080";
                        secret = (_d = process.env.INTERNAL_SECRET) !== null && _d !== void 0 ? _d : "";
                        return [4 /*yield*/, fetch("".concat(internalUrl, "/internal/email/verify"), {
                                method: "POST",
                                headers: {
                                    "Content-Type": "application/json",
                                    "X-Internal-Secret": secret,
                                },
                                body: JSON.stringify({ email: user.email, url: url }),
                            })];
                    case 1:
                        _e.sent();
                        return [2 /*return*/];
                }
            });
        }); },
    },
    socialProviders: {
        github: {
            clientId: (_c = process.env.GITHUB_CLIENT_ID) !== null && _c !== void 0 ? _c : "",
            clientSecret: (_d = process.env.GITHUB_CLIENT_SECRET) !== null && _d !== void 0 ? _d : "",
        },
        google: {
            clientId: (_e = process.env.GOOGLE_CLIENT_ID) !== null && _e !== void 0 ? _e : "",
            clientSecret: (_f = process.env.GOOGLE_CLIENT_SECRET) !== null && _f !== void 0 ? _f : "",
        },
    },
    plugins: [
        bearer(),
        jwt(),
        twoFactor({ issuer: "Hivemind Registry" }),
        oneTap({ clientId: (_g = process.env.GOOGLE_CLIENT_ID) !== null && _g !== void 0 ? _g : "" }),
        passkey(),
        username(),
        admin(),
        organization(),
        haveIBeenPwned(),
        lastLoginMethod(),
        apiKey(),
    ],
    user: {
        modelName: "users",
        fields: {
            emailVerified: "email_verified",
            createdAt: "created_at",
            updatedAt: "updated_at",
        },
    },
    session: {
        expiresIn: 60 * 60 * 24 * 7, // 7 days
        cookieCache: {
            enabled: true,
            maxAge: 60 * 5,
        },
    },
    trustedOrigins: [
        baseURL,
        "http://localhost:3000",
        "http://127.0.0.1:3000",
    ].filter(function (o, i, a) { return a.indexOf(o) === i; }),
});
