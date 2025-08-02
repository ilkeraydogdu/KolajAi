const path = require('path');
const HtmlWebpackPlugin = require('html-webpack-plugin');
const MiniCssExtractPlugin = require('mini-css-extract-plugin');
const CssMinimizerPlugin = require('css-minimizer-webpack-plugin');
const TerserPlugin = require('terser-webpack-plugin');
const { BundleAnalyzerPlugin } = require('webpack-bundle-analyzer');
const { WorkboxPlugin } = require('workbox-webpack-plugin');

module.exports = (env, argv) => {
  const isProduction = argv.mode === 'production';
  const isDevelopment = !isProduction;

  return {
    entry: {
      // Main application bundle
      main: './web/static/js/main.js',
      
      // Admin panel bundle
      admin: './web/static/js/admin/admin.js',
      
      // AI features bundle
      ai: './web/static/js/ai/ai-main.js',
      
      // Marketplace bundle
      marketplace: './web/static/js/marketplace/marketplace.js',
      
      // Vendor bundle for third-party libraries
      vendor: [
        'alpinejs',
        'axios',
        'chart.js',
        'date-fns',
        'lodash'
      ]
    },

    output: {
      path: path.resolve(__dirname, 'dist'),
      filename: isProduction ? 'js/[name].[contenthash:8].js' : 'js/[name].js',
      chunkFilename: isProduction ? 'js/[name].[contenthash:8].chunk.js' : 'js/[name].chunk.js',
      publicPath: '/static/',
      clean: true,
    },

    resolve: {
      extensions: ['.js', '.ts', '.json'],
      alias: {
        '@': path.resolve(__dirname, 'web/static/js'),
        '@components': path.resolve(__dirname, 'web/static/js/components'),
        '@utils': path.resolve(__dirname, 'web/static/js/utils'),
        '@services': path.resolve(__dirname, 'web/static/js/services'),
        '@styles': path.resolve(__dirname, 'web/static/css'),
      }
    },

    module: {
      rules: [
        // JavaScript/TypeScript
        {
          test: /\.(js|ts)$/,
          exclude: /node_modules/,
          use: {
            loader: 'babel-loader',
            options: {
              presets: [
                ['@babel/preset-env', {
                  targets: {
                    browsers: ['> 1%', 'last 2 versions', 'not dead', 'not ie 11']
                  },
                  useBuiltIns: 'usage',
                  corejs: 3
                }],
                '@babel/preset-typescript'
              ],
              plugins: [
                '@babel/plugin-proposal-class-properties',
                '@babel/plugin-proposal-optional-chaining',
                '@babel/plugin-proposal-nullish-coalescing-operator'
              ]
            }
          }
        },

        // CSS/SCSS
        {
          test: /\.(css|scss|sass)$/,
          use: [
            isDevelopment ? 'style-loader' : MiniCssExtractPlugin.loader,
            {
              loader: 'css-loader',
              options: {
                importLoaders: 2,
                sourceMap: isDevelopment
              }
            },
            {
              loader: 'postcss-loader',
              options: {
                postcssOptions: {
                  plugins: [
                    ['postcss-preset-env', {
                      autoprefixer: {
                        flexbox: 'no-2009'
                      },
                      stage: 3
                    }]
                  ]
                },
                sourceMap: isDevelopment
              }
            },
            {
              loader: 'sass-loader',
              options: {
                sourceMap: isDevelopment
              }
            }
          ]
        },

        // Images
        {
          test: /\.(png|jpe?g|gif|svg|webp)$/i,
          type: 'asset',
          parser: {
            dataUrlCondition: {
              maxSize: 8 * 1024 // 8kb
            }
          },
          generator: {
            filename: 'images/[name].[contenthash:8][ext]'
          }
        },

        // Fonts
        {
          test: /\.(woff|woff2|eot|ttf|otf)$/i,
          type: 'asset/resource',
          generator: {
            filename: 'fonts/[name].[contenthash:8][ext]'
          }
        }
      ]
    },

    plugins: [
      // Extract CSS into separate files
      new MiniCssExtractPlugin({
        filename: isProduction ? 'css/[name].[contenthash:8].css' : 'css/[name].css',
        chunkFilename: isProduction ? 'css/[name].[contenthash:8].chunk.css' : 'css/[name].chunk.css'
      }),

      // Bundle analyzer (only in analyze mode)
      ...(process.env.ANALYZE ? [new BundleAnalyzerPlugin()] : []),

      // Service Worker for PWA (production only)
      ...(isProduction ? [
        new WorkboxPlugin.GenerateSW({
          clientsClaim: true,
          skipWaiting: true,
          exclude: [/\.map$/, /manifest$/, /\.htaccess$/],
          runtimeCaching: [
            {
              urlPattern: /^https:\/\/fonts\.googleapis\.com/,
              handler: 'StaleWhileRevalidate',
              options: {
                cacheName: 'google-fonts-stylesheets',
              },
            },
            {
              urlPattern: /^https:\/\/fonts\.gstatic\.com/,
              handler: 'CacheFirst',
              options: {
                cacheName: 'google-fonts-webfonts',
                expiration: {
                  maxEntries: 30,
                  maxAgeSeconds: 60 * 60 * 24 * 365, // 1 year
                },
              },
            },
            {
              urlPattern: /\.(?:png|jpg|jpeg|svg|gif|webp)$/,
              handler: 'CacheFirst',
              options: {
                cacheName: 'images',
                expiration: {
                  maxEntries: 100,
                  maxAgeSeconds: 60 * 60 * 24 * 30, // 30 days
                },
              },
            },
            {
              urlPattern: /^https:\/\/api\.kolaj\.ai/,
              handler: 'NetworkFirst',
              options: {
                cacheName: 'api-cache',
                expiration: {
                  maxEntries: 50,
                  maxAgeSeconds: 60 * 5, // 5 minutes
                },
                cacheableResponse: {
                  statuses: [0, 200],
                },
              },
            }
          ]
        })
      ] : [])
    ],

    optimization: {
      minimize: isProduction,
      minimizer: [
        new TerserPlugin({
          terserOptions: {
            parse: {
              ecma: 8,
            },
            compress: {
              ecma: 5,
              warnings: false,
              comparisons: false,
              inline: 2,
              drop_console: isProduction,
              drop_debugger: isProduction,
            },
            mangle: {
              safari10: true,
            },
            output: {
              ecma: 5,
              comments: false,
              ascii_only: true,
            },
          },
        }),
        new CssMinimizerPlugin()
      ],

      // Code splitting
      splitChunks: {
        chunks: 'all',
        cacheGroups: {
          // Vendor chunk for third-party libraries
          vendor: {
            test: /[\\/]node_modules[\\/]/,
            name: 'vendors',
            priority: 10,
            reuseExistingChunk: true,
          },
          
          // Common chunk for shared code
          common: {
            name: 'common',
            minChunks: 2,
            priority: 5,
            reuseExistingChunk: true,
          },
          
          // CSS chunks
          styles: {
            name: 'styles',
            type: 'css/mini-extract',
            chunks: 'all',
            enforce: true,
          }
        }
      },

      // Runtime chunk
      runtimeChunk: {
        name: 'runtime'
      }
    },

    // Development server
    devServer: {
      static: {
        directory: path.join(__dirname, 'dist'),
      },
      compress: true,
      port: 9000,
      hot: true,
      open: true,
      historyApiFallback: true,
      proxy: {
        '/api': {
          target: 'http://localhost:8080',
          changeOrigin: true,
          secure: false
        }
      }
    },

    // Source maps
    devtool: isDevelopment ? 'eval-source-map' : 'source-map',

    // Performance hints
    performance: {
      hints: isProduction ? 'warning' : false,
      maxEntrypointSize: 512000,
      maxAssetSize: 512000
    },

    // Stats
    stats: {
      colors: true,
      modules: false,
      children: false,
      chunks: false,
      chunkModules: false
    }
  };
};