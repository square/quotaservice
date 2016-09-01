require('../styles/application.scss')

import React from 'react'
import { render } from 'react-dom'
import { Provider } from 'react-redux'

import configureStore from './store.js'
import QuotaService from './containers/QuotaService.jsx'

const store = configureStore()

const env = {
  environment: process.env.NODE_ENV || 'development'
}

render(
  <Provider store={store}>
    <QuotaService env={env} />
  </Provider>,
  document.getElementById('root')
)
