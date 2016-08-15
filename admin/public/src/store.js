import { createStore, applyMiddleware } from 'redux'
import thunk from 'redux-thunk'
import { apiMiddleware } from 'redux-api-middleware'

import rootReducer from './reducers/root.jsx'

export default function configureStore() {
  const middlewares = [apiMiddleware, thunk]

  if (process.env.NODE_ENV !== 'production') {
    const createLogger = require('redux-logger')
    const logger = createLogger()
    middlewares.push(logger)
  }

  return createStore(
    rootReducer,
    applyMiddleware(...middlewares)
  )
}
