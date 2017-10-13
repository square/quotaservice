import React from 'react';
import { render } from 'react-dom';
import { Provider } from 'react-redux';

import Root from './containers/Root.jsx';
import configureStore from './store.js';

require('../styles/application.scss')

const store = configureStore()

const env = {
  environment: process.env.NODE_ENV || 'development'
}

render(
  <Provider store={store}>
    <Root env={env} />
  </Provider>,
  document.getElementById('root')
)
