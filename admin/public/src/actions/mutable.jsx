import { SELECT_NAMESPACE } from './namespaces.jsx'

export const ADD_NAMESPACE = 'ADD_NAMESPACE'
export const UPDATE_NAMESPACE = 'UPDATE_NAMESPACE'
export const REMOVE_NAMESPACE = 'REMOVE_NAMESPACE'

export const ADD_BUCKET = 'ADD_BUCKET'
export const UPDATE_BUCKET = 'UPDATE_BUCKET'
export const REMOVE_BUCKET = 'REMOVE_BUCKET'

export function addNamespace(name) {
  return dispatch => {
    dispatch({
      type: ADD_NAMESPACE,
      namespace: name
    })
    dispatch({
      type: SELECT_NAMESPACE,
      namespace: name
    })
  }
}

export function updateNamespace(namespace, key, value) {
  return dispatch => {
    dispatch({
      type: UPDATE_NAMESPACE,
      namespace: namespace,
      key: key,
      value: value
    })
  }
}

export function removeNamespace(namespace) {
  return dispatch => {
    dispatch({
      type: SELECT_NAMESPACE,
      namespace: null
    })

    dispatch({
      type: REMOVE_NAMESPACE,
      namespace: namespace
    })
  }
}

export function addBucket(namespace, name) {
  return dispatch => {
    dispatch({
      type: ADD_BUCKET,
      namespace: namespace,
      bucket: name
    })
  }
}

export function updateBucket(namespace, bucket, key, value) {
  return dispatch => {
    dispatch({
      type: UPDATE_BUCKET,
      namespace: namespace,
      bucket: bucket,
      key: key,
      value: value
    })
  }
}

export function removeBucket(namespace, bucket) {
  return dispatch => {
    dispatch({
      type: REMOVE_BUCKET,
      namespace: namespace,
      bucket: bucket
    })
  }
}
