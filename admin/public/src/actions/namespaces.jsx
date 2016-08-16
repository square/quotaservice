import { CALL_API } from 'redux-api-middleware'

export const FAILURE = 'FAILURE'
export const REQUEST = 'REQUEST'
export const FETCH_SUCCESS = 'FETCH_SUCCESS'
export const COMMIT_SUCCESS = 'COMMIT_SUCCESS'

export const SELECT_NAMESPACE = 'SELECT_NAMESPACE'

export function commitNamespaces() {
  return (dispatch, getState) => {
    const state = getState()
    const json = JSON.stringify({namespaces: state.namespaces.items })

    dispatch({
      [CALL_API]: {
        endpoint: '/api',
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: json,
        credentials: 'same-origin',
        types: [REQUEST, COMMIT_SUCCESS, FAILURE]
      }
    })
  }
}

export function fetchNamespaces() {
  return dispatch => dispatch({
    [CALL_API]: {
      endpoint: '/api',
      method: 'GET',
      credentials: 'same-origin',
      types: [REQUEST, FETCH_SUCCESS, FAILURE]
    }
  })
}

export function selectNamespace(namespace) {
  return dispatch => dispatch({
    type: SELECT_NAMESPACE,
    namespace: namespace
  })
}
