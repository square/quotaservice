import React from 'react'
import { CALL_API } from 'redux-api-middleware'

import { confirm } from './confirmation.jsx'

export const CONFIGS_FAILURE = 'CONFIGS_FAILURE'
export const CONFIGS_REQUEST = 'CONFIGS_REQUEST'
export const CONFIGS_FETCH_SUCCESS = 'CONFIGS_FETCH_SUCCESS'
export const CONFIGS_COMMIT_SUCCESS = 'CONFIGS_COMMIT_SUCCESS'

export const LOAD_CONFIG = 'LOAD_CONFIG'

export function loadConfig(config) {
  return dispatch => dispatch({
    type: LOAD_CONFIG,
    config: config
  })
}

export function fetchConfigs() {
  return async dispatch => {
    try {
      dispatch({ type: CONFIGS_REQUEST });
      const response = await fetch('/api/configs', { method: 'GET', credentials: 'same-origin' });
      return dispatch({ type: CONFIGS_FETCH_SUCCESS, payload: await response.json() });
    } catch (e) {
      return dispatch({ type: CONFIGS_FAILURE });
    }
  }
}

export function commitConfig(namespaces, version) {
  const json = JSON.stringify({ namespaces: namespaces })

  return confirm({
    [CALL_API]: {
      endpoint: '/api',
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Version': version
      },
      body: json,
      credentials: 'same-origin',
      types: [CONFIGS_REQUEST, CONFIGS_COMMIT_SUCCESS, CONFIGS_FAILURE]
    }
  },
    'You are about to submit the following configuration.',
    <div className="code">{JSON.stringify(namespaces, null, 4)}</div>
  )
}
