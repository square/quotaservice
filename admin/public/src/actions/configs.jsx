import React from 'react'
import { CALL_API } from 'redux-api-middleware'

import { confirm } from './confirmation.jsx'

export const CONFIGS_FAILURE = 'CONFIGS_FAILURE'
export const CONFIGS_REQUEST = 'CONFIGS_REQUEST'
export const CONFIGS_FETCH_SUCCESS = 'CONFIGS_FETCH_SUCCESS'
export const CONFIGS_COMMIT_SUCCESS = 'CONFIGS_COMMIT_SUCCESS'

export const LOAD_CONFIG = 'LOAD_CONFIG'

function loadConfigState(config) {
  return {
    type: LOAD_CONFIG,
    config: config
  }
}

export function loadConfig(config) {
  return dispatch => dispatch(loadConfigState(config))
}

export function fetchConfigs() {
  return dispatch => dispatch({
    [CALL_API]: {
      endpoint: '/api/configs',
      method: 'GET',
      credentials: 'same-origin',
      types: [CONFIGS_REQUEST, CONFIGS_FETCH_SUCCESS, CONFIGS_FAILURE]
    }
  })
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
