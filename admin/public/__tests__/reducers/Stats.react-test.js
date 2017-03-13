import { stats } from '../../src/reducers/stats.jsx'

import {
  STATS_TOGGLE,
  STATS_REQUEST, STATS_FAILURE,
  STATS_FETCH_SUCCESS, STATS_COMMIT_SUCCESS
} from '../../src/actions/stats.jsx'

import { SELECT_NAMESPACE } from '../../src/actions/namespaces.jsx'
import { CONFIGS_REQUEST } from '../../src/actions/configs.jsx'

describe('stats reducer', () => {
  it('should return the initial state', () => {
    expect(stats(undefined, {})).toEqual({ show: false })
  })

  it('should handle STATS_TOGGLE', () => {
    expect(stats({ show: false }, { type: STATS_TOGGLE })).toEqual({ show: true })
    expect(stats({ show: true }, { type: STATS_TOGGLE })).toEqual({ show: false })
    expect(stats({ }, { type: STATS_TOGGLE })).toEqual({ show: true })
  })

  it('should handle CONFIGS_REQUEST, SELECT_NAMESPACE', () => {
    expect(stats({ show: true }, { type: CONFIGS_REQUEST })).toEqual({ show: false })
    expect(stats({ show: true }, { type: SELECT_NAMESPACE })).toEqual({ show: false })
  })

  it('should handle STATS_FAILURE', () => {
    const err = {}

    expect(stats(undefined, {
      type: STATS_FAILURE,
      error: true,
      payload: err
    })).toEqual({
      inRequest: false,
      error: err,
      show: false
    })
  })

  it('should handle STATS_REQUEST', () => {
    shouldHandleError(STATS_REQUEST)

    expect(stats(undefined, {
      type: STATS_REQUEST
    })).toEqual({
      inRequest: true,
      error: null,
      show: false
    })
  })

  it('should handle STATS_FETCH_SUCCESS', () => {
    shouldHandleError(STATS_FETCH_SUCCESS)

    const payload = { stats: 2 }

    expect(stats({ items: { foo: 1 } }, {
      type: STATS_FETCH_SUCCESS,
      payload: payload,
    })).toEqual({
      inRequest: false,
      items: {
        foo: 1,
        stats: 2
      }
    })
  })

  function shouldHandleError(type) {
    const err = {}

    expect(stats(undefined, {
      type: type,
      error: true,
      payload: err
    })).toEqual({
      inRequest: false,
      error: err,
      show: false
    })
  }
})
