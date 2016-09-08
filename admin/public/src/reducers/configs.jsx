import {
  CONFIGS_REQUEST, CONFIGS_FAILURE,
  CONFIGS_FETCH_SUCCESS, CONFIGS_COMMIT_SUCCESS
} from '../actions/configs.jsx'

const INITIAL_CONFIGS = {}

function handleConfigsRequest(state, action) {
  if (action.error) {
    return Object.assign({}, state, {
      inRequest: false,
      error: action.payload
    })
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

export function configs(state = INITIAL_CONFIGS, action) {
  switch (action.type) {
    case CONFIGS_REQUEST:
    case CONFIGS_FAILURE:
    case CONFIGS_FETCH_SUCCESS:
    case CONFIGS_COMMIT_SUCCESS:
      return handleConfigsRequest(state, action)
    default:
      return state
  }
}
