const path = require('path')
const webpack = require('webpack')
const VueLoaderPlugin = require('vue-loader/lib/plugin')
const HtmlWebpackPlugin = require('html-webpack-plugin')
const MiniCssExtractPlugin = require('mini-css-extract-plugin')
const UglifyJsPlugin = require('uglifyjs-webpack-plugin')
const FileManagerPlugin = require('filemanager-webpack-plugin')

function resolve(dir) {
  return path.resolve(__dirname, dir)
}

const distPath = resolve('dist')
const staticPath = resolve('../../static')

module.exports = (_, argv) => {
  const mode = argv.mode || process.env.NODE_ENV || 'development'
  const isProduction = mode === 'production'

  return {
    mode,
    context: resolve('.'),
    entry: {
      index: './src/index.js'
    },
    resolve: {
      extensions: ['.js', '.vue', '.json'],
      alias: {'vue$': 'vue/dist/vue.esm.js'}
    },
    output: {
      filename: 'static/js/[name].[hash].js',
      path: distPath
    },
    module: {
      rules: [{
        test: /\.vue$/,
        loader: 'vue-loader'
      }, {
        test: /\.(css|s[ac]ss)$/,
        use: [
          isProduction ? MiniCssExtractPlugin.loader : 'vue-style-loader',
          {loader: 'css-loader', options: {minimize: true}},
          'sass-loader'
        ]
      }]
    },
    plugins: [
      new VueLoaderPlugin(),
      new MiniCssExtractPlugin({
        filename: 'static/css/[name].[hash].css',
        chunkFilename: 'static/css/[name].[id].[hash].css'
      }),
      new HtmlWebpackPlugin({
        template: resolve('index.html'),
        inject: true,
        minify: {
          removeComments: true,
          collapseWhitespace: true,
          removeAttributeQuotes: true
        },
        chunksSortMode: 'dependency'
      }),
      new FileManagerPlugin({
        onStart: {
          delete: [`${staticPath}/js/`, `${staticPath}/css/`]
        },
        onEnd: {
          copy: [{source: `${distPath}/static/`, destination: staticPath}],
          delete: [`${distPath}/static/`]
        }
      }),
      new webpack.DefinePlugin({
        'process.env.NODE_ENV': JSON.stringify(mode)
      })
    ],
    optimization: {
      minimizer: [
        new UglifyJsPlugin({
          cache: true,
          parallel: true,
          uglifyOptions: {compress: {warnings: false}}
        })
      ]
    },
    externals: {
      'jquery': 'jQuery',
      'js-cookie': 'Cookies',
      'sortablejs': 'Sortable',
      'vue': 'Vue',
      'vue-meta': 'VueMeta',
      'vue-resource': 'VueResource',
      'vue-router': 'VueRouter'
    },
    node: {
      setImmediate: false,
      dgram: 'empty',
      fs: 'empty',
      net: 'empty',
      tls: 'empty',
      child_process: 'empty'
    }
  }
}
