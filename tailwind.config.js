/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    './web/templates/**/*.{html,gohtml}',
    './web/static/js/**/*.{js,ts}',
    './web/static/css/**/*.{css,scss}',
  ],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        // Brand colors
        brand: {
          50: '#eff6ff',
          100: '#dbeafe',
          200: '#bfdbfe',
          300: '#93c5fd',
          400: '#60a5fa',
          500: '#3b82f6',
          600: '#2563eb',
          700: '#1d4ed8',
          800: '#1e40af',
          900: '#1e3a8a',
          950: '#172554',
        },
        // Semantic colors
        success: {
          50: '#f0fdf4',
          100: '#dcfce7',
          200: '#bbf7d0',
          300: '#86efac',
          400: '#4ade80',
          500: '#22c55e',
          600: '#16a34a',
          700: '#15803d',
          800: '#166534',
          900: '#14532d',
          950: '#052e16',
        },
        warning: {
          50: '#fffbeb',
          100: '#fef3c7',
          200: '#fde68a',
          300: '#fcd34d',
          400: '#fbbf24',
          500: '#f59e0b',
          600: '#d97706',
          700: '#b45309',
          800: '#92400e',
          900: '#78350f',
          950: '#451a03',
        },
        danger: {
          50: '#fef2f2',
          100: '#fee2e2',
          200: '#fecaca',
          300: '#fca5a5',
          400: '#f87171',
          500: '#ef4444',
          600: '#dc2626',
          700: '#b91c1c',
          800: '#991b1b',
          900: '#7f1d1d',
          950: '#450a0a',
        },
        // Neutral colors with better contrast
        neutral: {
          50: '#fafafa',
          100: '#f5f5f5',
          200: '#e5e5e5',
          300: '#d4d4d4',
          400: '#a3a3a3',
          500: '#737373',
          600: '#525252',
          700: '#404040',
          800: '#262626',
          900: '#171717',
          950: '#0a0a0a',
        }
      },
      fontFamily: {
        sans: [
          'Inter',
          'ui-sans-serif',
          'system-ui',
          '-apple-system',
          'BlinkMacSystemFont',
          '"Segoe UI"',
          'Roboto',
          '"Helvetica Neue"',
          'Arial',
          '"Noto Sans"',
          'sans-serif',
          '"Apple Color Emoji"',
          '"Segoe UI Emoji"',
          '"Segoe UI Symbol"',
          '"Noto Color Emoji"',
        ],
        serif: [
          'ui-serif',
          'Georgia',
          'Cambria',
          '"Times New Roman"',
          'Times',
          'serif',
        ],
        mono: [
          'JetBrains Mono',
          'ui-monospace',
          'SFMono-Regular',
          '"Menlo"',
          'Monaco',
          'Consolas',
          '"Liberation Mono"',
          '"Courier New"',
          'monospace',
        ],
      },
      fontSize: {
        '2xs': ['0.625rem', { lineHeight: '0.75rem' }],
        xs: ['0.75rem', { lineHeight: '1rem' }],
        sm: ['0.875rem', { lineHeight: '1.25rem' }],
        base: ['1rem', { lineHeight: '1.5rem' }],
        lg: ['1.125rem', { lineHeight: '1.75rem' }],
        xl: ['1.25rem', { lineHeight: '1.75rem' }],
        '2xl': ['1.5rem', { lineHeight: '2rem' }],
        '3xl': ['1.875rem', { lineHeight: '2.25rem' }],
        '4xl': ['2.25rem', { lineHeight: '2.5rem' }],
        '5xl': ['3rem', { lineHeight: '1' }],
        '6xl': ['3.75rem', { lineHeight: '1' }],
        '7xl': ['4.5rem', { lineHeight: '1' }],
        '8xl': ['6rem', { lineHeight: '1' }],
        '9xl': ['8rem', { lineHeight: '1' }],
      },
      spacing: {
        '18': '4.5rem',
        '88': '22rem',
        '128': '32rem',
        '144': '36rem',
      },
      maxWidth: {
        '8xl': '88rem',
        '9xl': '96rem',
      },
      minHeight: {
        '12': '3rem',
        '16': '4rem',
        '20': '5rem',
        '24': '6rem',
      },
      borderRadius: {
        '4xl': '2rem',
        '5xl': '2.5rem',
      },
      boxShadow: {
        'inner-lg': 'inset 0 4px 8px 0 rgba(0, 0, 0, 0.1)',
        'glow': '0 0 20px rgba(59, 130, 246, 0.5)',
        'glow-lg': '0 0 40px rgba(59, 130, 246, 0.5)',
      },
      dropShadow: {
        'glow': '0 0 10px rgba(59, 130, 246, 0.5)',
        'glow-lg': '0 0 20px rgba(59, 130, 246, 0.5)',
      },
      animation: {
        'fade-in': 'fadeIn 0.5s ease-in-out',
        'fade-out': 'fadeOut 0.5s ease-in-out',
        'slide-in-up': 'slideInUp 0.3s ease-out',
        'slide-in-down': 'slideInDown 0.3s ease-out',
        'slide-in-left': 'slideInLeft 0.3s ease-out',
        'slide-in-right': 'slideInRight 0.3s ease-out',
        'bounce-in': 'bounceIn 0.6s ease-out',
        'shake': 'shake 0.5s ease-in-out',
        'pulse-slow': 'pulse 3s ease-in-out infinite',
        'spin-slow': 'spin 3s linear infinite',
        'ping-slow': 'ping 3s cubic-bezier(0, 0, 0.2, 1) infinite',
        'float': 'float 3s ease-in-out infinite',
        'gradient': 'gradient 8s ease infinite',
      },
      keyframes: {
        fadeIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' },
        },
        fadeOut: {
          '0%': { opacity: '1' },
          '100%': { opacity: '0' },
        },
        slideInUp: {
          '0%': { transform: 'translateY(100%)', opacity: '0' },
          '100%': { transform: 'translateY(0)', opacity: '1' },
        },
        slideInDown: {
          '0%': { transform: 'translateY(-100%)', opacity: '0' },
          '100%': { transform: 'translateY(0)', opacity: '1' },
        },
        slideInLeft: {
          '0%': { transform: 'translateX(-100%)', opacity: '0' },
          '100%': { transform: 'translateX(0)', opacity: '1' },
        },
        slideInRight: {
          '0%': { transform: 'translateX(100%)', opacity: '0' },
          '100%': { transform: 'translateX(0)', opacity: '1' },
        },
        bounceIn: {
          '0%': { transform: 'scale(0.3)', opacity: '0' },
          '50%': { transform: 'scale(1.05)' },
          '70%': { transform: 'scale(0.9)' },
          '100%': { transform: 'scale(1)', opacity: '1' },
        },
        shake: {
          '0%, 100%': { transform: 'translateX(0)' },
          '10%, 30%, 50%, 70%, 90%': { transform: 'translateX(-10px)' },
          '20%, 40%, 60%, 80%': { transform: 'translateX(10px)' },
        },
        float: {
          '0%, 100%': { transform: 'translateY(0px)' },
          '50%': { transform: 'translateY(-20px)' },
        },
        gradient: {
          '0%, 100%': { backgroundPosition: '0% 50%' },
          '50%': { backgroundPosition: '100% 50%' },
        },
      },
      backgroundImage: {
        'gradient-radial': 'radial-gradient(var(--tw-gradient-stops))',
        'gradient-conic': 'conic-gradient(from 180deg at 50% 50%, var(--tw-gradient-stops))',
        'gradient-mesh': 'linear-gradient(45deg, var(--tw-gradient-stops))',
      },
      backdropBlur: {
        xs: '2px',
      },
      zIndex: {
        '60': '60',
        '70': '70',
        '80': '80',
        '90': '90',
        '100': '100',
      },
      screens: {
        'xs': '475px',
        '3xl': '1600px',
      },
      aspectRatio: {
        '4/3': '4 / 3',
        '3/2': '3 / 2',
        '2/3': '2 / 3',
        '9/16': '9 / 16',
      },
      typography: (theme) => ({
        DEFAULT: {
          css: {
            color: theme('colors.gray.700'),
            maxWidth: 'none',
            a: {
              color: theme('colors.brand.600'),
              textDecoration: 'none',
              fontWeight: '500',
              '&:hover': {
                color: theme('colors.brand.700'),
                textDecoration: 'underline',
              },
            },
            'h1, h2, h3, h4, h5, h6': {
              color: theme('colors.gray.900'),
              fontWeight: '700',
            },
            code: {
              color: theme('colors.brand.600'),
              backgroundColor: theme('colors.gray.100'),
              paddingLeft: theme('spacing.1'),
              paddingRight: theme('spacing.1'),
              paddingTop: theme('spacing.0.5'),
              paddingBottom: theme('spacing.0.5'),
              borderRadius: theme('borderRadius.md'),
              fontWeight: '500',
            },
            'code::before': {
              content: '""',
            },
            'code::after': {
              content: '""',
            },
            pre: {
              backgroundColor: theme('colors.gray.900'),
              color: theme('colors.gray.100'),
            },
            'pre code': {
              backgroundColor: 'transparent',
              color: 'inherit',
            },
            blockquote: {
              borderLeftColor: theme('colors.brand.500'),
              color: theme('colors.gray.700'),
            },
          },
        },
        dark: {
          css: {
            color: theme('colors.gray.300'),
            a: {
              color: theme('colors.brand.400'),
              '&:hover': {
                color: theme('colors.brand.300'),
              },
            },
            'h1, h2, h3, h4, h5, h6': {
              color: theme('colors.gray.100'),
            },
            code: {
              color: theme('colors.brand.400'),
              backgroundColor: theme('colors.gray.800'),
            },
            blockquote: {
              borderLeftColor: theme('colors.brand.400'),
              color: theme('colors.gray.300'),
            },
            'thead th': {
              color: theme('colors.gray.100'),
            },
            'tbody td': {
              borderBottomColor: theme('colors.gray.700'),
            },
          },
        },
      }),
    },
  },
  plugins: [
    require('@tailwindcss/forms'),
    require('@tailwindcss/typography'),
    require('@tailwindcss/aspect-ratio'),
    require('@tailwindcss/container-queries'),
    
    // Custom plugins
    function({ addUtilities, addComponents, theme }) {
      // Custom utilities
      addUtilities({
        '.text-balance': {
          'text-wrap': 'balance',
        },
        '.text-pretty': {
          'text-wrap': 'pretty',
        },
        '.scrollbar-none': {
          'scrollbar-width': 'none',
          '-ms-overflow-style': 'none',
          '&::-webkit-scrollbar': {
            display: 'none',
          },
        },
        '.scrollbar-thin': {
          'scrollbar-width': 'thin',
          'scrollbar-color': `${theme('colors.gray.400')} ${theme('colors.gray.100')}`,
          '&::-webkit-scrollbar': {
            width: '8px',
            height: '8px',
          },
          '&::-webkit-scrollbar-track': {
            background: theme('colors.gray.100'),
            'border-radius': '4px',
          },
          '&::-webkit-scrollbar-thumb': {
            background: theme('colors.gray.400'),
            'border-radius': '4px',
            '&:hover': {
              background: theme('colors.gray.500'),
            },
          },
        },
        '.safe-area': {
          'padding-top': 'env(safe-area-inset-top)',
          'padding-right': 'env(safe-area-inset-right)',
          'padding-bottom': 'env(safe-area-inset-bottom)',
          'padding-left': 'env(safe-area-inset-left)',
        },
        '.safe-area-top': {
          'padding-top': 'env(safe-area-inset-top)',
        },
        '.safe-area-bottom': {
          'padding-bottom': 'env(safe-area-inset-bottom)',
        },
      });

      // Custom components
      addComponents({
        '.btn': {
          display: 'inline-flex',
          'align-items': 'center',
          'justify-content': 'center',
          padding: `${theme('spacing.2')} ${theme('spacing.4')}`,
          'font-size': theme('fontSize.sm[0]'),
          'line-height': theme('fontSize.sm[1].lineHeight'),
          'font-weight': theme('fontWeight.medium'),
          'border-radius': theme('borderRadius.md'),
          border: `1px solid transparent`,
          'box-shadow': theme('boxShadow.sm'),
          transition: 'all 0.2s ease-in-out',
          cursor: 'pointer',
          '&:focus': {
            outline: '2px solid transparent',
            'outline-offset': '2px',
            'box-shadow': `0 0 0 2px ${theme('colors.brand.500')}`,
          },
          '&:disabled': {
            opacity: '0.5',
            cursor: 'not-allowed',
          },
        },
        '.btn-primary': {
          'background-color': theme('colors.brand.600'),
          color: theme('colors.white'),
          '&:hover:not(:disabled)': {
            'background-color': theme('colors.brand.700'),
          },
        },
        '.btn-secondary': {
          'background-color': theme('colors.gray.600'),
          color: theme('colors.white'),
          '&:hover:not(:disabled)': {
            'background-color': theme('colors.gray.700'),
          },
        },
        '.btn-outline': {
          'background-color': 'transparent',
          'border-color': theme('colors.gray.300'),
          color: theme('colors.gray.700'),
          '&:hover:not(:disabled)': {
            'background-color': theme('colors.gray.50'),
          },
        },
        '.card': {
          'background-color': theme('colors.white'),
          'border-radius': theme('borderRadius.lg'),
          'box-shadow': theme('boxShadow.sm'),
          border: `1px solid ${theme('colors.gray.200')}`,
          overflow: 'hidden',
        },
        '.input': {
          display: 'block',
          width: '100%',
          padding: `${theme('spacing.2')} ${theme('spacing.3')}`,
          'font-size': theme('fontSize.sm[0]'),
          'line-height': theme('fontSize.sm[1].lineHeight'),
          'border-radius': theme('borderRadius.md'),
          border: `1px solid ${theme('colors.gray.300')}`,
          'box-shadow': theme('boxShadow.sm'),
          '&:focus': {
            outline: '2px solid transparent',
            'outline-offset': '2px',
            'border-color': theme('colors.brand.500'),
            'box-shadow': `0 0 0 1px ${theme('colors.brand.500')}`,
          },
        },
      });
    },
  ],
};