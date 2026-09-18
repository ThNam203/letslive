import next from "eslint-config-next/core-web-vitals";

const config = [
    { ignores: ["public/mockServiceWorker.js"] },
    ...next,
    {
        files: ["**/*.ts", "**/*.tsx"],
        rules: {
            // `any` switches type checking off exactly where the data is least
            // trustworthy — parsed request bodies, caught errors, API
            // envelopes. Use `unknown` and narrow it.
            "@typescript-eslint/no-explicit-any": "error",
        },
    },
];

export default config;
