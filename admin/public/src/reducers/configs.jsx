import {
  CONFIGS_REQUEST, CONFIGS_FAILURE,
  CONFIGS_FETCH_SUCCESS, CONFIGS_COMMIT_SUCCESS
} from '../actions/configs.jsx'

const INITIAL_CONFIGS = {}

function handleConfigsRequest(state, action) {
  if (action.error) {
    return handleConfigsFailure(state, action)
  }

  switch (action.type) {
    case CONFIGS_REQUEST:
      return Object.assign({}, state, {
        inRequest: true,
        error: null
      })
    case CONFIGS_FETCH_SUCCESS:
      return Object.assign({}, INITIAL_CONFIGS, {
        items: action.payload.configs
      })
    case CONFIGS_COMMIT_SUCCESS:
      return INITIAL_CONFIGS
  }
}

function handleConfigsFailure(state, action) {
  return Object.assign({}, state, {
    inRequest: false,
    error: action.payload
  })
}

export function configs(state = INITIAL_CONFIGS, action) {
  switch (action.type) {
    case CONFIGS_FAILURE:
      return handleConfigsFailure(state, action)
    case CONFIGS_REQUEST:
    case CONFIGS_FETCH_SUCCESS:
    case CONFIGS_COMMIT_SUCCESS:
      return handleConfigsRequest(state, action)
    default:
      return state
  }
}

export function currentVersion(state = 0, action) {
  switch (action.type) {
    case CONFIGS_FETCH_SUCCESS: {
      const configs = action.payload.configs
      if (configs.length > 0) {
        return configs[0].version || 0
      }
    }
    // fall through otherwise
    default:
      return state
  }
}
