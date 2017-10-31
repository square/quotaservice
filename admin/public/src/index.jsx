import 'babel-polyfill';

import React from 'react';
import { render } from 'react-dom';
import { Provider } from 'react-redux';

import Root from './containers/Root.jsx';
import configureStore from './store.js';

require('../styles/application.scss')

const store = configureStore()

const env = {
  environment: process.env.NODE_ENV || 'development',
  capabilities: true
}


// Capabilities response for testing

let capabilities;

function transformRawCapabilities(rawCapabilities) {
  const capabilities = {};

  Object.keys(rawCapabilities).forEach(namespaceName => {
    // splitting by : and taking first part allows complex names spaces to match application names
    namespaceName = namespaceName.split(/:/)[0];
    capabilities[namespaceName] = rawCapabilities[namespaceName].find(group => ['deployers', 'owners'].indexOf(group) !== -1) !== undefined;
  });

  return capabilities;
}

window.addEventListener('QuotaService.fetchCapabilities', e => {
  capabilities = fetch('/api/capabilities')
    .then(response => response.json())
    .then(transformRawCapabilities);

  capabilities.then(e.detail.callback);
});

window.addEventListener('QuotaService.getCapabilities', e => {
  const { callback } = e.detail;
  let { namespaceName } = e.detail;

  if (!capabilities) {
    callback(false);
    return;
  }

  namespaceName = namespaceName.split(/:/)[0];
  capabilities.then(capabilities => callback(capabilities[namespaceName]));
});

render(
  <Provider store={store}>
    <Root env={env} />
  </Provider>,
  document.getElementById('root')
)