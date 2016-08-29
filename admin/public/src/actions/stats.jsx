import { CALL_API } from 'redux-api-middleware'

export const STATS_TOGGLE = 'STATS_TOGGLE'
export const STATS_FAILURE = 'STATS_FAILURE'
export const STATS_REQUEST = 'STATS_REQUEST'
export const STATS_FETCH_SUCCESS = 'STATS_FETCH_SUCCESS'

export function toggleStats() {
  return (dispatch, getState) => dispatch({
    type: STATS_TOGGLE
  }).then(() => {
    const state = getState()
    if (state.stats.show) {
      dispatch(fetchStatsAction(state.selectedNamespace, null))
    }
  })
}

export function fetchStats(namespace, bucket) {
  return dispatch => dispatch(fetchStatsAction(namespace, bucket))
}

function fetchStatsAction(namespace, bucket) {
  let url = `/api/stats/${namespace}`

  if (bucket) {
    url = `${url}/${bucket}`
  }

  return {
    [CALL_API]: {
      endpoint: url,
      method: 'GET',
      credentials: 'same-origin',
      types: [STATS_REQUEST, STATS_FETCH_SUCCESS, STATS_FAILURE]
    }
  }
}
