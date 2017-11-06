const express = require('express');
const webpack = require('webpack');
const webpackDevServer = require('webpack-dev-server');
const config = require('./webpack.config');

const compiler = webpack(config);
const FIXTURES = `${__dirname}/__tests__/fixtures`;

const server = new webpackDevServer(compiler, {
  publicPath: config.output.publicPath,
  stats: {
    colors: true,
  },
  /* Enable this to use a QuotaService instance directly for development.
  proxy: {
    '/api': 'http://localhost:8080'
  }
  */
});

server.app.get('/api/capabilities', (req, res) =>
  res.sendFile(`${FIXTURES}/capabilities.json`)
);

server.app.get('/api/configs', (req, res) =>
  res.sendFile(`${FIXTURES}/configs.json`)
);

server.app.use(express.static(__dirname));

server.listen(3000, 'localhost', err => {
  if (err) {
    console.log(err);
  }

  console.log('👉 Running at http://localhost:3000');
  console.log('✋ configs & capabilities are served from ./__tests__/fixtures/')
});
