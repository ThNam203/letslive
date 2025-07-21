/**
 * @see https://prettier.io/docs/configuration
 * @type {import("prettier").Config}
 */

const config = {
    tabWidth: 4,
    semi: true,
    singleQuote: false,
    plugins: ["prettier-plugin-tailwindcss"],
};
  
module.exports = config;