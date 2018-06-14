const path = require('path');
const webpack = require('webpack');

module.exports = {
  mode: 'production',
  entry: [
    './src/index.jsx'
  ],
  bail: true,
  profile: false,
  devtool: false,
  output: {
    path: path.join(__dirname, 'js'),
    filename: 'bundle.js'
  },
  optimization: {
    minimize: true
  },
  plugins: [
    new webpack.DefinePlugin({
      'process.env': {
        NODE_ENV: JSON.stringify('production')
      }
    })
  ],
  module: {
    rules: [
      {
        test: /\.jsx?$/,
        loader: 'babel-loader',
        include: path.join(__dirname, 'src')
      },
      {
        test: /\.scss$/,
        loaders: ['style-loader', 'css-loader', 'sass-loader']
      }
    ]
  }
}
