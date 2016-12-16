import Immutable from 'seamless-immutable'

import { SELECT_NAMESPACE } from '../actions/namespaces.jsx'

import {
  ADD_NAMESPACE, UPDATE_NAMESPACE, REMOVE_NAMESPACE,
  ADD_BUCKET, UPDATE_BUCKET, REMOVE_BUCKET
} from '../actions/mutable.jsx'

import { INITIAL_HISTORY } from './history.jsx'
import { CONFIGS_REQUEST, LOAD_CONFIG } from '../actions/configs.jsx'

// These are special buckets that exist on the top-level
// namespace object and need to be special-cased
const BUCKET_KEY_MAP = {
  '___DEFAULT_BUCKET___': 'default_bucket',
  '___DYNAMIC_BUCKET_TPL___': 'dynamic_bucket_template'
}

function addNamespace(state, action) {
  return Object.assign({}, state, {
    change: {
      type: ADD_NAMESPACE,
      key: action.namespace
    },
    items: state.items.set(action.namespace, {
      buckets: {},
      name: action.namespace
    })
  })
}

function updateNamespace(state, action) {
  return Object.assign({}, state, {
    change: {
      type: UPDATE_NAMESPACE,
      key: `${action.namespace}.${action.key}`,
      value: action.value
    },
    items: state.items.setIn(
      [action.namespace, action.key],
      action.value
    )
  })
}

function removeNamespace(state, action) {
  return Object.assign({}, state, {
    change: {
      type: REMOVE_NAMESPACE,
      key: action.namespace
    },
    items: state.items.without(action.namespace)
  })
}

function addBucket(state, action) {
  let bucketPath = ['buckets', action.bucket]
  const bucketKey = BUCKET_KEY_MAP[action.bucket]

  if (bucketKey) {
    bucketPath = [bucketKey]
  }

  return Object.assign({}, state, {
    change: {
      type: ADD_BUCKET,
      key: `${action.namespace}.${action.bucket}`
    },
    items: state.items.setIn(
      [action.namespace, ...bucketPath],
      {
        name: action.bucket,
        namespace: action.namespace
      }
    )
  })
}

function updateBucket(state, action) {
  let bucketPath = ['buckets', action.bucket]
  const bucketKey = BUCKET_KEY_MAP[action.bucket]

  if (bucketKey) {
    action.bucket = bucketKey
    bucketPath = [bucketKey]
  }

  return Object.assign({}, state, {
    change: {
      type: UPDATE_BUCKET,
      key: `${action.namespace}.${action.bucket}.${action.key}`,
      value: action.value
    },
    items: state.items.setIn(
      [action.namespace, ...bucketPath, action.key],
      action.value
    )
  })
}

function removeBucket(state, action) {
  let bucketPath = ['buckets']
  const bucketKey = BUCKET_KEY_MAP[action.bucket]

  if (bucketKey) {
    action.bucket = bucketKey
    bucketPath = []
  }

  return Object.assign({}, state, {
    change: {
      type: REMOVE_BUCKET,
      key: `${action.namespace}.${action.bucket}`
    },
    items: state.items.updateIn(
      [action.namespace, ...bucketPath],
      (buckets) => buckets.without(action.bucket)
    )
  })
}

export function namespaces(state, action) {
  switch (action.type) {
    case ADD_NAMESPACE:
      return addNamespace(state, action)
    case UPDATE_NAMESPACE:
      return updateNamespace(state, action)
    case REMOVE_NAMESPACE:
      return removeNamespace(state, action)
    case ADD_BUCKET:
      return addBucket(state, action)
    case UPDATE_BUCKET:
      return updateBucket(state, action)
    case REMOVE_BUCKET:
      return removeBucket(state, action)
    case LOAD_CONFIG:
      return Object.assign({}, INITIAL_HISTORY, {
        version: action.config.version,
        items: Immutable.from(action.config.namespaces || {})
      })
    case CONFIGS_REQUEST:
      return INITIAL_HISTORY
    default:
      return state
  }
}

export function selectedNamespace(state = null, action) {
  switch (action.type) {
    case SELECT_NAMESPACE:
      return action.namespace
    case CONFIGS_REQUEST:
      return null
    default:
      return state
  }
}
