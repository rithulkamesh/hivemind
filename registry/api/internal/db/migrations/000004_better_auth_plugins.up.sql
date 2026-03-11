-- Better Auth plugin schema: twoFactor (2FA) plugin.
-- See https://www.better-auth.com/docs/plugins/2fa (Schema section)

ALTER TABLE "user" ADD COLUMN IF NOT EXISTS "twoFactorEnabled" BOOLEAN DEFAULT false;

CREATE TABLE IF NOT EXISTS "twoFactor" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "userId" TEXT NOT NULL REFERENCES "user"("id") ON DELETE CASCADE,
    "secret" TEXT,
    "backupCodes" TEXT
);

CREATE INDEX IF NOT EXISTS idx_twoFactor_userId ON "twoFactor"("userId");
