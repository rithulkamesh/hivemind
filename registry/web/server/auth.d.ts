import { Pool } from "pg";
export declare const auth: import("better-auth").Auth<{
    basePath: string;
    baseURL: string;
    appName: string;
    database: Pool;
    emailAndPassword: {
        enabled: true;
        requireEmailVerification: true;
    };
    emailVerification: {
        sendVerificationEmail: ({ user, url }: {
            user: import("better-auth").User;
            url: string;
            token: string;
        }) => Promise<void>;
    };
    socialProviders: {
        github: {
            clientId: string;
            clientSecret: string;
        };
        google: {
            clientId: string;
            clientSecret: string;
        };
    };
    plugins: [{
        id: "bearer";
        hooks: {
            before: {
                matcher(context: import("better-auth").HookEndpointContext): boolean;
                handler: (inputContext: import("better-call").MiddlewareInputContext<import("better-call").MiddlewareOptions>) => Promise<{
                    context: {
                        headers: Headers;
                    };
                } | undefined>;
            }[];
            after: {
                matcher(context: import("better-auth").HookEndpointContext): true;
                handler: (inputContext: import("better-call").MiddlewareInputContext<import("better-call").MiddlewareOptions>) => Promise<void>;
            }[];
        };
        options: import("better-auth/plugins").BearerOptions | undefined;
    }, {
        id: "jwt";
        options: NoInfer<import("better-auth/plugins").JwtOptions>;
        endpoints: {
            getJwks: import("better-call").StrictEndpoint<string, {
                method: "GET";
                metadata: {
                    openapi: {
                        operationId: string;
                        description: string;
                        responses: {
                            "200": {
                                description: string;
                                content: {
                                    "application/json": {
                                        schema: {
                                            type: "object";
                                            properties: {
                                                keys: {
                                                    type: string;
                                                    description: string;
                                                    items: {
                                                        type: string;
                                                        properties: {
                                                            kid: {
                                                                type: string;
                                                                description: string;
                                                            };
                                                            kty: {
                                                                type: string;
                                                                description: string;
                                                            };
                                                            alg: {
                                                                type: string;
                                                                description: string;
                                                            };
                                                            use: {
                                                                type: string;
                                                                description: string;
                                                                enum: string[];
                                                                nullable: boolean;
                                                            };
                                                            n: {
                                                                type: string;
                                                                description: string;
                                                                nullable: boolean;
                                                            };
                                                            e: {
                                                                type: string;
                                                                description: string;
                                                                nullable: boolean;
                                                            };
                                                            crv: {
                                                                type: string;
                                                                description: string;
                                                                nullable: boolean;
                                                            };
                                                            x: {
                                                                type: string;
                                                                description: string;
                                                                nullable: boolean;
                                                            };
                                                            y: {
                                                                type: string;
                                                                description: string;
                                                                nullable: boolean;
                                                            };
                                                        };
                                                        required: string[];
                                                    };
                                                };
                                            };
                                            required: string[];
                                        };
                                    };
                                };
                            };
                        };
                    };
                };
            }, import("jose").JSONWebKeySet>;
            getToken: import("better-call").StrictEndpoint<"/token", {
                method: "GET";
                requireHeaders: true;
                use: ((inputContext: import("better-call").MiddlewareInputContext<import("better-call").MiddlewareOptions>) => Promise<{
                    session: {
                        session: Record<string, any> & {
                            id: string;
                            createdAt: Date;
                            updatedAt: Date;
                            userId: string;
                            expiresAt: Date;
                            token: string;
                            ipAddress?: string | null | undefined;
                            userAgent?: string | null | undefined;
                        };
                        user: Record<string, any> & {
                            id: string;
                            createdAt: Date;
                            updatedAt: Date;
                            email: string;
                            emailVerified: boolean;
                            name: string;
                            image?: string | null | undefined;
                        };
                    };
                }>)[];
                metadata: {
                    openapi: {
                        operationId: string;
                        description: string;
                        responses: {
                            200: {
                                description: string;
                                content: {
                                    "application/json": {
                                        schema: {
                                            type: "object";
                                            properties: {
                                                token: {
                                                    type: string;
                                                };
                                            };
                                        };
                                    };
                                };
                            };
                        };
                    };
                };
            }, {
                token: string;
            }>;
            signJWT: import("better-call").StrictEndpoint<string, {
                method: "POST";
                metadata: {
                    $Infer: {
                        body: {
                            payload: import("jose").JWTPayload;
                            overrideOptions?: import("better-auth/plugins").JwtOptions | undefined;
                        };
                    };
                };
                body: import("zod").ZodObject<{
                    payload: import("zod").ZodRecord<import("zod").ZodString, import("zod").ZodAny>;
                    overrideOptions: import("zod").ZodOptional<import("zod").ZodRecord<import("zod").ZodString, import("zod").ZodAny>>;
                }, import("better-auth").$strip>;
            }, {
                token: string;
            }>;
            verifyJWT: import("better-call").StrictEndpoint<string, {
                method: "POST";
                metadata: {
                    $Infer: {
                        body: {
                            token: string;
                            issuer?: string;
                        };
                        response: {
                            payload: {
                                sub: string;
                                aud: string;
                                [key: string]: any;
                            } | null;
                        };
                    };
                };
                body: import("zod").ZodObject<{
                    token: import("zod").ZodString;
                    issuer: import("zod").ZodOptional<import("zod").ZodString>;
                }, import("better-auth").$strip>;
            }, {
                payload: (import("jose").JWTPayload & Required<Pick<import("jose").JWTPayload, "sub" | "aud">>) | null;
            }>;
        };
        hooks: {
            after: {
                matcher(context: import("better-auth").HookEndpointContext): boolean;
                handler: (inputContext: import("better-call").MiddlewareInputContext<import("better-call").MiddlewareOptions>) => Promise<void>;
            }[];
        };
        schema: {
            jwks: {
                fields: {
                    publicKey: {
                        type: "string";
                        required: true;
                    };
                    privateKey: {
                        type: "string";
                        required: true;
                    };
                    createdAt: {
                        type: "date";
                        required: true;
                    };
                    expiresAt: {
                        type: "date";
                        required: false;
                    };
                };
            };
        };
    }, {
        id: "two-factor";
        endpoints: {
            enableTwoFactor: import("better-call").StrictEndpoint<"/two-factor/enable", {
                method: "POST";
                body: import("zod").ZodObject<{
                    password: import("zod").ZodString;
                    issuer: import("zod").ZodOptional<import("zod").ZodString>;
                }, import("better-auth").$strip>;
                use: ((inputContext: import("better-call").MiddlewareInputContext<import("better-call").MiddlewareOptions>) => Promise<{
                    session: {
                        session: Record<string, any> & {
                            id: string;
                            createdAt: Date;
                            updatedAt: Date;
                            userId: string;
                            expiresAt: Date;
                            token: string;
                            ipAddress?: string | null | undefined;
                            userAgent?: string | null | undefined;
                        };
                        user: Record<string, any> & {
                            id: string;
                            createdAt: Date;
                            updatedAt: Date;
                            email: string;
                            emailVerified: boolean;
                            name: string;
                            image?: string | null | undefined;
                        };
                    };
                }>)[];
                metadata: {
                    openapi: {
                        summary: string;
                        description: string;
                        responses: {
                            200: {
                                description: string;
                                content: {
                                    "application/json": {
                                        schema: {
                                            type: "object";
                                            properties: {
                                                totpURI: {
                                                    type: string;
                                                    description: string;
                                                };
                                                backupCodes: {
                                                    type: string;
                                                    items: {
                                                        type: string;
                                                    };
                                                    description: string;
                                                };
                                            };
                                        };
                                    };
                                };
                            };
                        };
                    };
                };
            }, {
                totpURI: string;
                backupCodes: string[];
            }>;
            disableTwoFactor: import("better-call").StrictEndpoint<"/two-factor/disable", {
                method: "POST";
                body: import("zod").ZodObject<{
                    password: import("zod").ZodString;
                }, import("better-auth").$strip>;
                use: ((inputContext: import("better-call").MiddlewareInputContext<import("better-call").MiddlewareOptions>) => Promise<{
                    session: {
                        session: Record<string, any> & {
                            id: string;
                            createdAt: Date;
                            updatedAt: Date;
                            userId: string;
                            expiresAt: Date;
                            token: string;
                            ipAddress?: string | null | undefined;
                            userAgent?: string | null | undefined;
                        };
                        user: Record<string, any> & {
                            id: string;
                            createdAt: Date;
                            updatedAt: Date;
                            email: string;
                            emailVerified: boolean;
                            name: string;
                            image?: string | null | undefined;
                        };
                    };
                }>)[];
                metadata: {
                    openapi: {
                        summary: string;
                        description: string;
                        responses: {
                            200: {
                                description: string;
                                content: {
                                    "application/json": {
                                        schema: {
                                            type: "object";
                                            properties: {
                                                status: {
                                                    type: string;
                                                };
                                            };
                                        };
                                    };
                                };
                            };
                        };
                    };
                };
            }, {
                status: boolean;
            }>;
            verifyBackupCode: import("better-call").StrictEndpoint<"/two-factor/verify-backup-code", {
                method: "POST";
                body: import("zod").ZodObject<{
                    code: import("zod").ZodString;
                    disableSession: import("zod").ZodOptional<import("zod").ZodBoolean>;
                    trustDevice: import("zod").ZodOptional<import("zod").ZodBoolean>;
                }, import("better-auth").$strip>;
                metadata: {
                    openapi: {
                        description: string;
                        responses: {
                            "200": {
                                description: string;
                                content: {
                                    "application/json": {
                                        schema: {
                                            type: "object";
                                            properties: {
                                                user: {
                                                    type: string;
                                                    properties: {
                                                        id: {
                                                            type: string;
                                                            description: string;
                                                        };
                                                        email: {
                                                            type: string;
                                                            format: string;
                                                            nullable: boolean;
                                                            description: string;
                                                        };
                                                        emailVerified: {
                                                            type: string;
                                                            nullable: boolean;
                                                            description: string;
                                                        };
                                                        name: {
                                                            type: string;
                                                            nullable: boolean;
                                                            description: string;
                                                        };
                                                        image: {
                                                            type: string;
                                                            format: string;
                                                            nullable: boolean;
                                                            description: string;
                                                        };
                                                        twoFactorEnabled: {
                                                            type: string;
                                                            description: string;
                                                        };
                                                        createdAt: {
                                                            type: string;
                                                            format: string;
                                                            description: string;
                                                        };
                                                        updatedAt: {
                                                            type: string;
                                                            format: string;
                                                            description: string;
                                                        };
                                                    };
                                                    required: string[];
                                                    description: string;
                                                };
                                                session: {
                                                    type: string;
                                                    properties: {
                                                        token: {
                                                            type: string;
                                                            description: string;
                                                        };
                                                        userId: {
                                                            type: string;
                                                            description: string;
                                                        };
                                                        createdAt: {
                                                            type: string;
                                                            format: string;
                                                            description: string;
                                                        };
                                                        expiresAt: {
                                                            type: string;
                                                            format: string;
                                                            description: string;
                                                        };
                                                    };
                                                    required: string[];
                                                    description: string;
                                                };
                                            };
                                            required: string[];
                                        };
                                    };
                                };
                            };
                        };
                    };
                };
            }, {
                token: string | undefined;
                user: (Record<string, any> & {
                    id: string;
                    createdAt: Date;
                    updatedAt: Date;
                    email: string;
                    emailVerified: boolean;
                    name: string;
                    image?: string | null | undefined;
                }) | import("better-auth/plugins").UserWithTwoFactor;
            }>;
            generateBackupCodes: import("better-call").StrictEndpoint<"/two-factor/generate-backup-codes", {
                method: "POST";
                body: import("zod").ZodObject<{
                    password: import("zod").ZodString;
                }, import("better-auth").$strip>;
                use: ((inputContext: import("better-call").MiddlewareInputContext<import("better-call").MiddlewareOptions>) => Promise<{
                    session: {
                        session: Record<string, any> & {
                            id: string;
                            createdAt: Date;
                            updatedAt: Date;
                            userId: string;
                            expiresAt: Date;
                            token: string;
                            ipAddress?: string | null | undefined;
                            userAgent?: string | null | undefined;
                        };
                        user: Record<string, any> & {
                            id: string;
                            createdAt: Date;
                            updatedAt: Date;
                            email: string;
                            emailVerified: boolean;
                            name: string;
                            image?: string | null | undefined;
                        };
                    };
                }>)[];
                metadata: {
                    openapi: {
                        description: string;
                        responses: {
                            "200": {
                                description: string;
                                content: {
                                    "application/json": {
                                        schema: {
                                            type: "object";
                                            properties: {
                                                status: {
                                                    type: string;
                                                    description: string;
                                                    enum: boolean[];
                                                };
                                                backupCodes: {
                                                    type: string;
                                                    items: {
                                                        type: string;
                                                    };
                                                    description: string;
                                                };
                                            };
                                            required: string[];
                                        };
                                    };
                                };
                            };
                        };
                    };
                };
            }, {
                status: boolean;
                backupCodes: string[];
            }>;
            viewBackupCodes: import("better-call").StrictEndpoint<string, {
                method: "POST";
                body: import("zod").ZodObject<{
                    userId: import("zod").ZodCoercedString<unknown>;
                }, import("better-auth").$strip>;
            }, {
                status: boolean;
                backupCodes: string[];
            }>;
            sendTwoFactorOTP: import("better-call").StrictEndpoint<"/two-factor/send-otp", {
                method: "POST";
                body: import("zod").ZodOptional<import("zod").ZodObject<{
                    trustDevice: import("zod").ZodOptional<import("zod").ZodBoolean>;
                }, import("better-auth").$strip>>;
                metadata: {
                    openapi: {
                        summary: string;
                        description: string;
                        responses: {
                            200: {
                                description: string;
                                content: {
                                    "application/json": {
                                        schema: {
                                            type: "object";
                                            properties: {
                                                status: {
                                                    type: string;
                                                };
                                            };
                                        };
                                    };
                                };
                            };
                        };
                    };
                };
            }, {
                status: boolean;
            }>;
            verifyTwoFactorOTP: import("better-call").StrictEndpoint<"/two-factor/verify-otp", {
                method: "POST";
                body: import("zod").ZodObject<{
                    code: import("zod").ZodString;
                    trustDevice: import("zod").ZodOptional<import("zod").ZodBoolean>;
                }, import("better-auth").$strip>;
                metadata: {
                    openapi: {
                        summary: string;
                        description: string;
                        responses: {
                            "200": {
                                description: string;
                                content: {
                                    "application/json": {
                                        schema: {
                                            type: "object";
                                            properties: {
                                                token: {
                                                    type: string;
                                                    description: string;
                                                };
                                                user: {
                                                    type: string;
                                                    properties: {
                                                        id: {
                                                            type: string;
                                                            description: string;
                                                        };
                                                        email: {
                                                            type: string;
                                                            format: string;
                                                            nullable: boolean;
                                                            description: string;
                                                        };
                                                        emailVerified: {
                                                            type: string;
                                                            nullable: boolean;
                                                            description: string;
                                                        };
                                                        name: {
                                                            type: string;
                                                            nullable: boolean;
                                                            description: string;
                                                        };
                                                        image: {
                                                            type: string;
                                                            format: string;
                                                            nullable: boolean;
                                                            description: string;
                                                        };
                                                        createdAt: {
                                                            type: string;
                                                            format: string;
                                                            description: string;
                                                        };
                                                        updatedAt: {
                                                            type: string;
                                                            format: string;
                                                            description: string;
                                                        };
                                                    };
                                                    required: string[];
                                                    description: string;
                                                };
                                            };
                                            required: string[];
                                        };
                                    };
                                };
                            };
                        };
                    };
                };
            }, {
                token: string;
                user: import("better-auth/plugins").UserWithTwoFactor;
            } | {
                token: string;
                user: Record<string, any> & {
                    id: string;
                    createdAt: Date;
                    updatedAt: Date;
                    email: string;
                    emailVerified: boolean;
                    name: string;
                    image?: string | null | undefined;
                };
            }>;
            generateTOTP: import("better-call").StrictEndpoint<string, {
                method: "POST";
                body: import("zod").ZodObject<{
                    secret: import("zod").ZodString;
                }, import("better-auth").$strip>;
                metadata: {
                    openapi: {
                        summary: string;
                        description: string;
                        responses: {
                            200: {
                                description: string;
                                content: {
                                    "application/json": {
                                        schema: {
                                            type: "object";
                                            properties: {
                                                code: {
                                                    type: string;
                                                };
                                            };
                                        };
                                    };
                                };
                            };
                        };
                    };
                };
            }, {
                code: string;
            }>;
            getTOTPURI: import("better-call").StrictEndpoint<"/two-factor/get-totp-uri", {
                method: "POST";
                use: ((inputContext: import("better-call").MiddlewareInputContext<import("better-call").MiddlewareOptions>) => Promise<{
                    session: {
                        session: Record<string, any> & {
                            id: string;
                            createdAt: Date;
                            updatedAt: Date;
                            userId: string;
                            expiresAt: Date;
                            token: string;
                            ipAddress?: string | null | undefined;
                            userAgent?: string | null | undefined;
                        };
                        user: Record<string, any> & {
                            id: string;
                            createdAt: Date;
                            updatedAt: Date;
                            email: string;
                            emailVerified: boolean;
                            name: string;
                            image?: string | null | undefined;
                        };
                    };
                }>)[];
                body: import("zod").ZodObject<{
                    password: import("zod").ZodString;
                }, import("better-auth").$strip>;
                metadata: {
                    openapi: {
                        summary: string;
                        description: string;
                        responses: {
                            200: {
                                description: string;
                                content: {
                                    "application/json": {
                                        schema: {
                                            type: "object";
                                            properties: {
                                                totpURI: {
                                                    type: string;
                                                };
                                            };
                                        };
                                    };
                                };
                            };
                        };
                    };
                };
            }, {
                totpURI: string;
            }>;
            verifyTOTP: import("better-call").StrictEndpoint<"/two-factor/verify-totp", {
                method: "POST";
                body: import("zod").ZodObject<{
                    code: import("zod").ZodString;
                    trustDevice: import("zod").ZodOptional<import("zod").ZodBoolean>;
                }, import("better-auth").$strip>;
                metadata: {
                    openapi: {
                        summary: string;
                        description: string;
                        responses: {
                            200: {
                                description: string;
                                content: {
                                    "application/json": {
                                        schema: {
                                            type: "object";
                                            properties: {
                                                status: {
                                                    type: string;
                                                };
                                            };
                                        };
                                    };
                                };
                            };
                        };
                    };
                };
            }, {
                token: string;
                user: import("better-auth/plugins").UserWithTwoFactor;
            } | {
                token: string;
                user: Record<string, any> & {
                    id: string;
                    createdAt: Date;
                    updatedAt: Date;
                    email: string;
                    emailVerified: boolean;
                    name: string;
                    image?: string | null | undefined;
                };
            }>;
        };
        options: NoInfer<{
            issuer: string;
        }>;
        hooks: {
            after: {
                matcher(context: import("better-auth").HookEndpointContext): boolean;
                handler: (inputContext: import("better-call").MiddlewareInputContext<import("better-call").MiddlewareOptions>) => Promise<{
                    twoFactorRedirect: boolean;
                } | undefined>;
            }[];
        };
        schema: {
            user: {
                fields: {
                    twoFactorEnabled: {
                        type: "boolean";
                        required: false;
                        defaultValue: false;
                        input: false;
                    };
                };
            };
            twoFactor: {
                fields: {
                    secret: {
                        type: "string";
                        required: true;
                        returned: false;
                        index: true;
                    };
                    backupCodes: {
                        type: "string";
                        required: true;
                        returned: false;
                    };
                    userId: {
                        type: "string";
                        required: true;
                        returned: false;
                        references: {
                            model: string;
                            field: string;
                        };
                        index: true;
                    };
                };
            };
        };
        rateLimit: {
            pathMatcher(path: string): boolean;
            window: number;
            max: number;
        }[];
        $ERROR_CODES: {
            OTP_NOT_ENABLED: import("better-auth").RawError<"OTP_NOT_ENABLED">;
            OTP_HAS_EXPIRED: import("better-auth").RawError<"OTP_HAS_EXPIRED">;
            TOTP_NOT_ENABLED: import("better-auth").RawError<"TOTP_NOT_ENABLED">;
            TWO_FACTOR_NOT_ENABLED: import("better-auth").RawError<"TWO_FACTOR_NOT_ENABLED">;
            BACKUP_CODES_NOT_ENABLED: import("better-auth").RawError<"BACKUP_CODES_NOT_ENABLED">;
            INVALID_BACKUP_CODE: import("better-auth").RawError<"INVALID_BACKUP_CODE">;
            INVALID_CODE: import("better-auth").RawError<"INVALID_CODE">;
            TOO_MANY_ATTEMPTS_REQUEST_NEW_CODE: import("better-auth").RawError<"TOO_MANY_ATTEMPTS_REQUEST_NEW_CODE">;
            INVALID_TWO_FACTOR_COOKIE: import("better-auth").RawError<"INVALID_TWO_FACTOR_COOKIE">;
        };
    }, {
        id: "one-tap";
        endpoints: {
            oneTapCallback: import("better-call").StrictEndpoint<"/one-tap/callback", {
                method: "POST";
                body: import("zod").ZodObject<{
                    idToken: import("zod").ZodString;
                }, import("better-auth").$strip>;
                metadata: {
                    openapi: {
                        summary: string;
                        description: string;
                        responses: {
                            200: {
                                description: string;
                                content: {
                                    "application/json": {
                                        schema: {
                                            type: "object";
                                            properties: {
                                                session: {
                                                    $ref: string;
                                                };
                                                user: {
                                                    $ref: string;
                                                };
                                            };
                                        };
                                    };
                                };
                            };
                            400: {
                                description: string;
                            };
                        };
                    };
                };
            }, {
                error: string;
            } | {
                token: string;
                user: {
                    id: string;
                    createdAt: Date;
                    updatedAt: Date;
                    email: string;
                    emailVerified: boolean;
                    name: string;
                    image?: string | null | undefined;
                };
            }>;
        };
        options: import("better-auth/plugins").OneTapOptions | undefined;
    }, {
        id: "passkey";
        endpoints: {
            generatePasskeyRegistrationOptions: import("better-call").StrictEndpoint<"/passkey/generate-register-options", {
                method: "GET";
                use: ((inputContext: import("better-call").MiddlewareInputContext<import("better-call").MiddlewareOptions>) => Promise<{
                    session: {
                        session: Record<string, any> & {
                            id: string;
                            createdAt: Date;
                            updatedAt: Date;
                            userId: string;
                            expiresAt: Date;
                            token: string;
                            ipAddress?: string | null | undefined;
                            userAgent?: string | null | undefined;
                        };
                        user: Record<string, any> & {
                            id: string;
                            createdAt: Date;
                            updatedAt: Date;
                            email: string;
                            emailVerified: boolean;
                            name: string;
                            image?: string | null | undefined;
                        };
                    };
                }>)[];
                query: import("zod").ZodOptional<import("zod").ZodObject<{
                    authenticatorAttachment: import("zod").ZodOptional<import("zod").ZodEnum<{
                        platform: "platform";
                        "cross-platform": "cross-platform";
                    }>>;
                    name: import("zod").ZodOptional<import("zod").ZodString>;
                }, import("better-auth").$strip>>;
                metadata: {
                    openapi: {
                        operationId: string;
                        description: string;
                        responses: {
                            200: {
                                description: string;
                                parameters: {
                                    query: {
                                        authenticatorAttachment: {
                                            description: string;
                                            required: boolean;
                                        };
                                        name: {
                                            description: string;
                                            required: boolean;
                                        };
                                    };
                                };
                                content: {
                                    "application/json": {
                                        schema: {
                                            type: "object";
                                            properties: {
                                                challenge: {
                                                    type: string;
                                                };
                                                rp: {
                                                    type: string;
                                                    properties: {
                                                        name: {
                                                            type: string;
                                                        };
                                                        id: {
                                                            type: string;
                                                        };
                                                    };
                                                };
                                                user: {
                                                    type: string;
                                                    properties: {
                                                        id: {
                                                            type: string;
                                                        };
                                                        name: {
                                                            type: string;
                                                        };
                                                        displayName: {
                                                            type: string;
                                                        };
                                                    };
                                                };
                                                pubKeyCredParams: {
                                                    type: string;
                                                    items: {
                                                        type: string;
                                                        properties: {
                                                            type: {
                                                                type: string;
                                                            };
                                                            alg: {
                                                                type: string;
                                                            };
                                                        };
                                                    };
                                                };
                                                timeout: {
                                                    type: string;
                                                };
                                                excludeCredentials: {
                                                    type: string;
                                                    items: {
                                                        type: string;
                                                        properties: {
                                                            id: {
                                                                type: string;
                                                            };
                                                            type: {
                                                                type: string;
                                                            };
                                                            transports: {
                                                                type: string;
                                                                items: {
                                                                    type: string;
                                                                };
                                                            };
                                                        };
                                                    };
                                                };
                                                authenticatorSelection: {
                                                    type: string;
                                                    properties: {
                                                        authenticatorAttachment: {
                                                            type: string;
                                                        };
                                                        requireResidentKey: {
                                                            type: string;
                                                        };
                                                        userVerification: {
                                                            type: string;
                                                        };
                                                    };
                                                };
                                                attestation: {
                                                    type: string;
                                                };
                                                extensions: {
                                                    type: string;
                                                };
                                            };
                                        };
                                    };
                                };
                            };
                        };
                    };
                };
            }, import("@simplewebauthn/server").PublicKeyCredentialCreationOptionsJSON>;
            generatePasskeyAuthenticationOptions: import("better-call").StrictEndpoint<"/passkey/generate-authenticate-options", {
                method: "GET";
                metadata: {
                    openapi: {
                        operationId: string;
                        description: string;
                        responses: {
                            200: {
                                description: string;
                                content: {
                                    "application/json": {
                                        schema: {
                                            type: "object";
                                            properties: {
                                                challenge: {
                                                    type: string;
                                                };
                                                rp: {
                                                    type: string;
                                                    properties: {
                                                        name: {
                                                            type: string;
                                                        };
                                                        id: {
                                                            type: string;
                                                        };
                                                    };
                                                };
                                                user: {
                                                    type: string;
                                                    properties: {
                                                        id: {
                                                            type: string;
                                                        };
                                                        name: {
                                                            type: string;
                                                        };
                                                        displayName: {
                                                            type: string;
                                                        };
                                                    };
                                                };
                                                timeout: {
                                                    type: string;
                                                };
                                                allowCredentials: {
                                                    type: string;
                                                    items: {
                                                        type: string;
                                                        properties: {
                                                            id: {
                                                                type: string;
                                                            };
                                                            type: {
                                                                type: string;
                                                            };
                                                            transports: {
                                                                type: string;
                                                                items: {
                                                                    type: string;
                                                                };
                                                            };
                                                        };
                                                    };
                                                };
                                                userVerification: {
                                                    type: string;
                                                };
                                                authenticatorSelection: {
                                                    type: string;
                                                    properties: {
                                                        authenticatorAttachment: {
                                                            type: string;
                                                        };
                                                        requireResidentKey: {
                                                            type: string;
                                                        };
                                                        userVerification: {
                                                            type: string;
                                                        };
                                                    };
                                                };
                                                extensions: {
                                                    type: string;
                                                };
                                            };
                                        };
                                    };
                                };
                            };
                        };
                    };
                };
            }, import("@simplewebauthn/server").PublicKeyCredentialRequestOptionsJSON>;
            verifyPasskeyRegistration: import("better-call").StrictEndpoint<"/passkey/verify-registration", {
                method: "POST";
                body: import("zod").ZodObject<{
                    response: import("zod").ZodAny;
                    name: import("zod").ZodOptional<import("zod").ZodString>;
                }, import("better-auth").$strip>;
                use: ((inputContext: import("better-call").MiddlewareInputContext<import("better-call").MiddlewareOptions>) => Promise<{
                    session: {
                        session: Record<string, any> & {
                            id: string;
                            createdAt: Date;
                            updatedAt: Date;
                            userId: string;
                            expiresAt: Date;
                            token: string;
                            ipAddress?: string | null | undefined;
                            userAgent?: string | null | undefined;
                        };
                        user: Record<string, any> & {
                            id: string;
                            createdAt: Date;
                            updatedAt: Date;
                            email: string;
                            emailVerified: boolean;
                            name: string;
                            image?: string | null | undefined;
                        };
                    };
                }>)[];
                metadata: {
                    openapi: {
                        operationId: string;
                        description: string;
                        responses: {
                            200: {
                                description: string;
                                content: {
                                    "application/json": {
                                        schema: {
                                            $ref: string;
                                        };
                                    };
                                };
                            };
                            400: {
                                description: string;
                            };
                        };
                    };
                };
            }, import("@better-auth/passkey").Passkey>;
            verifyPasskeyAuthentication: import("better-call").StrictEndpoint<"/passkey/verify-authentication", {
                method: "POST";
                body: import("zod").ZodObject<{
                    response: import("zod").ZodRecord<import("zod").ZodAny, import("zod").ZodAny>;
                }, import("better-auth").$strip>;
                metadata: {
                    openapi: {
                        operationId: string;
                        description: string;
                        responses: {
                            200: {
                                description: string;
                                content: {
                                    "application/json": {
                                        schema: {
                                            type: "object";
                                            properties: {
                                                session: {
                                                    $ref: string;
                                                };
                                                user: {
                                                    $ref: string;
                                                };
                                            };
                                        };
                                    };
                                };
                            };
                        };
                    };
                    $Infer: {
                        body: {
                            response: import("@simplewebauthn/server").AuthenticationResponseJSON;
                        };
                    };
                };
            }, {
                session: {
                    id: string;
                    createdAt: Date;
                    updatedAt: Date;
                    userId: string;
                    expiresAt: Date;
                    token: string;
                    ipAddress?: string | null | undefined;
                    userAgent?: string | null | undefined;
                };
            }>;
            listPasskeys: import("better-call").StrictEndpoint<"/passkey/list-user-passkeys", {
                method: "GET";
                use: ((inputContext: import("better-call").MiddlewareInputContext<import("better-call").MiddlewareOptions>) => Promise<{
                    session: {
                        session: Record<string, any> & {
                            id: string;
                            createdAt: Date;
                            updatedAt: Date;
                            userId: string;
                            expiresAt: Date;
                            token: string;
                            ipAddress?: string | null | undefined;
                            userAgent?: string | null | undefined;
                        };
                        user: Record<string, any> & {
                            id: string;
                            createdAt: Date;
                            updatedAt: Date;
                            email: string;
                            emailVerified: boolean;
                            name: string;
                            image?: string | null | undefined;
                        };
                    };
                }>)[];
                metadata: {
                    openapi: {
                        description: string;
                        responses: {
                            "200": {
                                description: string;
                                content: {
                                    "application/json": {
                                        schema: {
                                            type: "array";
                                            items: {
                                                $ref: string;
                                                required: string[];
                                            };
                                            description: string;
                                        };
                                    };
                                };
                            };
                        };
                    };
                };
            }, import("@better-auth/passkey").Passkey[]>;
            deletePasskey: import("better-call").StrictEndpoint<"/passkey/delete-passkey", {
                method: "POST";
                body: import("zod").ZodObject<{
                    id: import("zod").ZodString;
                }, import("better-auth").$strip>;
                use: ((inputContext: import("better-call").MiddlewareInputContext<import("better-call").MiddlewareOptions>) => Promise<{
                    session: {
                        session: Record<string, any> & {
                            id: string;
                            createdAt: Date;
                            updatedAt: Date;
                            userId: string;
                            expiresAt: Date;
                            token: string;
                            ipAddress?: string | null | undefined;
                            userAgent?: string | null | undefined;
                        };
                        user: Record<string, any> & {
                            id: string;
                            createdAt: Date;
                            updatedAt: Date;
                            email: string;
                            emailVerified: boolean;
                            name: string;
                            image?: string | null | undefined;
                        };
                    };
                }>)[];
                metadata: {
                    openapi: {
                        description: string;
                        responses: {
                            "200": {
                                description: string;
                                content: {
                                    "application/json": {
                                        schema: {
                                            type: "object";
                                            properties: {
                                                status: {
                                                    type: string;
                                                    description: string;
                                                };
                                            };
                                            required: string[];
                                        };
                                    };
                                };
                            };
                        };
                    };
                };
            }, {
                status: boolean;
            }>;
            updatePasskey: import("better-call").StrictEndpoint<"/passkey/update-passkey", {
                method: "POST";
                body: import("zod").ZodObject<{
                    id: import("zod").ZodString;
                    name: import("zod").ZodString;
                }, import("better-auth").$strip>;
                use: ((inputContext: import("better-call").MiddlewareInputContext<import("better-call").MiddlewareOptions>) => Promise<{
                    session: {
                        session: Record<string, any> & {
                            id: string;
                            createdAt: Date;
                            updatedAt: Date;
                            userId: string;
                            expiresAt: Date;
                            token: string;
                            ipAddress?: string | null | undefined;
                            userAgent?: string | null | undefined;
                        };
                        user: Record<string, any> & {
                            id: string;
                            createdAt: Date;
                            updatedAt: Date;
                            email: string;
                            emailVerified: boolean;
                            name: string;
                            image?: string | null | undefined;
                        };
                    };
                }>)[];
                metadata: {
                    openapi: {
                        description: string;
                        responses: {
                            "200": {
                                description: string;
                                content: {
                                    "application/json": {
                                        schema: {
                                            type: "object";
                                            properties: {
                                                passkey: {
                                                    $ref: string;
                                                };
                                            };
                                            required: string[];
                                        };
                                    };
                                };
                            };
                        };
                    };
                };
            }, {
                passkey: import("@better-auth/passkey").Passkey;
            }>;
        };
        schema: {
            passkey: {
                fields: {
                    name: {
                        type: "string";
                        required: false;
                    };
                    publicKey: {
                        type: "string";
                        required: true;
                    };
                    userId: {
                        type: "string";
                        references: {
                            model: string;
                            field: string;
                        };
                        required: true;
                        index: true;
                    };
                    credentialID: {
                        type: "string";
                        required: true;
                        index: true;
                    };
                    counter: {
                        type: "number";
                        required: true;
                    };
                    deviceType: {
                        type: "string";
                        required: true;
                    };
                    backedUp: {
                        type: "boolean";
                        required: true;
                    };
                    transports: {
                        type: "string";
                        required: false;
                    };
                    createdAt: {
                        type: "date";
                        required: false;
                    };
                    aaguid: {
                        type: "string";
                        required: false;
                    };
                };
            };
        };
        $ERROR_CODES: {
            CHALLENGE_NOT_FOUND: import("better-auth").RawError<"CHALLENGE_NOT_FOUND">;
            YOU_ARE_NOT_ALLOWED_TO_REGISTER_THIS_PASSKEY: import("better-auth").RawError<"YOU_ARE_NOT_ALLOWED_TO_REGISTER_THIS_PASSKEY">;
            FAILED_TO_VERIFY_REGISTRATION: import("better-auth").RawError<"FAILED_TO_VERIFY_REGISTRATION">;
            PASSKEY_NOT_FOUND: import("better-auth").RawError<"PASSKEY_NOT_FOUND">;
            AUTHENTICATION_FAILED: import("better-auth").RawError<"AUTHENTICATION_FAILED">;
            UNABLE_TO_CREATE_SESSION: import("better-auth").RawError<"UNABLE_TO_CREATE_SESSION">;
            FAILED_TO_UPDATE_PASSKEY: import("better-auth").RawError<"FAILED_TO_UPDATE_PASSKEY">;
        };
        options: import("@better-auth/passkey").PasskeyOptions | undefined;
    }, {
        id: "username";
        init(ctx: import("better-auth").AuthContext): {
            options: {
                databaseHooks: {
                    user: {
                        create: {
                            before(user: {
                                id: string;
                                createdAt: Date;
                                updatedAt: Date;
                                email: string;
                                emailVerified: boolean;
                                name: string;
                                image?: string | null | undefined;
                            } & Record<string, unknown>, context: import("better-auth").GenericEndpointContext | null): Promise<{
                                data: {
                                    displayUsername?: string | undefined;
                                    username?: string | undefined;
                                    id: string;
                                    createdAt: Date;
                                    updatedAt: Date;
                                    email: string;
                                    emailVerified: boolean;
                                    name: string;
                                    image?: string | null | undefined;
                                };
                            }>;
                        };
                        update: {
                            before(user: Partial<{
                                id: string;
                                createdAt: Date;
                                updatedAt: Date;
                                email: string;
                                emailVerified: boolean;
                                name: string;
                                image?: string | null | undefined;
                            }> & Record<string, unknown>, context: import("better-auth").GenericEndpointContext | null): Promise<{
                                data: {
                                    displayUsername?: string | undefined;
                                    username?: string | undefined;
                                    id?: string | undefined;
                                    createdAt?: Date | undefined;
                                    updatedAt?: Date | undefined;
                                    email?: string | undefined;
                                    emailVerified?: boolean | undefined;
                                    name?: string | undefined;
                                    image?: string | null | undefined;
                                };
                            }>;
                        };
                    };
                };
            };
        };
        endpoints: {
            signInUsername: import("better-call").StrictEndpoint<"/sign-in/username", {
                method: "POST";
                body: import("zod").ZodObject<{
                    username: import("zod").ZodString;
                    password: import("zod").ZodString;
                    rememberMe: import("zod").ZodOptional<import("zod").ZodBoolean>;
                    callbackURL: import("zod").ZodOptional<import("zod").ZodString>;
                }, import("better-auth").$strip>;
                metadata: {
                    openapi: {
                        summary: string;
                        description: string;
                        responses: {
                            200: {
                                description: string;
                                content: {
                                    "application/json": {
                                        schema: {
                                            type: "object";
                                            properties: {
                                                token: {
                                                    type: string;
                                                    description: string;
                                                };
                                                user: {
                                                    $ref: string;
                                                };
                                            };
                                            required: string[];
                                        };
                                    };
                                };
                            };
                            422: {
                                description: string;
                                content: {
                                    "application/json": {
                                        schema: {
                                            type: "object";
                                            properties: {
                                                message: {
                                                    type: string;
                                                };
                                            };
                                        };
                                    };
                                };
                            };
                        };
                    };
                };
            }, {
                token: string;
                user: {
                    id: string;
                    createdAt: Date;
                    updatedAt: Date;
                    email: string;
                    emailVerified: boolean;
                    name: string;
                    image?: string | null | undefined;
                } & {
                    username: string;
                    displayUsername: string;
                };
            }>;
            isUsernameAvailable: import("better-call").StrictEndpoint<"/is-username-available", {
                method: "POST";
                body: import("zod").ZodObject<{
                    username: import("zod").ZodString;
                }, import("better-auth").$strip>;
            }, {
                available: boolean;
            }>;
        };
        schema: {
            user: {
                fields: {
                    username: {
                        type: "string";
                        required: false;
                        sortable: true;
                        unique: true;
                        returned: true;
                        transform: {
                            input(value: import("better-auth").DBPrimitive): string | number | boolean | Date | Record<string, unknown> | unknown[] | null | undefined;
                        };
                    };
                    displayUsername: {
                        type: "string";
                        required: false;
                        transform: {
                            input(value: import("better-auth").DBPrimitive): string | number | boolean | Date | Record<string, unknown> | unknown[] | null | undefined;
                        };
                    };
                };
            };
        };
        hooks: {
            before: {
                matcher(context: import("better-auth").HookEndpointContext): boolean;
                handler: (inputContext: import("better-call").MiddlewareInputContext<import("better-call").MiddlewareOptions>) => Promise<void>;
            }[];
        };
        options: import("better-auth/plugins").UsernameOptions | undefined;
        $ERROR_CODES: {
            EMAIL_NOT_VERIFIED: import("better-auth").RawError<"EMAIL_NOT_VERIFIED">;
            UNEXPECTED_ERROR: import("better-auth").RawError<"UNEXPECTED_ERROR">;
            INVALID_USERNAME_OR_PASSWORD: import("better-auth").RawError<"INVALID_USERNAME_OR_PASSWORD">;
            USERNAME_IS_ALREADY_TAKEN: import("better-auth").RawError<"USERNAME_IS_ALREADY_TAKEN">;
            USERNAME_TOO_SHORT: import("better-auth").RawError<"USERNAME_TOO_SHORT">;
            USERNAME_TOO_LONG: import("better-auth").RawError<"USERNAME_TOO_LONG">;
            INVALID_USERNAME: import("better-auth").RawError<"INVALID_USERNAME">;
            INVALID_DISPLAY_USERNAME: import("better-auth").RawError<"INVALID_DISPLAY_USERNAME">;
        };
    }, {
        id: "admin";
        init(): {
            options: {
                databaseHooks: {
                    user: {
                        create: {
                            before(user: {
                                id: string;
                                createdAt: Date;
                                updatedAt: Date;
                                email: string;
                                emailVerified: boolean;
                                name: string;
                                image?: string | null | undefined;
                            } & Record<string, unknown>): Promise<{
                                data: {
                                    id: string;
                                    createdAt: Date;
                                    updatedAt: Date;
                                    email: string;
                                    emailVerified: boolean;
                                    name: string;
                                    image?: string | null | undefined;
                                    role: string;
                                };
                            }>;
                        };
                    };
                    session: {
                        create: {
                            before(session: {
                                id: string;
                                createdAt: Date;
                                updatedAt: Date;
                                userId: string;
                                expiresAt: Date;
                                token: string;
                                ipAddress?: string | null | undefined;
                                userAgent?: string | null | undefined;
                            } & Record<string, unknown>, ctx: import("better-auth").GenericEndpointContext | null): Promise<void>;
                        };
                    };
                };
            };
        };
        hooks: {
            after: {
                matcher(context: import("better-auth").HookEndpointContext): boolean;
                handler: (inputContext: import("better-call").MiddlewareInputContext<import("better-call").MiddlewareOptions>) => Promise<import("better-auth/plugins").SessionWithImpersonatedBy[] | undefined>;
            }[];
        };
        endpoints: {
            setRole: import("better-call").StrictEndpoint<"/admin/set-role", {
                method: "POST";
                body: import("zod").ZodObject<{
                    userId: import("zod").ZodCoercedString<unknown>;
                    role: import("zod").ZodUnion<readonly [import("zod").ZodString, import("zod").ZodArray<import("zod").ZodString>]>;
                }, import("better-auth").$strip>;
                requireHeaders: true;
                use: ((inputContext: import("better-call").MiddlewareInputContext<import("better-call").MiddlewareOptions>) => Promise<{
                    session: {
                        user: import("better-auth/plugins").UserWithRole;
                        session: {
                            id: string;
                            createdAt: Date;
                            updatedAt: Date;
                            userId: string;
                            expiresAt: Date;
                            token: string;
                            ipAddress?: string | null | undefined;
                            userAgent?: string | null | undefined;
                        };
                    };
                }>)[];
                metadata: {
                    openapi: {
                        operationId: string;
                        summary: string;
                        description: string;
                        responses: {
                            200: {
                                description: string;
                                content: {
                                    "application/json": {
                                        schema: {
                                            type: "object";
                                            properties: {
                                                user: {
                                                    $ref: string;
                                                };
                                            };
                                        };
                                    };
                                };
                            };
                        };
                    };
                    $Infer: {
                        body: {
                            userId: string;
                            role: "user" | "admin" | ("user" | "admin")[];
                        };
                    };
                };
            }, {
                user: import("better-auth/plugins").UserWithRole;
            }>;
            getUser: import("better-call").StrictEndpoint<"/admin/get-user", {
                method: "GET";
                query: import("zod").ZodObject<{
                    id: import("zod").ZodString;
                }, import("better-auth").$strip>;
                use: ((inputContext: import("better-call").MiddlewareInputContext<import("better-call").MiddlewareOptions>) => Promise<{
                    session: {
                        user: import("better-auth/plugins").UserWithRole;
                        session: {
                            id: string;
                            createdAt: Date;
                            updatedAt: Date;
                            userId: string;
                            expiresAt: Date;
                            token: string;
                            ipAddress?: string | null | undefined;
                            userAgent?: string | null | undefined;
                        };
                    };
                }>)[];
                metadata: {
                    openapi: {
                        operationId: string;
                        summary: string;
                        description: string;
                        responses: {
                            200: {
                                description: string;
                                content: {
                                    "application/json": {
                                        schema: {
                                            type: "object";
                                            properties: {
                                                user: {
                                                    $ref: string;
                                                };
                                            };
                                        };
                                    };
                                };
                            };
                        };
                    };
                };
            }, import("better-auth/plugins").UserWithRole>;
            createUser: import("better-call").StrictEndpoint<"/admin/create-user", {
                method: "POST";
                body: import("zod").ZodObject<{
                    email: import("zod").ZodString;
                    password: import("zod").ZodOptional<import("zod").ZodString>;
                    name: import("zod").ZodString;
                    role: import("zod").ZodOptional<import("zod").ZodUnion<readonly [import("zod").ZodString, import("zod").ZodArray<import("zod").ZodString>]>>;
                    data: import("zod").ZodOptional<import("zod").ZodRecord<import("zod").ZodString, import("zod").ZodAny>>;
                }, import("better-auth").$strip>;
                metadata: {
                    openapi: {
                        operationId: string;
                        summary: string;
                        description: string;
                        responses: {
                            200: {
                                description: string;
                                content: {
                                    "application/json": {
                                        schema: {
                                            type: "object";
                                            properties: {
                                                user: {
                                                    $ref: string;
                                                };
                                            };
                                        };
                                    };
                                };
                            };
                        };
                    };
                    $Infer: {
                        body: {
                            email: string;
                            password?: string | undefined;
                            name: string;
                            role?: "user" | "admin" | ("user" | "admin")[] | undefined;
                            data?: Record<string, any> | undefined;
                        };
                    };
                };
            }, {
                user: import("better-auth/plugins").UserWithRole;
            }>;
            adminUpdateUser: import("better-call").StrictEndpoint<"/admin/update-user", {
                method: "POST";
                body: import("zod").ZodObject<{
                    userId: import("zod").ZodCoercedString<unknown>;
                    data: import("zod").ZodRecord<import("zod").ZodAny, import("zod").ZodAny>;
                }, import("better-auth").$strip>;
                use: ((inputContext: import("better-call").MiddlewareInputContext<import("better-call").MiddlewareOptions>) => Promise<{
                    session: {
                        user: import("better-auth/plugins").UserWithRole;
                        session: {
                            id: string;
                            createdAt: Date;
                            updatedAt: Date;
                            userId: string;
                            expiresAt: Date;
                            token: string;
                            ipAddress?: string | null | undefined;
                            userAgent?: string | null | undefined;
                        };
                    };
                }>)[];
                metadata: {
                    openapi: {
                        operationId: string;
                        summary: string;
                        description: string;
                        responses: {
                            200: {
                                description: string;
                                content: {
                                    "application/json": {
                                        schema: {
                                            type: "object";
                                            properties: {
                                                user: {
                                                    $ref: string;
                                                };
                                            };
                                        };
                                    };
                                };
                            };
                        };
                    };
                };
            }, import("better-auth/plugins").UserWithRole>;
            listUsers: import("better-call").StrictEndpoint<"/admin/list-users", {
                method: "GET";
                use: ((inputContext: import("better-call").MiddlewareInputContext<import("better-call").MiddlewareOptions>) => Promise<{
                    session: {
                        user: import("better-auth/plugins").UserWithRole;
                        session: {
                            id: string;
                            createdAt: Date;
                            updatedAt: Date;
                            userId: string;
                            expiresAt: Date;
                            token: string;
                            ipAddress?: string | null | undefined;
                            userAgent?: string | null | undefined;
                        };
                    };
                }>)[];
                query: import("zod").ZodObject<{
                    searchValue: import("zod").ZodOptional<import("zod").ZodString>;
                    searchField: import("zod").ZodOptional<import("zod").ZodEnum<{
                        name: "name";
                        email: "email";
                    }>>;
                    searchOperator: import("zod").ZodOptional<import("zod").ZodEnum<{
                        contains: "contains";
                        starts_with: "starts_with";
                        ends_with: "ends_with";
                    }>>;
                    limit: import("zod").ZodOptional<import("zod").ZodUnion<[import("zod").ZodString, import("zod").ZodNumber]>>;
                    offset: import("zod").ZodOptional<import("zod").ZodUnion<[import("zod").ZodString, import("zod").ZodNumber]>>;
                    sortBy: import("zod").ZodOptional<import("zod").ZodString>;
                    sortDirection: import("zod").ZodOptional<import("zod").ZodEnum<{
                        asc: "asc";
                        desc: "desc";
                    }>>;
                    filterField: import("zod").ZodOptional<import("zod").ZodString>;
                    filterValue: import("zod").ZodOptional<import("zod").ZodUnion<[import("zod").ZodUnion<[import("zod").ZodUnion<[import("zod").ZodUnion<[import("zod").ZodString, import("zod").ZodNumber]>, import("zod").ZodBoolean]>, import("zod").ZodArray<import("zod").ZodString>]>, import("zod").ZodArray<import("zod").ZodNumber>]>>;
                    filterOperator: import("zod").ZodOptional<import("zod").ZodEnum<{
                        eq: "eq";
                        ne: "ne";
                        gt: "gt";
                        gte: "gte";
                        lt: "lt";
                        lte: "lte";
                        in: "in";
                        not_in: "not_in";
                        contains: "contains";
                        starts_with: "starts_with";
                        ends_with: "ends_with";
                    }>>;
                }, import("better-auth").$strip>;
                metadata: {
                    openapi: {
                        operationId: string;
                        summary: string;
                        description: string;
                        responses: {
                            200: {
                                description: string;
                                content: {
                                    "application/json": {
                                        schema: {
                                            type: "object";
                                            properties: {
                                                users: {
                                                    type: string;
                                                    items: {
                                                        $ref: string;
                                                    };
                                                };
                                                total: {
                                                    type: string;
                                                };
                                                limit: {
                                                    type: string;
                                                };
                                                offset: {
                                                    type: string;
                                                };
                                            };
                                            required: string[];
                                        };
                                    };
                                };
                            };
                        };
                    };
                };
            }, {
                users: import("better-auth/plugins").UserWithRole[];
                total: number;
            }>;
            listUserSessions: import("better-call").StrictEndpoint<"/admin/list-user-sessions", {
                method: "POST";
                use: ((inputContext: import("better-call").MiddlewareInputContext<import("better-call").MiddlewareOptions>) => Promise<{
                    session: {
                        user: import("better-auth/plugins").UserWithRole;
                        session: {
                            id: string;
                            createdAt: Date;
                            updatedAt: Date;
                            userId: string;
                            expiresAt: Date;
                            token: string;
                            ipAddress?: string | null | undefined;
                            userAgent?: string | null | undefined;
                        };
                    };
                }>)[];
                body: import("zod").ZodObject<{
                    userId: import("zod").ZodCoercedString<unknown>;
                }, import("better-auth").$strip>;
                metadata: {
                    openapi: {
                        operationId: string;
                        summary: string;
                        description: string;
                        responses: {
                            200: {
                                description: string;
                                content: {
                                    "application/json": {
                                        schema: {
                                            type: "object";
                                            properties: {
                                                sessions: {
                                                    type: string;
                                                    items: {
                                                        $ref: string;
                                                    };
                                                };
                                            };
                                        };
                                    };
                                };
                            };
                        };
                    };
                };
            }, {
                sessions: import("better-auth/plugins").SessionWithImpersonatedBy[];
            }>;
            unbanUser: import("better-call").StrictEndpoint<"/admin/unban-user", {
                method: "POST";
                body: import("zod").ZodObject<{
                    userId: import("zod").ZodCoercedString<unknown>;
                }, import("better-auth").$strip>;
                use: ((inputContext: import("better-call").MiddlewareInputContext<import("better-call").MiddlewareOptions>) => Promise<{
                    session: {
                        user: import("better-auth/plugins").UserWithRole;
                        session: {
                            id: string;
                            createdAt: Date;
                            updatedAt: Date;
                            userId: string;
                            expiresAt: Date;
                            token: string;
                            ipAddress?: string | null | undefined;
                            userAgent?: string | null | undefined;
                        };
                    };
                }>)[];
                metadata: {
                    openapi: {
                        operationId: string;
                        summary: string;
                        description: string;
                        responses: {
                            200: {
                                description: string;
                                content: {
                                    "application/json": {
                                        schema: {
                                            type: "object";
                                            properties: {
                                                user: {
                                                    $ref: string;
                                                };
                                            };
                                        };
                                    };
                                };
                            };
                        };
                    };
                };
            }, {
                user: import("better-auth/plugins").UserWithRole;
            }>;
            banUser: import("better-call").StrictEndpoint<"/admin/ban-user", {
                method: "POST";
                body: import("zod").ZodObject<{
                    userId: import("zod").ZodCoercedString<unknown>;
                    banReason: import("zod").ZodOptional<import("zod").ZodString>;
                    banExpiresIn: import("zod").ZodOptional<import("zod").ZodNumber>;
                }, import("better-auth").$strip>;
                use: ((inputContext: import("better-call").MiddlewareInputContext<import("better-call").MiddlewareOptions>) => Promise<{
                    session: {
                        user: import("better-auth/plugins").UserWithRole;
                        session: {
                            id: string;
                            createdAt: Date;
                            updatedAt: Date;
                            userId: string;
                            expiresAt: Date;
                            token: string;
                            ipAddress?: string | null | undefined;
                            userAgent?: string | null | undefined;
                        };
                    };
                }>)[];
                metadata: {
                    openapi: {
                        operationId: string;
                        summary: string;
                        description: string;
                        responses: {
                            200: {
                                description: string;
                                content: {
                                    "application/json": {
                                        schema: {
                                            type: "object";
                                            properties: {
                                                user: {
                                                    $ref: string;
                                                };
                                            };
                                        };
                                    };
                                };
                            };
                        };
                    };
                };
            }, {
                user: import("better-auth/plugins").UserWithRole;
            }>;
            impersonateUser: import("better-call").StrictEndpoint<"/admin/impersonate-user", {
                method: "POST";
                body: import("zod").ZodObject<{
                    userId: import("zod").ZodCoercedString<unknown>;
                }, import("better-auth").$strip>;
                use: ((inputContext: import("better-call").MiddlewareInputContext<import("better-call").MiddlewareOptions>) => Promise<{
                    session: {
                        user: import("better-auth/plugins").UserWithRole;
                        session: {
                            id: string;
                            createdAt: Date;
                            updatedAt: Date;
                            userId: string;
                            expiresAt: Date;
                            token: string;
                            ipAddress?: string | null | undefined;
                            userAgent?: string | null | undefined;
                        };
                    };
                }>)[];
                metadata: {
                    openapi: {
                        operationId: string;
                        summary: string;
                        description: string;
                        responses: {
                            200: {
                                description: string;
                                content: {
                                    "application/json": {
                                        schema: {
                                            type: "object";
                                            properties: {
                                                session: {
                                                    $ref: string;
                                                };
                                                user: {
                                                    $ref: string;
                                                };
                                            };
                                        };
                                    };
                                };
                            };
                        };
                    };
                };
            }, {
                session: {
                    id: string;
                    createdAt: Date;
                    updatedAt: Date;
                    userId: string;
                    expiresAt: Date;
                    token: string;
                    ipAddress?: string | null | undefined;
                    userAgent?: string | null | undefined;
                };
                user: import("better-auth/plugins").UserWithRole;
            }>;
            stopImpersonating: import("better-call").StrictEndpoint<"/admin/stop-impersonating", {
                method: "POST";
                requireHeaders: true;
            }, {
                session: {
                    id: string;
                    createdAt: Date;
                    updatedAt: Date;
                    userId: string;
                    expiresAt: Date;
                    token: string;
                    ipAddress?: string | null | undefined;
                    userAgent?: string | null | undefined;
                } & Record<string, any>;
                user: {
                    id: string;
                    createdAt: Date;
                    updatedAt: Date;
                    email: string;
                    emailVerified: boolean;
                    name: string;
                    image?: string | null | undefined;
                } & Record<string, any>;
            }>;
            revokeUserSession: import("better-call").StrictEndpoint<"/admin/revoke-user-session", {
                method: "POST";
                body: import("zod").ZodObject<{
                    sessionToken: import("zod").ZodString;
                }, import("better-auth").$strip>;
                use: ((inputContext: import("better-call").MiddlewareInputContext<import("better-call").MiddlewareOptions>) => Promise<{
                    session: {
                        user: import("better-auth/plugins").UserWithRole;
                        session: {
                            id: string;
                            createdAt: Date;
                            updatedAt: Date;
                            userId: string;
                            expiresAt: Date;
                            token: string;
                            ipAddress?: string | null | undefined;
                            userAgent?: string | null | undefined;
                        };
                    };
                }>)[];
                metadata: {
                    openapi: {
                        operationId: string;
                        summary: string;
                        description: string;
                        responses: {
                            200: {
                                description: string;
                                content: {
                                    "application/json": {
                                        schema: {
                                            type: "object";
                                            properties: {
                                                success: {
                                                    type: string;
                                                };
                                            };
                                        };
                                    };
                                };
                            };
                        };
                    };
                };
            }, {
                success: boolean;
            }>;
            revokeUserSessions: import("better-call").StrictEndpoint<"/admin/revoke-user-sessions", {
                method: "POST";
                body: import("zod").ZodObject<{
                    userId: import("zod").ZodCoercedString<unknown>;
                }, import("better-auth").$strip>;
                use: ((inputContext: import("better-call").MiddlewareInputContext<import("better-call").MiddlewareOptions>) => Promise<{
                    session: {
                        user: import("better-auth/plugins").UserWithRole;
                        session: {
                            id: string;
                            createdAt: Date;
                            updatedAt: Date;
                            userId: string;
                            expiresAt: Date;
                            token: string;
                            ipAddress?: string | null | undefined;
                            userAgent?: string | null | undefined;
                        };
                    };
                }>)[];
                metadata: {
                    openapi: {
                        operationId: string;
                        summary: string;
                        description: string;
                        responses: {
                            200: {
                                description: string;
                                content: {
                                    "application/json": {
                                        schema: {
                                            type: "object";
                                            properties: {
                                                success: {
                                                    type: string;
                                                };
                                            };
                                        };
                                    };
                                };
                            };
                        };
                    };
                };
            }, {
                success: boolean;
            }>;
            removeUser: import("better-call").StrictEndpoint<"/admin/remove-user", {
                method: "POST";
                body: import("zod").ZodObject<{
                    userId: import("zod").ZodCoercedString<unknown>;
                }, import("better-auth").$strip>;
                use: ((inputContext: import("better-call").MiddlewareInputContext<import("better-call").MiddlewareOptions>) => Promise<{
                    session: {
                        user: import("better-auth/plugins").UserWithRole;
                        session: {
                            id: string;
                            createdAt: Date;
                            updatedAt: Date;
                            userId: string;
                            expiresAt: Date;
                            token: string;
                            ipAddress?: string | null | undefined;
                            userAgent?: string | null | undefined;
                        };
                    };
                }>)[];
                metadata: {
                    openapi: {
                        operationId: string;
                        summary: string;
                        description: string;
                        responses: {
                            200: {
                                description: string;
                                content: {
                                    "application/json": {
                                        schema: {
                                            type: "object";
                                            properties: {
                                                success: {
                                                    type: string;
                                                };
                                            };
                                        };
                                    };
                                };
                            };
                        };
                    };
                };
            }, {
                success: boolean;
            }>;
            setUserPassword: import("better-call").StrictEndpoint<"/admin/set-user-password", {
                method: "POST";
                body: import("zod").ZodObject<{
                    newPassword: import("zod").ZodString;
                    userId: import("zod").ZodCoercedString<unknown>;
                }, import("better-auth").$strip>;
                use: ((inputContext: import("better-call").MiddlewareInputContext<import("better-call").MiddlewareOptions>) => Promise<{
                    session: {
                        user: import("better-auth/plugins").UserWithRole;
                        session: {
                            id: string;
                            createdAt: Date;
                            updatedAt: Date;
                            userId: string;
                            expiresAt: Date;
                            token: string;
                            ipAddress?: string | null | undefined;
                            userAgent?: string | null | undefined;
                        };
                    };
                }>)[];
                metadata: {
                    openapi: {
                        operationId: string;
                        summary: string;
                        description: string;
                        responses: {
                            200: {
                                description: string;
                                content: {
                                    "application/json": {
                                        schema: {
                                            type: "object";
                                            properties: {
                                                status: {
                                                    type: string;
                                                };
                                            };
                                        };
                                    };
                                };
                            };
                        };
                    };
                };
            }, {
                status: boolean;
            }>;
            userHasPermission: import("better-call").StrictEndpoint<"/admin/has-permission", {
                method: "POST";
                body: import("zod").ZodIntersection<import("zod").ZodObject<{
                    userId: import("zod").ZodOptional<import("zod").ZodCoercedString<unknown>>;
                    role: import("zod").ZodOptional<import("zod").ZodString>;
                }, import("better-auth").$strip>, import("zod").ZodUnion<readonly [import("zod").ZodObject<{
                    permission: import("zod").ZodRecord<import("zod").ZodString, import("zod").ZodArray<import("zod").ZodString>>;
                    permissions: import("zod").ZodUndefined;
                }, import("better-auth").$strip>, import("zod").ZodObject<{
                    permission: import("zod").ZodUndefined;
                    permissions: import("zod").ZodRecord<import("zod").ZodString, import("zod").ZodArray<import("zod").ZodString>>;
                }, import("better-auth").$strip>]>>;
                metadata: {
                    openapi: {
                        description: string;
                        requestBody: {
                            content: {
                                "application/json": {
                                    schema: {
                                        type: "object";
                                        properties: {
                                            permissions: {
                                                type: string;
                                                description: string;
                                            };
                                        };
                                        required: string[];
                                    };
                                };
                            };
                        };
                        responses: {
                            "200": {
                                description: string;
                                content: {
                                    "application/json": {
                                        schema: {
                                            type: "object";
                                            properties: {
                                                error: {
                                                    type: string;
                                                };
                                                success: {
                                                    type: string;
                                                };
                                            };
                                            required: string[];
                                        };
                                    };
                                };
                            };
                        };
                    };
                    $Infer: {
                        body: {
                            permissions: {
                                readonly user?: ("set-role" | "create" | "update" | "delete" | "list" | "get" | "ban" | "impersonate" | "impersonate-admins" | "set-password")[] | undefined;
                                readonly session?: ("delete" | "list" | "revoke")[] | undefined;
                            };
                        } & {
                            userId?: string | undefined;
                            role?: "user" | "admin" | undefined;
                        };
                    };
                };
            }, {
                error: null;
                success: boolean;
            }>;
        };
        $ERROR_CODES: {
            USER_ALREADY_EXISTS_USE_ANOTHER_EMAIL: import("better-auth").RawError<"USER_ALREADY_EXISTS_USE_ANOTHER_EMAIL">;
            FAILED_TO_CREATE_USER: import("better-auth").RawError<"FAILED_TO_CREATE_USER">;
            USER_ALREADY_EXISTS: import("better-auth").RawError<"USER_ALREADY_EXISTS">;
            YOU_CANNOT_BAN_YOURSELF: import("better-auth").RawError<"YOU_CANNOT_BAN_YOURSELF">;
            YOU_ARE_NOT_ALLOWED_TO_CHANGE_USERS_ROLE: import("better-auth").RawError<"YOU_ARE_NOT_ALLOWED_TO_CHANGE_USERS_ROLE">;
            YOU_ARE_NOT_ALLOWED_TO_CREATE_USERS: import("better-auth").RawError<"YOU_ARE_NOT_ALLOWED_TO_CREATE_USERS">;
            YOU_ARE_NOT_ALLOWED_TO_LIST_USERS: import("better-auth").RawError<"YOU_ARE_NOT_ALLOWED_TO_LIST_USERS">;
            YOU_ARE_NOT_ALLOWED_TO_LIST_USERS_SESSIONS: import("better-auth").RawError<"YOU_ARE_NOT_ALLOWED_TO_LIST_USERS_SESSIONS">;
            YOU_ARE_NOT_ALLOWED_TO_BAN_USERS: import("better-auth").RawError<"YOU_ARE_NOT_ALLOWED_TO_BAN_USERS">;
            YOU_ARE_NOT_ALLOWED_TO_IMPERSONATE_USERS: import("better-auth").RawError<"YOU_ARE_NOT_ALLOWED_TO_IMPERSONATE_USERS">;
            YOU_ARE_NOT_ALLOWED_TO_REVOKE_USERS_SESSIONS: import("better-auth").RawError<"YOU_ARE_NOT_ALLOWED_TO_REVOKE_USERS_SESSIONS">;
            YOU_ARE_NOT_ALLOWED_TO_DELETE_USERS: import("better-auth").RawError<"YOU_ARE_NOT_ALLOWED_TO_DELETE_USERS">;
            YOU_ARE_NOT_ALLOWED_TO_SET_USERS_PASSWORD: import("better-auth").RawError<"YOU_ARE_NOT_ALLOWED_TO_SET_USERS_PASSWORD">;
            BANNED_USER: import("better-auth").RawError<"BANNED_USER">;
            YOU_ARE_NOT_ALLOWED_TO_GET_USER: import("better-auth").RawError<"YOU_ARE_NOT_ALLOWED_TO_GET_USER">;
            NO_DATA_TO_UPDATE: import("better-auth").RawError<"NO_DATA_TO_UPDATE">;
            YOU_ARE_NOT_ALLOWED_TO_UPDATE_USERS: import("better-auth").RawError<"YOU_ARE_NOT_ALLOWED_TO_UPDATE_USERS">;
            YOU_CANNOT_REMOVE_YOURSELF: import("better-auth").RawError<"YOU_CANNOT_REMOVE_YOURSELF">;
            YOU_ARE_NOT_ALLOWED_TO_SET_NON_EXISTENT_VALUE: import("better-auth").RawError<"YOU_ARE_NOT_ALLOWED_TO_SET_NON_EXISTENT_VALUE">;
            YOU_CANNOT_IMPERSONATE_ADMINS: import("better-auth").RawError<"YOU_CANNOT_IMPERSONATE_ADMINS">;
            INVALID_ROLE_TYPE: import("better-auth").RawError<"INVALID_ROLE_TYPE">;
        };
        schema: {
            user: {
                fields: {
                    role: {
                        type: "string";
                        required: false;
                        input: false;
                    };
                    banned: {
                        type: "boolean";
                        defaultValue: false;
                        required: false;
                        input: false;
                    };
                    banReason: {
                        type: "string";
                        required: false;
                        input: false;
                    };
                    banExpires: {
                        type: "date";
                        required: false;
                        input: false;
                    };
                };
            };
            session: {
                fields: {
                    impersonatedBy: {
                        type: "string";
                        required: false;
                    };
                };
            };
        };
        options: NoInfer<import("better-auth/plugins").AdminOptions>;
    }, {
        id: "organization";
        endpoints: import("better-auth/plugins").OrganizationEndpoints<import("better-auth/plugins").OrganizationOptions & {
            teams: {
                enabled: true;
            };
            dynamicAccessControl?: {
                enabled?: false | undefined;
            } | undefined;
        }> & import("better-auth/plugins").TeamEndpoints<import("better-auth/plugins").OrganizationOptions & {
            teams: {
                enabled: true;
            };
            dynamicAccessControl?: {
                enabled?: false | undefined;
            } | undefined;
        }>;
        schema: import("better-auth/plugins").OrganizationSchema<import("better-auth/plugins").OrganizationOptions & {
            teams: {
                enabled: true;
            };
            dynamicAccessControl?: {
                enabled?: false | undefined;
            } | undefined;
        }>;
        $Infer: {
            Organization: {
                id: string;
                name: string;
                slug: string;
                createdAt: Date;
                logo?: string | null | undefined;
                metadata?: any;
            };
            Invitation: {
                id: string;
                organizationId: string;
                email: string;
                role: "admin" | "member" | "owner";
                status: import("better-auth/plugins").InvitationStatus;
                inviterId: string;
                expiresAt: Date;
                createdAt: Date;
                teamId?: string | undefined | undefined;
            };
            Member: {
                id: string;
                organizationId: string;
                role: "admin" | "member" | "owner";
                createdAt: Date;
                userId: string;
                teamId?: string | undefined | undefined;
                user: {
                    id: string;
                    email: string;
                    name: string;
                    image?: string | undefined;
                };
            };
            Team: {
                id: string;
                name: string;
                organizationId: string;
                createdAt: Date;
                updatedAt?: Date | undefined;
            };
            TeamMember: {
                id: string;
                teamId: string;
                userId: string;
                createdAt: Date;
            };
            ActiveOrganization: {
                members: {
                    id: string;
                    organizationId: string;
                    role: "admin" | "member" | "owner";
                    createdAt: Date;
                    userId: string;
                    teamId?: string | undefined | undefined;
                    user: {
                        id: string;
                        email: string;
                        name: string;
                        image?: string | undefined;
                    };
                }[];
                invitations: {
                    id: string;
                    organizationId: string;
                    email: string;
                    role: "admin" | "member" | "owner";
                    status: import("better-auth/plugins").InvitationStatus;
                    inviterId: string;
                    expiresAt: Date;
                    createdAt: Date;
                    teamId?: string | undefined | undefined;
                }[];
                teams: {
                    id: string;
                    name: string;
                    organizationId: string;
                    createdAt: Date;
                    updatedAt?: Date | undefined;
                }[];
            } & {
                id: string;
                name: string;
                slug: string;
                createdAt: Date;
                logo?: string | null | undefined;
                metadata?: any;
            };
        };
        $ERROR_CODES: {
            YOU_ARE_NOT_ALLOWED_TO_CREATE_A_NEW_ORGANIZATION: import("better-auth").RawError<"YOU_ARE_NOT_ALLOWED_TO_CREATE_A_NEW_ORGANIZATION">;
            YOU_HAVE_REACHED_THE_MAXIMUM_NUMBER_OF_ORGANIZATIONS: import("better-auth").RawError<"YOU_HAVE_REACHED_THE_MAXIMUM_NUMBER_OF_ORGANIZATIONS">;
            ORGANIZATION_ALREADY_EXISTS: import("better-auth").RawError<"ORGANIZATION_ALREADY_EXISTS">;
            ORGANIZATION_SLUG_ALREADY_TAKEN: import("better-auth").RawError<"ORGANIZATION_SLUG_ALREADY_TAKEN">;
            ORGANIZATION_NOT_FOUND: import("better-auth").RawError<"ORGANIZATION_NOT_FOUND">;
            USER_IS_NOT_A_MEMBER_OF_THE_ORGANIZATION: import("better-auth").RawError<"USER_IS_NOT_A_MEMBER_OF_THE_ORGANIZATION">;
            YOU_ARE_NOT_ALLOWED_TO_UPDATE_THIS_ORGANIZATION: import("better-auth").RawError<"YOU_ARE_NOT_ALLOWED_TO_UPDATE_THIS_ORGANIZATION">;
            YOU_ARE_NOT_ALLOWED_TO_DELETE_THIS_ORGANIZATION: import("better-auth").RawError<"YOU_ARE_NOT_ALLOWED_TO_DELETE_THIS_ORGANIZATION">;
            NO_ACTIVE_ORGANIZATION: import("better-auth").RawError<"NO_ACTIVE_ORGANIZATION">;
            USER_IS_ALREADY_A_MEMBER_OF_THIS_ORGANIZATION: import("better-auth").RawError<"USER_IS_ALREADY_A_MEMBER_OF_THIS_ORGANIZATION">;
            MEMBER_NOT_FOUND: import("better-auth").RawError<"MEMBER_NOT_FOUND">;
            ROLE_NOT_FOUND: import("better-auth").RawError<"ROLE_NOT_FOUND">;
            YOU_ARE_NOT_ALLOWED_TO_CREATE_A_NEW_TEAM: import("better-auth").RawError<"YOU_ARE_NOT_ALLOWED_TO_CREATE_A_NEW_TEAM">;
            TEAM_ALREADY_EXISTS: import("better-auth").RawError<"TEAM_ALREADY_EXISTS">;
            TEAM_NOT_FOUND: import("better-auth").RawError<"TEAM_NOT_FOUND">;
            YOU_CANNOT_LEAVE_THE_ORGANIZATION_AS_THE_ONLY_OWNER: import("better-auth").RawError<"YOU_CANNOT_LEAVE_THE_ORGANIZATION_AS_THE_ONLY_OWNER">;
            YOU_CANNOT_LEAVE_THE_ORGANIZATION_WITHOUT_AN_OWNER: import("better-auth").RawError<"YOU_CANNOT_LEAVE_THE_ORGANIZATION_WITHOUT_AN_OWNER">;
            YOU_ARE_NOT_ALLOWED_TO_DELETE_THIS_MEMBER: import("better-auth").RawError<"YOU_ARE_NOT_ALLOWED_TO_DELETE_THIS_MEMBER">;
            YOU_ARE_NOT_ALLOWED_TO_INVITE_USERS_TO_THIS_ORGANIZATION: import("better-auth").RawError<"YOU_ARE_NOT_ALLOWED_TO_INVITE_USERS_TO_THIS_ORGANIZATION">;
            USER_IS_ALREADY_INVITED_TO_THIS_ORGANIZATION: import("better-auth").RawError<"USER_IS_ALREADY_INVITED_TO_THIS_ORGANIZATION">;
            INVITATION_NOT_FOUND: import("better-auth").RawError<"INVITATION_NOT_FOUND">;
            YOU_ARE_NOT_THE_RECIPIENT_OF_THE_INVITATION: import("better-auth").RawError<"YOU_ARE_NOT_THE_RECIPIENT_OF_THE_INVITATION">;
            EMAIL_VERIFICATION_REQUIRED_BEFORE_ACCEPTING_OR_REJECTING_INVITATION: import("better-auth").RawError<"EMAIL_VERIFICATION_REQUIRED_BEFORE_ACCEPTING_OR_REJECTING_INVITATION">;
            YOU_ARE_NOT_ALLOWED_TO_CANCEL_THIS_INVITATION: import("better-auth").RawError<"YOU_ARE_NOT_ALLOWED_TO_CANCEL_THIS_INVITATION">;
            INVITER_IS_NO_LONGER_A_MEMBER_OF_THE_ORGANIZATION: import("better-auth").RawError<"INVITER_IS_NO_LONGER_A_MEMBER_OF_THE_ORGANIZATION">;
            YOU_ARE_NOT_ALLOWED_TO_INVITE_USER_WITH_THIS_ROLE: import("better-auth").RawError<"YOU_ARE_NOT_ALLOWED_TO_INVITE_USER_WITH_THIS_ROLE">;
            FAILED_TO_RETRIEVE_INVITATION: import("better-auth").RawError<"FAILED_TO_RETRIEVE_INVITATION">;
            YOU_HAVE_REACHED_THE_MAXIMUM_NUMBER_OF_TEAMS: import("better-auth").RawError<"YOU_HAVE_REACHED_THE_MAXIMUM_NUMBER_OF_TEAMS">;
            UNABLE_TO_REMOVE_LAST_TEAM: import("better-auth").RawError<"UNABLE_TO_REMOVE_LAST_TEAM">;
            YOU_ARE_NOT_ALLOWED_TO_UPDATE_THIS_MEMBER: import("better-auth").RawError<"YOU_ARE_NOT_ALLOWED_TO_UPDATE_THIS_MEMBER">;
            ORGANIZATION_MEMBERSHIP_LIMIT_REACHED: import("better-auth").RawError<"ORGANIZATION_MEMBERSHIP_LIMIT_REACHED">;
            YOU_ARE_NOT_ALLOWED_TO_CREATE_TEAMS_IN_THIS_ORGANIZATION: import("better-auth").RawError<"YOU_ARE_NOT_ALLOWED_TO_CREATE_TEAMS_IN_THIS_ORGANIZATION">;
            YOU_ARE_NOT_ALLOWED_TO_DELETE_TEAMS_IN_THIS_ORGANIZATION: import("better-auth").RawError<"YOU_ARE_NOT_ALLOWED_TO_DELETE_TEAMS_IN_THIS_ORGANIZATION">;
            YOU_ARE_NOT_ALLOWED_TO_UPDATE_THIS_TEAM: import("better-auth").RawError<"YOU_ARE_NOT_ALLOWED_TO_UPDATE_THIS_TEAM">;
            YOU_ARE_NOT_ALLOWED_TO_DELETE_THIS_TEAM: import("better-auth").RawError<"YOU_ARE_NOT_ALLOWED_TO_DELETE_THIS_TEAM">;
            INVITATION_LIMIT_REACHED: import("better-auth").RawError<"INVITATION_LIMIT_REACHED">;
            TEAM_MEMBER_LIMIT_REACHED: import("better-auth").RawError<"TEAM_MEMBER_LIMIT_REACHED">;
            USER_IS_NOT_A_MEMBER_OF_THE_TEAM: import("better-auth").RawError<"USER_IS_NOT_A_MEMBER_OF_THE_TEAM">;
            YOU_CAN_NOT_ACCESS_THE_MEMBERS_OF_THIS_TEAM: import("better-auth").RawError<"YOU_CAN_NOT_ACCESS_THE_MEMBERS_OF_THIS_TEAM">;
            YOU_DO_NOT_HAVE_AN_ACTIVE_TEAM: import("better-auth").RawError<"YOU_DO_NOT_HAVE_AN_ACTIVE_TEAM">;
            YOU_ARE_NOT_ALLOWED_TO_CREATE_A_NEW_TEAM_MEMBER: import("better-auth").RawError<"YOU_ARE_NOT_ALLOWED_TO_CREATE_A_NEW_TEAM_MEMBER">;
            YOU_ARE_NOT_ALLOWED_TO_REMOVE_A_TEAM_MEMBER: import("better-auth").RawError<"YOU_ARE_NOT_ALLOWED_TO_REMOVE_A_TEAM_MEMBER">;
            YOU_ARE_NOT_ALLOWED_TO_ACCESS_THIS_ORGANIZATION: import("better-auth").RawError<"YOU_ARE_NOT_ALLOWED_TO_ACCESS_THIS_ORGANIZATION">;
            YOU_ARE_NOT_A_MEMBER_OF_THIS_ORGANIZATION: import("better-auth").RawError<"YOU_ARE_NOT_A_MEMBER_OF_THIS_ORGANIZATION">;
            MISSING_AC_INSTANCE: import("better-auth").RawError<"MISSING_AC_INSTANCE">;
            YOU_MUST_BE_IN_AN_ORGANIZATION_TO_CREATE_A_ROLE: import("better-auth").RawError<"YOU_MUST_BE_IN_AN_ORGANIZATION_TO_CREATE_A_ROLE">;
            YOU_ARE_NOT_ALLOWED_TO_CREATE_A_ROLE: import("better-auth").RawError<"YOU_ARE_NOT_ALLOWED_TO_CREATE_A_ROLE">;
            YOU_ARE_NOT_ALLOWED_TO_UPDATE_A_ROLE: import("better-auth").RawError<"YOU_ARE_NOT_ALLOWED_TO_UPDATE_A_ROLE">;
            YOU_ARE_NOT_ALLOWED_TO_DELETE_A_ROLE: import("better-auth").RawError<"YOU_ARE_NOT_ALLOWED_TO_DELETE_A_ROLE">;
            YOU_ARE_NOT_ALLOWED_TO_READ_A_ROLE: import("better-auth").RawError<"YOU_ARE_NOT_ALLOWED_TO_READ_A_ROLE">;
            YOU_ARE_NOT_ALLOWED_TO_LIST_A_ROLE: import("better-auth").RawError<"YOU_ARE_NOT_ALLOWED_TO_LIST_A_ROLE">;
            YOU_ARE_NOT_ALLOWED_TO_GET_A_ROLE: import("better-auth").RawError<"YOU_ARE_NOT_ALLOWED_TO_GET_A_ROLE">;
            TOO_MANY_ROLES: import("better-auth").RawError<"TOO_MANY_ROLES">;
            INVALID_RESOURCE: import("better-auth").RawError<"INVALID_RESOURCE">;
            ROLE_NAME_IS_ALREADY_TAKEN: import("better-auth").RawError<"ROLE_NAME_IS_ALREADY_TAKEN">;
            CANNOT_DELETE_A_PRE_DEFINED_ROLE: import("better-auth").RawError<"CANNOT_DELETE_A_PRE_DEFINED_ROLE">;
            ROLE_IS_ASSIGNED_TO_MEMBERS: import("better-auth").RawError<"ROLE_IS_ASSIGNED_TO_MEMBERS">;
        };
        options: NoInfer<import("better-auth/plugins").OrganizationOptions & {
            teams: {
                enabled: true;
            };
            dynamicAccessControl?: {
                enabled?: false | undefined;
            } | undefined;
        }>;
    }, {
        id: "have-i-been-pwned";
        init(ctx: import("better-auth").AuthContext): {
            context: {
                password: {
                    hash(password: string): Promise<string>;
                    verify: (data: {
                        password: string;
                        hash: string;
                    }) => Promise<boolean>;
                    config: {
                        minPasswordLength: number;
                        maxPasswordLength: number;
                    };
                    checkPassword: (userId: string, ctx: import("better-auth").GenericEndpointContext<import("better-auth").BetterAuthOptions>) => Promise<boolean>;
                };
            };
        };
        options: import("better-auth/plugins").HaveIBeenPwnedOptions | undefined;
        $ERROR_CODES: {
            PASSWORD_COMPROMISED: import("better-auth").RawError<"PASSWORD_COMPROMISED">;
        };
    }, {
        id: "last-login-method";
        init(ctx: import("better-auth").AuthContext): {
            options: {
                databaseHooks: {
                    user: {
                        create: {
                            before(user: {
                                id: string;
                                createdAt: Date;
                                updatedAt: Date;
                                email: string;
                                emailVerified: boolean;
                                name: string;
                                image?: string | null | undefined;
                            } & Record<string, unknown>, context: import("better-auth").GenericEndpointContext | null): Promise<{
                                data: {
                                    lastLoginMethod: any;
                                    id: string;
                                    createdAt: Date;
                                    updatedAt: Date;
                                    email: string;
                                    emailVerified: boolean;
                                    name: string;
                                    image?: string | null | undefined;
                                };
                            } | undefined>;
                        };
                    };
                    session: {
                        create: {
                            after(session: {
                                id: string;
                                createdAt: Date;
                                updatedAt: Date;
                                userId: string;
                                expiresAt: Date;
                                token: string;
                                ipAddress?: string | null | undefined;
                                userAgent?: string | null | undefined;
                            } & Record<string, unknown>, context: import("better-auth").GenericEndpointContext | null): Promise<void>;
                        };
                    };
                };
            };
        };
        hooks: {
            after: {
                matcher(): true;
                handler: (inputContext: import("better-call").MiddlewareInputContext<import("better-call").MiddlewareOptions>) => Promise<void>;
            }[];
        };
        schema: undefined;
        options: NoInfer<import("better-auth/plugins").LastLoginMethodOptions>;
    }, {
        id: "api-key";
        $ERROR_CODES: {
            INVALID_METADATA_TYPE: import("better-auth").RawError<"INVALID_METADATA_TYPE">;
            REFILL_AMOUNT_AND_INTERVAL_REQUIRED: import("better-auth").RawError<"REFILL_AMOUNT_AND_INTERVAL_REQUIRED">;
            REFILL_INTERVAL_AND_AMOUNT_REQUIRED: import("better-auth").RawError<"REFILL_INTERVAL_AND_AMOUNT_REQUIRED">;
            USER_BANNED: import("better-auth").RawError<"USER_BANNED">;
            UNAUTHORIZED_SESSION: import("better-auth").RawError<"UNAUTHORIZED_SESSION">;
            KEY_NOT_FOUND: import("better-auth").RawError<"KEY_NOT_FOUND">;
            KEY_DISABLED: import("better-auth").RawError<"KEY_DISABLED">;
            KEY_EXPIRED: import("better-auth").RawError<"KEY_EXPIRED">;
            USAGE_EXCEEDED: import("better-auth").RawError<"USAGE_EXCEEDED">;
            KEY_NOT_RECOVERABLE: import("better-auth").RawError<"KEY_NOT_RECOVERABLE">;
            EXPIRES_IN_IS_TOO_SMALL: import("better-auth").RawError<"EXPIRES_IN_IS_TOO_SMALL">;
            EXPIRES_IN_IS_TOO_LARGE: import("better-auth").RawError<"EXPIRES_IN_IS_TOO_LARGE">;
            INVALID_REMAINING: import("better-auth").RawError<"INVALID_REMAINING">;
            INVALID_PREFIX_LENGTH: import("better-auth").RawError<"INVALID_PREFIX_LENGTH">;
            INVALID_NAME_LENGTH: import("better-auth").RawError<"INVALID_NAME_LENGTH">;
            METADATA_DISABLED: import("better-auth").RawError<"METADATA_DISABLED">;
            RATE_LIMIT_EXCEEDED: import("better-auth").RawError<"RATE_LIMIT_EXCEEDED">;
            NO_VALUES_TO_UPDATE: import("better-auth").RawError<"NO_VALUES_TO_UPDATE">;
            KEY_DISABLED_EXPIRATION: import("better-auth").RawError<"KEY_DISABLED_EXPIRATION">;
            INVALID_API_KEY: import("better-auth").RawError<"INVALID_API_KEY">;
            INVALID_USER_ID_FROM_API_KEY: import("better-auth").RawError<"INVALID_USER_ID_FROM_API_KEY">;
            INVALID_REFERENCE_ID_FROM_API_KEY: import("better-auth").RawError<"INVALID_REFERENCE_ID_FROM_API_KEY">;
            INVALID_API_KEY_GETTER_RETURN_TYPE: import("better-auth").RawError<"INVALID_API_KEY_GETTER_RETURN_TYPE">;
            SERVER_ONLY_PROPERTY: import("better-auth").RawError<"SERVER_ONLY_PROPERTY">;
            FAILED_TO_UPDATE_API_KEY: import("better-auth").RawError<"FAILED_TO_UPDATE_API_KEY">;
            NAME_REQUIRED: import("better-auth").RawError<"NAME_REQUIRED">;
            ORGANIZATION_ID_REQUIRED: import("better-auth").RawError<"ORGANIZATION_ID_REQUIRED">;
            USER_NOT_MEMBER_OF_ORGANIZATION: import("better-auth").RawError<"USER_NOT_MEMBER_OF_ORGANIZATION">;
            INSUFFICIENT_API_KEY_PERMISSIONS: import("better-auth").RawError<"INSUFFICIENT_API_KEY_PERMISSIONS">;
            NO_DEFAULT_API_KEY_CONFIGURATION_FOUND: import("better-auth").RawError<"NO_DEFAULT_API_KEY_CONFIGURATION_FOUND">;
            ORGANIZATION_PLUGIN_REQUIRED: import("better-auth").RawError<"ORGANIZATION_PLUGIN_REQUIRED">;
        };
        hooks: {
            before: {
                matcher: (ctx: import("better-auth").HookEndpointContext) => boolean;
                handler: (inputContext: import("better-call").MiddlewareInputContext<import("better-call").MiddlewareOptions>) => Promise<{
                    user: {
                        id: string;
                        createdAt: Date;
                        updatedAt: Date;
                        email: string;
                        emailVerified: boolean;
                        name: string;
                        image?: string | null | undefined;
                    };
                    session: {
                        id: string;
                        token: string;
                        userId: string;
                        userAgent: string | null;
                        ipAddress: string | null;
                        createdAt: Date;
                        updatedAt: Date;
                        expiresAt: Date;
                    };
                } | {
                    context: import("better-call").MiddlewareContext<import("better-call").MiddlewareOptions, {
                        returned?: unknown | undefined;
                        responseHeaders?: Headers | undefined;
                    } & import("better-auth").PluginContext<import("better-auth").BetterAuthOptions> & import("better-auth").InfoContext & {
                        options: import("better-auth").BetterAuthOptions;
                        trustedOrigins: string[];
                        trustedProviders: string[];
                        isTrustedOrigin: (url: string, settings?: {
                            allowRelativePaths: boolean;
                        }) => boolean;
                        oauthConfig: {
                            skipStateCookieCheck?: boolean | undefined;
                            storeStateStrategy: "database" | "cookie";
                        };
                        newSession: {
                            session: {
                                id: string;
                                createdAt: Date;
                                updatedAt: Date;
                                userId: string;
                                expiresAt: Date;
                                token: string;
                                ipAddress?: string | null | undefined;
                                userAgent?: string | null | undefined;
                            } & Record<string, any>;
                            user: {
                                id: string;
                                createdAt: Date;
                                updatedAt: Date;
                                email: string;
                                emailVerified: boolean;
                                name: string;
                                image?: string | null | undefined;
                            } & Record<string, any>;
                        } | null;
                        session: {
                            session: {
                                id: string;
                                createdAt: Date;
                                updatedAt: Date;
                                userId: string;
                                expiresAt: Date;
                                token: string;
                                ipAddress?: string | null | undefined;
                                userAgent?: string | null | undefined;
                            } & Record<string, any>;
                            user: {
                                id: string;
                                createdAt: Date;
                                updatedAt: Date;
                                email: string;
                                emailVerified: boolean;
                                name: string;
                                image?: string | null | undefined;
                            } & Record<string, any>;
                        } | null;
                        setNewSession: (session: {
                            session: {
                                id: string;
                                createdAt: Date;
                                updatedAt: Date;
                                userId: string;
                                expiresAt: Date;
                                token: string;
                                ipAddress?: string | null | undefined;
                                userAgent?: string | null | undefined;
                            } & Record<string, any>;
                            user: {
                                id: string;
                                createdAt: Date;
                                updatedAt: Date;
                                email: string;
                                emailVerified: boolean;
                                name: string;
                                image?: string | null | undefined;
                            } & Record<string, any>;
                        } | null) => void;
                        socialProviders: import("better-auth").OAuthProvider[];
                        authCookies: import("better-auth").BetterAuthCookies;
                        logger: ReturnType<(options?: import("better-auth").Logger | undefined) => import("better-auth").InternalLogger>;
                        rateLimit: {
                            enabled: boolean;
                            window: number;
                            max: number;
                            storage: "memory" | "database" | "secondary-storage";
                        } & Omit<import("better-auth").BetterAuthRateLimitOptions, "enabled" | "window" | "max" | "storage">;
                        adapter: import("better-auth").DBAdapter<import("better-auth").BetterAuthOptions>;
                        internalAdapter: import("better-auth").InternalAdapter<import("better-auth").BetterAuthOptions>;
                        createAuthCookie: (cookieName: string, overrideAttributes?: Partial<import("better-call").CookieOptions> | undefined) => import("better-auth").BetterAuthCookie;
                        secret: string;
                        secretConfig: string | import("better-auth").SecretConfig;
                        sessionConfig: {
                            updateAge: number;
                            expiresIn: number;
                            freshAge: number;
                            cookieRefreshCache: false | {
                                enabled: true;
                                updateAge: number;
                            };
                        };
                        generateId: (options: {
                            model: import("better-auth").ModelNames;
                            size?: number | undefined;
                        }) => string | false;
                        secondaryStorage: import("better-auth").SecondaryStorage | undefined;
                        password: {
                            hash: (password: string) => Promise<string>;
                            verify: (data: {
                                password: string;
                                hash: string;
                            }) => Promise<boolean>;
                            config: {
                                minPasswordLength: number;
                                maxPasswordLength: number;
                            };
                            checkPassword: (userId: string, ctx: import("better-auth").GenericEndpointContext<import("better-auth").BetterAuthOptions>) => Promise<boolean>;
                        };
                        tables: import("better-auth").BetterAuthDBSchema;
                        runMigrations: () => Promise<void>;
                        publishTelemetry: (event: {
                            type: string;
                            anonymousId?: string | undefined;
                            payload: Record<string, any>;
                        }) => Promise<void>;
                        skipOriginCheck: boolean | string[];
                        skipCSRFCheck: boolean;
                        runInBackground: (promise: Promise<unknown>) => void;
                        runInBackgroundOrAwait: (promise: Promise<unknown> | void) => import("better-auth").Awaitable<unknown>;
                    }>;
                }>;
            }[];
        };
        endpoints: {
            createApiKey: import("better-call").StrictEndpoint<"/api-key/create", {
                method: "POST";
                body: import("zod").ZodObject<{
                    configId: import("zod").ZodOptional<import("zod").ZodString>;
                    name: import("zod").ZodOptional<import("zod").ZodString>;
                    expiresIn: import("zod").ZodDefault<import("zod").ZodNullable<import("zod").ZodOptional<import("zod").ZodNumber>>>;
                    prefix: import("zod").ZodOptional<import("zod").ZodString>;
                    remaining: import("zod").ZodDefault<import("zod").ZodNullable<import("zod").ZodOptional<import("zod").ZodNumber>>>;
                    metadata: import("zod").ZodOptional<import("zod").ZodAny>;
                    refillAmount: import("zod").ZodOptional<import("zod").ZodNumber>;
                    refillInterval: import("zod").ZodOptional<import("zod").ZodNumber>;
                    rateLimitTimeWindow: import("zod").ZodOptional<import("zod").ZodNumber>;
                    rateLimitMax: import("zod").ZodOptional<import("zod").ZodNumber>;
                    rateLimitEnabled: import("zod").ZodOptional<import("zod").ZodBoolean>;
                    permissions: import("zod").ZodOptional<import("zod").ZodRecord<import("zod").ZodString, import("zod").ZodArray<import("zod").ZodString>>>;
                    userId: import("zod").ZodOptional<import("zod").ZodCoercedString<unknown>>;
                    organizationId: import("zod").ZodOptional<import("zod").ZodCoercedString<unknown>>;
                }, import("better-auth").$strip>;
                metadata: {
                    openapi: {
                        description: string;
                        responses: {
                            "200": {
                                description: string;
                                content: {
                                    "application/json": {
                                        schema: {
                                            type: "object";
                                            properties: {
                                                id: {
                                                    type: string;
                                                    description: string;
                                                };
                                                createdAt: {
                                                    type: string;
                                                    format: string;
                                                    description: string;
                                                };
                                                updatedAt: {
                                                    type: string;
                                                    format: string;
                                                    description: string;
                                                };
                                                name: {
                                                    type: string;
                                                    nullable: boolean;
                                                    description: string;
                                                };
                                                prefix: {
                                                    type: string;
                                                    nullable: boolean;
                                                    description: string;
                                                };
                                                start: {
                                                    type: string;
                                                    nullable: boolean;
                                                    description: string;
                                                };
                                                key: {
                                                    type: string;
                                                    description: string;
                                                };
                                                enabled: {
                                                    type: string;
                                                    description: string;
                                                };
                                                expiresAt: {
                                                    type: string;
                                                    format: string;
                                                    nullable: boolean;
                                                    description: string;
                                                };
                                                referenceId: {
                                                    type: string;
                                                    description: string;
                                                };
                                                lastRefillAt: {
                                                    type: string;
                                                    format: string;
                                                    nullable: boolean;
                                                    description: string;
                                                };
                                                lastRequest: {
                                                    type: string;
                                                    format: string;
                                                    nullable: boolean;
                                                    description: string;
                                                };
                                                metadata: {
                                                    type: string;
                                                    nullable: boolean;
                                                    additionalProperties: boolean;
                                                    description: string;
                                                };
                                                rateLimitMax: {
                                                    type: string;
                                                    nullable: boolean;
                                                    description: string;
                                                };
                                                rateLimitTimeWindow: {
                                                    type: string;
                                                    nullable: boolean;
                                                    description: string;
                                                };
                                                remaining: {
                                                    type: string;
                                                    nullable: boolean;
                                                    description: string;
                                                };
                                                refillAmount: {
                                                    type: string;
                                                    nullable: boolean;
                                                    description: string;
                                                };
                                                refillInterval: {
                                                    type: string;
                                                    nullable: boolean;
                                                    description: string;
                                                };
                                                rateLimitEnabled: {
                                                    type: string;
                                                    description: string;
                                                };
                                                requestCount: {
                                                    type: string;
                                                    description: string;
                                                };
                                                permissions: {
                                                    type: string;
                                                    nullable: boolean;
                                                    additionalProperties: {
                                                        type: string;
                                                        items: {
                                                            type: string;
                                                        };
                                                    };
                                                    description: string;
                                                };
                                            };
                                            required: string[];
                                        };
                                    };
                                };
                            };
                        };
                    };
                };
            }, {
                key: string;
                metadata: any;
                permissions: any;
                id: string;
                configId: string;
                name: string | null;
                start: string | null;
                prefix: string | null;
                referenceId: string;
                refillInterval: number | null;
                refillAmount: number | null;
                lastRefillAt: Date | null;
                enabled: boolean;
                rateLimitEnabled: boolean;
                rateLimitTimeWindow: number | null;
                rateLimitMax: number | null;
                requestCount: number;
                remaining: number | null;
                lastRequest: Date | null;
                expiresAt: Date | null;
                createdAt: Date;
                updatedAt: Date;
            }>;
            verifyApiKey: import("better-call").StrictEndpoint<string, {
                method: "POST";
                body: import("zod").ZodObject<{
                    configId: import("zod").ZodOptional<import("zod").ZodString>;
                    key: import("zod").ZodString;
                    permissions: import("zod").ZodOptional<import("zod").ZodRecord<import("zod").ZodString, import("zod").ZodArray<import("zod").ZodString>>>;
                }, import("better-auth").$strip>;
            }, {
                valid: boolean;
                error: {
                    message: import("better-auth").RawError<"INVALID_API_KEY">;
                    code: "KEY_NOT_FOUND";
                };
                key: null;
            } | {
                valid: boolean;
                error: {
                    message: string | undefined;
                    code: string;
                    cause?: unknown;
                };
                key: null;
            } | {
                valid: boolean;
                error: {
                    message: import("better-auth").RawError<"INVALID_API_KEY">;
                    code: "INVALID_API_KEY";
                };
                key: null;
            } | {
                valid: boolean;
                error: null;
                key: Omit<import("@better-auth/api-key").ApiKey, "key"> | null;
            }>;
            getApiKey: import("better-call").StrictEndpoint<"/api-key/get", {
                method: "GET";
                query: import("zod").ZodObject<{
                    configId: import("zod").ZodOptional<import("zod").ZodString>;
                    id: import("zod").ZodString;
                }, import("better-auth").$strip>;
                use: ((inputContext: import("better-call").MiddlewareInputContext<import("better-call").MiddlewareOptions>) => Promise<{
                    session: {
                        session: Record<string, any> & {
                            id: string;
                            createdAt: Date;
                            updatedAt: Date;
                            userId: string;
                            expiresAt: Date;
                            token: string;
                            ipAddress?: string | null | undefined;
                            userAgent?: string | null | undefined;
                        };
                        user: Record<string, any> & {
                            id: string;
                            createdAt: Date;
                            updatedAt: Date;
                            email: string;
                            emailVerified: boolean;
                            name: string;
                            image?: string | null | undefined;
                        };
                    };
                }>)[];
                metadata: {
                    openapi: {
                        description: string;
                        responses: {
                            "200": {
                                description: string;
                                content: {
                                    "application/json": {
                                        schema: {
                                            type: "object";
                                            properties: {
                                                id: {
                                                    type: string;
                                                    description: string;
                                                };
                                                name: {
                                                    type: string;
                                                    nullable: boolean;
                                                    description: string;
                                                };
                                                start: {
                                                    type: string;
                                                    nullable: boolean;
                                                    description: string;
                                                };
                                                prefix: {
                                                    type: string;
                                                    nullable: boolean;
                                                    description: string;
                                                };
                                                userId: {
                                                    type: string;
                                                    description: string;
                                                };
                                                refillInterval: {
                                                    type: string;
                                                    nullable: boolean;
                                                    description: string;
                                                };
                                                refillAmount: {
                                                    type: string;
                                                    nullable: boolean;
                                                    description: string;
                                                };
                                                lastRefillAt: {
                                                    type: string;
                                                    format: string;
                                                    nullable: boolean;
                                                    description: string;
                                                };
                                                enabled: {
                                                    type: string;
                                                    description: string;
                                                    default: boolean;
                                                };
                                                rateLimitEnabled: {
                                                    type: string;
                                                    description: string;
                                                };
                                                rateLimitTimeWindow: {
                                                    type: string;
                                                    nullable: boolean;
                                                    description: string;
                                                };
                                                rateLimitMax: {
                                                    type: string;
                                                    nullable: boolean;
                                                    description: string;
                                                };
                                                requestCount: {
                                                    type: string;
                                                    description: string;
                                                };
                                                remaining: {
                                                    type: string;
                                                    nullable: boolean;
                                                    description: string;
                                                };
                                                lastRequest: {
                                                    type: string;
                                                    format: string;
                                                    nullable: boolean;
                                                    description: string;
                                                };
                                                expiresAt: {
                                                    type: string;
                                                    format: string;
                                                    nullable: boolean;
                                                    description: string;
                                                };
                                                createdAt: {
                                                    type: string;
                                                    format: string;
                                                    description: string;
                                                };
                                                updatedAt: {
                                                    type: string;
                                                    format: string;
                                                    description: string;
                                                };
                                                metadata: {
                                                    type: string;
                                                    nullable: boolean;
                                                    additionalProperties: boolean;
                                                    description: string;
                                                };
                                                permissions: {
                                                    type: string;
                                                    nullable: boolean;
                                                    description: string;
                                                };
                                            };
                                            required: string[];
                                        };
                                    };
                                };
                            };
                        };
                    };
                };
            }, {
                metadata: Record<string, any> | null;
                permissions: {
                    [key: string]: string[];
                } | null;
                id: string;
                configId: string;
                name: string | null;
                start: string | null;
                prefix: string | null;
                referenceId: string;
                refillInterval: number | null;
                refillAmount: number | null;
                lastRefillAt: Date | null;
                enabled: boolean;
                rateLimitEnabled: boolean;
                rateLimitTimeWindow: number | null;
                rateLimitMax: number | null;
                requestCount: number;
                remaining: number | null;
                lastRequest: Date | null;
                expiresAt: Date | null;
                createdAt: Date;
                updatedAt: Date;
            }>;
            updateApiKey: import("better-call").StrictEndpoint<"/api-key/update", {
                method: "POST";
                body: import("zod").ZodObject<{
                    configId: import("zod").ZodOptional<import("zod").ZodString>;
                    keyId: import("zod").ZodString;
                    userId: import("zod").ZodOptional<import("zod").ZodCoercedString<unknown>>;
                    name: import("zod").ZodOptional<import("zod").ZodString>;
                    enabled: import("zod").ZodOptional<import("zod").ZodBoolean>;
                    remaining: import("zod").ZodOptional<import("zod").ZodNumber>;
                    refillAmount: import("zod").ZodOptional<import("zod").ZodNumber>;
                    refillInterval: import("zod").ZodOptional<import("zod").ZodNumber>;
                    metadata: import("zod").ZodOptional<import("zod").ZodAny>;
                    expiresIn: import("zod").ZodNullable<import("zod").ZodOptional<import("zod").ZodNumber>>;
                    rateLimitEnabled: import("zod").ZodOptional<import("zod").ZodBoolean>;
                    rateLimitTimeWindow: import("zod").ZodOptional<import("zod").ZodNumber>;
                    rateLimitMax: import("zod").ZodOptional<import("zod").ZodNumber>;
                    permissions: import("zod").ZodNullable<import("zod").ZodOptional<import("zod").ZodRecord<import("zod").ZodString, import("zod").ZodArray<import("zod").ZodString>>>>;
                }, import("better-auth").$strip>;
                metadata: {
                    openapi: {
                        description: string;
                        responses: {
                            "200": {
                                description: string;
                                content: {
                                    "application/json": {
                                        schema: {
                                            type: "object";
                                            properties: {
                                                id: {
                                                    type: string;
                                                    description: string;
                                                };
                                                name: {
                                                    type: string;
                                                    nullable: boolean;
                                                    description: string;
                                                };
                                                start: {
                                                    type: string;
                                                    nullable: boolean;
                                                    description: string;
                                                };
                                                prefix: {
                                                    type: string;
                                                    nullable: boolean;
                                                    description: string;
                                                };
                                                userId: {
                                                    type: string;
                                                    description: string;
                                                };
                                                refillInterval: {
                                                    type: string;
                                                    nullable: boolean;
                                                    description: string;
                                                };
                                                refillAmount: {
                                                    type: string;
                                                    nullable: boolean;
                                                    description: string;
                                                };
                                                lastRefillAt: {
                                                    type: string;
                                                    format: string;
                                                    nullable: boolean;
                                                    description: string;
                                                };
                                                enabled: {
                                                    type: string;
                                                    description: string;
                                                    default: boolean;
                                                };
                                                rateLimitEnabled: {
                                                    type: string;
                                                    description: string;
                                                };
                                                rateLimitTimeWindow: {
                                                    type: string;
                                                    nullable: boolean;
                                                    description: string;
                                                };
                                                rateLimitMax: {
                                                    type: string;
                                                    nullable: boolean;
                                                    description: string;
                                                };
                                                requestCount: {
                                                    type: string;
                                                    description: string;
                                                };
                                                remaining: {
                                                    type: string;
                                                    nullable: boolean;
                                                    description: string;
                                                };
                                                lastRequest: {
                                                    type: string;
                                                    format: string;
                                                    nullable: boolean;
                                                    description: string;
                                                };
                                                expiresAt: {
                                                    type: string;
                                                    format: string;
                                                    nullable: boolean;
                                                    description: string;
                                                };
                                                createdAt: {
                                                    type: string;
                                                    format: string;
                                                    description: string;
                                                };
                                                updatedAt: {
                                                    type: string;
                                                    format: string;
                                                    description: string;
                                                };
                                                metadata: {
                                                    type: string;
                                                    nullable: boolean;
                                                    additionalProperties: boolean;
                                                    description: string;
                                                };
                                                permissions: {
                                                    type: string;
                                                    nullable: boolean;
                                                    description: string;
                                                };
                                            };
                                            required: string[];
                                        };
                                    };
                                };
                            };
                        };
                    };
                };
            }, {
                metadata: Record<string, any> | null;
                permissions: {
                    [key: string]: string[];
                } | null;
                id: string;
                configId: string;
                name: string | null;
                start: string | null;
                prefix: string | null;
                referenceId: string;
                refillInterval: number | null;
                refillAmount: number | null;
                lastRefillAt: Date | null;
                enabled: boolean;
                rateLimitEnabled: boolean;
                rateLimitTimeWindow: number | null;
                rateLimitMax: number | null;
                requestCount: number;
                remaining: number | null;
                lastRequest: Date | null;
                expiresAt: Date | null;
                createdAt: Date;
                updatedAt: Date;
            }>;
            deleteApiKey: import("better-call").StrictEndpoint<"/api-key/delete", {
                method: "POST";
                body: import("zod").ZodObject<{
                    configId: import("zod").ZodOptional<import("zod").ZodString>;
                    keyId: import("zod").ZodString;
                }, import("better-auth").$strip>;
                use: ((inputContext: import("better-call").MiddlewareInputContext<import("better-call").MiddlewareOptions>) => Promise<{
                    session: {
                        session: Record<string, any> & {
                            id: string;
                            createdAt: Date;
                            updatedAt: Date;
                            userId: string;
                            expiresAt: Date;
                            token: string;
                            ipAddress?: string | null | undefined;
                            userAgent?: string | null | undefined;
                        };
                        user: Record<string, any> & {
                            id: string;
                            createdAt: Date;
                            updatedAt: Date;
                            email: string;
                            emailVerified: boolean;
                            name: string;
                            image?: string | null | undefined;
                        };
                    };
                }>)[];
                metadata: {
                    openapi: {
                        description: string;
                        requestBody: {
                            content: {
                                "application/json": {
                                    schema: {
                                        type: "object";
                                        properties: {
                                            keyId: {
                                                type: string;
                                                description: string;
                                            };
                                        };
                                        required: string[];
                                    };
                                };
                            };
                        };
                        responses: {
                            "200": {
                                description: string;
                                content: {
                                    "application/json": {
                                        schema: {
                                            type: "object";
                                            properties: {
                                                success: {
                                                    type: string;
                                                    description: string;
                                                };
                                            };
                                            required: string[];
                                        };
                                    };
                                };
                            };
                        };
                    };
                };
            }, {
                success: boolean;
            }>;
            listApiKeys: import("better-call").StrictEndpoint<"/api-key/list", {
                method: "GET";
                use: ((inputContext: import("better-call").MiddlewareInputContext<import("better-call").MiddlewareOptions>) => Promise<{
                    session: {
                        session: Record<string, any> & {
                            id: string;
                            createdAt: Date;
                            updatedAt: Date;
                            userId: string;
                            expiresAt: Date;
                            token: string;
                            ipAddress?: string | null | undefined;
                            userAgent?: string | null | undefined;
                        };
                        user: Record<string, any> & {
                            id: string;
                            createdAt: Date;
                            updatedAt: Date;
                            email: string;
                            emailVerified: boolean;
                            name: string;
                            image?: string | null | undefined;
                        };
                    };
                }>)[];
                query: import("zod").ZodOptional<import("zod").ZodObject<{
                    configId: import("zod").ZodOptional<import("zod").ZodString>;
                    organizationId: import("zod").ZodOptional<import("zod").ZodString>;
                    limit: import("zod").ZodOptional<import("zod").ZodCoercedNumber<unknown>>;
                    offset: import("zod").ZodOptional<import("zod").ZodCoercedNumber<unknown>>;
                    sortBy: import("zod").ZodOptional<import("zod").ZodString>;
                    sortDirection: import("zod").ZodOptional<import("zod").ZodEnum<{
                        asc: "asc";
                        desc: "desc";
                    }>>;
                }, import("better-auth").$strip>>;
                metadata: {
                    openapi: {
                        description: string;
                        responses: {
                            "200": {
                                description: string;
                                content: {
                                    "application/json": {
                                        schema: {
                                            type: "object";
                                            properties: {
                                                apiKeys: {
                                                    type: string;
                                                    items: {
                                                        type: string;
                                                        properties: {
                                                            id: {
                                                                type: string;
                                                                description: string;
                                                            };
                                                            name: {
                                                                type: string;
                                                                nullable: boolean;
                                                                description: string;
                                                            };
                                                            start: {
                                                                type: string;
                                                                nullable: boolean;
                                                                description: string;
                                                            };
                                                            prefix: {
                                                                type: string;
                                                                nullable: boolean;
                                                                description: string;
                                                            };
                                                            userId: {
                                                                type: string;
                                                                description: string;
                                                            };
                                                            refillInterval: {
                                                                type: string;
                                                                nullable: boolean;
                                                                description: string;
                                                            };
                                                            refillAmount: {
                                                                type: string;
                                                                nullable: boolean;
                                                                description: string;
                                                            };
                                                            lastRefillAt: {
                                                                type: string;
                                                                format: string;
                                                                nullable: boolean;
                                                                description: string;
                                                            };
                                                            enabled: {
                                                                type: string;
                                                                description: string;
                                                                default: boolean;
                                                            };
                                                            rateLimitEnabled: {
                                                                type: string;
                                                                description: string;
                                                            };
                                                            rateLimitTimeWindow: {
                                                                type: string;
                                                                nullable: boolean;
                                                                description: string;
                                                            };
                                                            rateLimitMax: {
                                                                type: string;
                                                                nullable: boolean;
                                                                description: string;
                                                            };
                                                            requestCount: {
                                                                type: string;
                                                                description: string;
                                                            };
                                                            remaining: {
                                                                type: string;
                                                                nullable: boolean;
                                                                description: string;
                                                            };
                                                            lastRequest: {
                                                                type: string;
                                                                format: string;
                                                                nullable: boolean;
                                                                description: string;
                                                            };
                                                            expiresAt: {
                                                                type: string;
                                                                format: string;
                                                                nullable: boolean;
                                                                description: string;
                                                            };
                                                            createdAt: {
                                                                type: string;
                                                                format: string;
                                                                description: string;
                                                            };
                                                            updatedAt: {
                                                                type: string;
                                                                format: string;
                                                                description: string;
                                                            };
                                                            metadata: {
                                                                type: string;
                                                                nullable: boolean;
                                                                additionalProperties: boolean;
                                                                description: string;
                                                            };
                                                            permissions: {
                                                                type: string;
                                                                nullable: boolean;
                                                                description: string;
                                                            };
                                                        };
                                                        required: string[];
                                                    };
                                                };
                                                total: {
                                                    type: string;
                                                    description: string;
                                                };
                                                limit: {
                                                    type: string;
                                                    nullable: boolean;
                                                    description: string;
                                                };
                                                offset: {
                                                    type: string;
                                                    nullable: boolean;
                                                    description: string;
                                                };
                                            };
                                            required: string[];
                                        };
                                    };
                                };
                            };
                        };
                    };
                };
            }, {
                apiKeys: {
                    metadata: Record<string, any> | null;
                    permissions: {
                        [key: string]: string[];
                    } | null;
                    id: string;
                    configId: string;
                    name: string | null;
                    start: string | null;
                    prefix: string | null;
                    referenceId: string;
                    refillInterval: number | null;
                    refillAmount: number | null;
                    lastRefillAt: Date | null;
                    enabled: boolean;
                    rateLimitEnabled: boolean;
                    rateLimitTimeWindow: number | null;
                    rateLimitMax: number | null;
                    requestCount: number;
                    remaining: number | null;
                    lastRequest: Date | null;
                    expiresAt: Date | null;
                    createdAt: Date;
                    updatedAt: Date;
                }[];
                total: number;
                limit: number | undefined;
                offset: number | undefined;
            }>;
            deleteAllExpiredApiKeys: import("better-call").StrictEndpoint<string, {
                method: "POST";
            }, {
                success: boolean;
                error: unknown;
            }>;
        };
        schema: {
            apikey: {
                fields: {
                    configId: {
                        type: "string";
                        required: true;
                        defaultValue: string;
                        input: false;
                        index: true;
                    };
                    name: {
                        type: "string";
                        required: false;
                        input: false;
                    };
                    start: {
                        type: "string";
                        required: false;
                        input: false;
                    };
                    referenceId: {
                        type: "string";
                        required: true;
                        input: false;
                        index: true;
                    };
                    prefix: {
                        type: "string";
                        required: false;
                        input: false;
                    };
                    key: {
                        type: "string";
                        required: true;
                        input: false;
                        index: true;
                    };
                    refillInterval: {
                        type: "number";
                        required: false;
                        input: false;
                    };
                    refillAmount: {
                        type: "number";
                        required: false;
                        input: false;
                    };
                    lastRefillAt: {
                        type: "date";
                        required: false;
                        input: false;
                    };
                    enabled: {
                        type: "boolean";
                        required: false;
                        input: false;
                        defaultValue: true;
                    };
                    rateLimitEnabled: {
                        type: "boolean";
                        required: false;
                        input: false;
                        defaultValue: true;
                    };
                    rateLimitTimeWindow: {
                        type: "number";
                        required: false;
                        input: false;
                        defaultValue: number;
                    };
                    rateLimitMax: {
                        type: "number";
                        required: false;
                        input: false;
                        defaultValue: number;
                    };
                    requestCount: {
                        type: "number";
                        required: false;
                        input: false;
                        defaultValue: number;
                    };
                    remaining: {
                        type: "number";
                        required: false;
                        input: false;
                    };
                    lastRequest: {
                        type: "date";
                        required: false;
                        input: false;
                    };
                    expiresAt: {
                        type: "date";
                        required: false;
                        input: false;
                    };
                    createdAt: {
                        type: "date";
                        required: true;
                        input: false;
                    };
                    updatedAt: {
                        type: "date";
                        required: true;
                        input: false;
                    };
                    permissions: {
                        type: "string";
                        required: false;
                        input: false;
                    };
                    metadata: {
                        type: "string";
                        required: false;
                        input: true;
                        transform: {
                            input(value: import("better-auth").DBPrimitive): string;
                            output(value: import("better-auth").DBPrimitive): any;
                        };
                    };
                };
            };
        };
        configurations: (import("@better-auth/api-key").ApiKeyConfigurationOptions & Required<Pick<import("@better-auth/api-key").ApiKeyConfigurationOptions, "rateLimit" | "storage" | "apiKeyHeaders" | "defaultKeyLength" | "keyExpiration" | "maximumPrefixLength" | "minimumPrefixLength" | "maximumNameLength" | "disableKeyHashing" | "minimumNameLength" | "requireName" | "enableMetadata" | "enableSessionForAPIKeys" | "startingCharactersConfig" | "fallbackToDatabase" | "deferUpdates">> & {
            keyExpiration: Required<NonNullable<import("@better-auth/api-key").ApiKeyConfigurationOptions["keyExpiration"]>>;
            startingCharactersConfig: Required<NonNullable<import("@better-auth/api-key").ApiKeyConfigurationOptions["startingCharactersConfig"]>>;
            rateLimit: Required<NonNullable<import("@better-auth/api-key").ApiKeyConfigurationOptions["rateLimit"]>>;
        })[];
    }];
    user: {
        modelName: "users";
        fields: {
            emailVerified: string;
            createdAt: string;
            updatedAt: string;
        };
    };
    session: {
        expiresIn: number;
        cookieCache: {
            enabled: true;
            maxAge: number;
        };
    };
    trustedOrigins: string[];
}>;
