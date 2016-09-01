import { CALL_API } from 'redux-api-middleware'

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
  return (dispatch, getState) => dispatch({
    [CALL_API]: {
      endpoint: '/api/configs',
      method: 'GET',
      credentials: 'same-origin',
      types: [CONFIGS_REQUEST, CONFIGS_FETCH_SUCCESS, CONFIGS_FAILURE]
    }
  }).then(() => {
    const configs = getState().configs

    if (configs.items.length > 0) {
      dispatch(loadConfigState(configs.items[0]))
    }
  })
}

export function commitConfig(namespaces) {
  return dispatch => {
    const json = JSON.stringify({ namespaces: namespaces })

    dispatch({
      [CALL_API]: {
        endpoint: '/api',
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: json,
        credentials: 'same-origin',
        types: [CONFIGS_REQUEST, CONFIGS_COMMIT_SUCCESS, CONFIGS_FAILURE]
      }
    })
  }
}
