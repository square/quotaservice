import Immutable from 'seamless-immutable'

import { SELECT_NAMESPACE } from '../actions/namespaces.jsx'
import { CONFIGS_REQUEST } from '../actions/configs.jsx'

import {
  STATS_TOGGLE,
  STATS_FETCH_SUCCESS,
  STATS_FAILURE, STATS_REQUEST
} from '../actions/stats.jsx'

const INITIAL_STATS = { show: false }

export function stats(state = INITIAL_STATS, action) {
  switch (action.type) {
    case STATS_TOGGLE:
      return Object.assign({}, state, {
        show: !state.show
      })
    case SELECT_NAMESPACE:
    case CONFIGS_REQUEST:
      return INITIAL_STATS
    case STATS_REQUEST:
    case STATS_FAILURE:
    case STATS_FETCH_SUCCESS:
      return handleStatsRequest(state, action)
    default:
      return state
  }
}

function handleStatsRequest(state, action) {
  if (action.error) {
    return Object.assign({}, state, {
      inRequest: false,
      error: action.payload
    })
  }

  switch (action.type) {
    case STATS_REQUEST:
      return Object.assign({}, state, {
        inRequest: true,
        error: null
      })
    case STATS_FETCH_SUCCESS: {
      let { items } = state
      let newItems = Object.assign({}, items || {}, action.payload || {})

      return Object.assign({}, state, {
        inRequest: false,
        items: Immutable.from(newItems)
      })
    }
  }
}
