const express = require('express');
const webpack = require('webpack');
const webpackDevMiddleware = require('webpack-dev-middleware');
const config = require('./webpack.config');

const app = express();
const compiler = webpack(config);
const FIXTURES = `${__dirname}/__tests__/fixtures`;

app.get('/api/capabilities', (req, res) =>
  res.sendFile(`${FIXTURES}/capabilities.json`)
);

app.get('/api/configs', (req, res) =>
  res.sendFile(`${FIXTURES}/configs.json`)
);

app.use(webpackDevMiddleware(compiler, {
  publicPath: config.output.publicPath,
  hot: false,
  stats: {
    colors: true,
  },
  proxy: {
    '/api': 'http://localhost:8080'
  }
}));

app.use(express.static(__dirname));

app.listen(3000, 'localhost', err => {
  if (err) {
    console.log(err);
  }

  console.log('👉 Running at http://localhost:3000');
  console.log('✋ configs & capabilities are served from ./__tests__/fixtures/')
});