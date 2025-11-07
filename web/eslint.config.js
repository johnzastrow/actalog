import js from '@eslint/js'
import vue from 'eslint-plugin-vue'
import prettier from 'eslint-config-prettier'
import globals from 'globals'

export default [
  // Ignore patterns
  {
    ignores: ['dist/**', 'node_modules/**', '.cache/**', 'coverage/**']
  },

  // Base JavaScript/ES configuration
  js.configs.recommended,

  // Vue recommended configuration
  ...vue.configs['flat/recommended'],

  // Global configuration
  {
    files: ['**/*.{js,mjs,cjs,vue}'],
    languageOptions: {
      ecmaVersion: 'latest',
      sourceType: 'module',
      globals: {
        ...globals.browser,
        ...globals.node,
        ...globals.es2021
      }
    },
    rules: {
      // Vue.js specific rules
      'vue/multi-word-component-names': 'warn',
      'vue/no-unused-vars': 'error',
      'vue/require-default-prop': 'warn',
      'vue/require-prop-types': 'warn',
      'vue/component-name-in-template-casing': ['error', 'kebab-case'],
      'vue/html-self-closing': [
        'error',
        {
          html: {
            void: 'always',
            normal: 'always',
            component: 'always'
          }
        }
      ],

      // General JavaScript rules
      'no-console': process.env.NODE_ENV === 'production' ? 'warn' : 'off',
      'no-debugger': process.env.NODE_ENV === 'production' ? 'error' : 'off',
      'no-unused-vars': ['error', { argsIgnorePattern: '^_' }],
      'prefer-const': 'error',
      'no-var': 'error',
      eqeqeq: ['error', 'always'],
      curly: ['error', 'all'],
      'brace-style': ['error', '1tbs']
    }
  },

  // Prettier config (must be last to override other configs)
  prettier
]
