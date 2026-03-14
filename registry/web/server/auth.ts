import { betterAuth } from "better-auth";
import { jwt, twoFactor, oneTap, username, admin, organization, haveIBeenPwned, lastLoginMethod, bearer } from "better-auth/plugins";
import { passkey } from "@better-auth/passkey";
import { apiKey } from "@better-auth/api-key";
import { Pool } from "pg";

const baseURL = process.env.BETTER_AUTH_URL ?? process.env.FRONTEND_URL ?? "http://localhost:3000";

export const auth = betterAuth({
  basePath: "/auth",
  baseURL,
  appName: "Hivemind Registry",
  database: new Pool({
    connectionString: process.env.DATABASE_URL,
  }),

  emailAndPassword: {
    enabled: true,
    requireEmailVerification: true,
  },

  emailVerification: {
    sendVerificationEmail: async ({ user, url }) => {
      const internalUrl = process.env.GO_API_INTERNAL_URL ?? "http://localhost:8080";
      const secret = process.env.INTERNAL_SECRET ?? "";
      await fetch(`${internalUrl}/internal/email/verify`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "X-Internal-Secret": secret,
        },
        body: JSON.stringify({ email: user.email, url }),
      });
    },
  },

  socialProviders: {
    github: {
      clientId: process.env.GITHUB_CLIENT_ID ?? "",
      clientSecret: process.env.GITHUB_CLIENT_SECRET ?? "",
    },
    google: {
      clientId: process.env.GOOGLE_CLIENT_ID ?? "",
      clientSecret: process.env.GOOGLE_CLIENT_SECRET ?? "",
    },
  },

  plugins: [
    bearer(),
    jwt(),
    twoFactor({ issuer: "Hivemind Registry" }),
    oneTap({ clientId: process.env.GOOGLE_CLIENT_ID ?? "" }),
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
  ].filter((o, i, a) => a.indexOf(o) === i),

});
