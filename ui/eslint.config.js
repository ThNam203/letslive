import { tanstackConfig } from "@tanstack/eslint-config";

/**
 * ESLint configuration with best practices for TanStack Start + React + TypeScript
 *
 * Best practices included:
 * - Import organization and ordering
 * - TypeScript strictness rules
 * - React hooks and component best practices
 * - File-specific rule overrides
 * - Path alias consistency (@/ imports)
 */
export default [
    ...tanstackConfig,
    {
        // Global ignores - files that ESLint should skip
        ignores: [
            "**/node_modules/**",
            "**/dist/**",
            "**/.output/**",
            "**/.next/**",
            "**/build/**",
            "**/coverage/**",
            "**/*.config.js",
            "**/*.config.ts",
            "**/routeTree.gen.ts", // Auto-generated TanStack Router file
        ],
    },
    {
        // General rules for all source files
        files: ["src/**/*.{ts,tsx}"],
        rules: {
            // Import organization: group and sort imports consistently
            "import/order": [
                "warn",
                {
                    groups: [
                        "builtin", // Node.js built-in modules
                        "external", // npm packages
                        "internal", // Internal modules (using path aliases)
                        ["parent", "sibling"], // Relative imports
                        "index", // Index imports
                        "type", // Type-only imports
                    ],
                    "newlines-between": "always",
                    alphabetize: {
                        order: "asc",
                        caseInsensitive: true,
                    },
                    pathGroups: [
                        {
                            pattern: "@/**",
                            group: "internal",
                            position: "before",
                        },
                    ],
                    pathGroupsExcludedImportTypes: ["builtin"],
                },
            ],

            // Prevent unused variables (except those prefixed with _)
            "@typescript-eslint/no-unused-vars": [
                "error",
                {
                    argsIgnorePattern: "^_",
                    varsIgnorePattern: "^_",
                    caughtErrorsIgnorePattern: "^_",
                },
            ],

            // Enforce consistent naming conventions
            "@typescript-eslint/naming-convention": [
                "warn",
                {
                    selector: "variableLike",
                    format: ["camelCase", "PascalCase", "UPPER_CASE"],
                    leadingUnderscore: "allow", // Allow _unused variables
                },
                {
                    selector: "typeLike",
                    format: ["PascalCase"],
                },
                {
                    selector: "interface",
                    format: ["PascalCase"],
                    custom: {
                        regex: "^I[A-Z]", // Prefer interfaces without 'I' prefix
                        match: false,
                    },
                },
            ],

            // TypeScript best practices
            "@typescript-eslint/prefer-as-const": "error",
            "@typescript-eslint/no-explicit-any": "warn",
            "@typescript-eslint/explicit-module-boundary-types": "off", // Can enable if preferred
            "@typescript-eslint/prefer-nullish-coalescing": "warn",
            "@typescript-eslint/prefer-optional-chain": "warn",

            // React best practices
            "react/react-in-jsx-scope": "off", // Not needed in React 17+
            "react/prop-types": "off", // TypeScript handles this
            "react-hooks/rules-of-hooks": "error",
            "react-hooks/exhaustive-deps": "warn",
            "react/function-component-definition": [
                "warn",
                {
                    namedComponents: "arrow-function",
                    unnamedComponents: "arrow-function",
                },
            ],

            // Code quality
            "no-console": [
                "warn",
                {
                    allow: ["warn", "error"], // Allow console.warn and console.error
                },
            ],
            "no-debugger": "error",
            "prefer-const": "error",
            "no-var": "error",
        },
    },
    {
        // Rules specific to route files (TanStack Start convention)
        files: ["src/routes/**/*.{ts,tsx}"],
        rules: {
            "import/no-default-export": "off", // Routes use default exports
        },
    },
    {
        // Rules for component files
        files: ["src/components/**/*.{ts,tsx}"],
        rules: {
            "import/prefer-default-export": "off", // Prefer named exports for components
        },
    },
    {
        // Rules for hooks
        files: ["src/hooks/**/*.{ts,tsx}"],
        rules: {
            // Hooks should start with 'use'
            "@typescript-eslint/naming-convention": [
                "warn",
                {
                    selector: "function",
                    format: ["camelCase"],
                    filter: {
                        regex: "^use",
                        match: true,
                    },
                },
            ],
        },
    },
    {
        // Rules for test files (if you add tests later)
        files: [
            "**/*.test.{ts,tsx}",
            "**/*.spec.{ts,tsx}",
            "**/__tests__/**/*.{ts,tsx}",
        ],
        rules: {
            "@typescript-eslint/no-explicit-any": "off",
            "no-console": "off",
        },
    },
    {
        // Rules for config and build files
        files: ["*.config.{js,ts}", "*.config.*.{js,ts}", "vite.config.ts"],
        rules: {
            "import/no-default-export": "off",
            "@typescript-eslint/no-var-requires": "off",
            "no-console": "off",
        },
    },
];
