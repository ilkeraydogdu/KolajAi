module.exports = {
  env: {
    browser: true,
    es2021: true,
    node: true,
  },
  extends: [
    'eslint:recommended',
    'prettier'
  ],
  parserOptions: {
    ecmaVersion: 12,
    sourceType: 'module',
  },
  rules: {
    'no-console': 'warn',
    'no-unused-vars': 'warn',
    'no-undef': 'error',
    'prefer-const': 'error',
    'no-var': 'error',
  },
  globals: {
    Alpine: 'readonly',
    axios: 'readonly',
    Lobibox: 'readonly',
    feather: 'readonly',
    Chart: 'readonly',
    KolajAI: 'writable',
    $: 'readonly',
    jQuery: 'readonly'
  }
};