var webpack = require('webpack');
var WebpackDevServer = require('webpack-dev-server');
var config = require('./webpack.config');

new WebpackDevServer(webpack(config), {
  publicPath: config.output.publicPath,
  proxy: {
    '/api': 'http://localhost:8080'
  }
}).listen(3000, 'localhost', function(err, result) {
  if (err) {
    console.log(err);
  }

  console.log('Running at http://localhost:3000');
});
