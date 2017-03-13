import { configs } from '../../src/reducers/configs.jsx'

import {
  CONFIGS_REQUEST, CONFIGS_FAILURE,
  CONFIGS_FETCH_SUCCESS, CONFIGS_COMMIT_SUCCESS
} from '../../src/actions/configs.jsx'

describe('configs reducer', () => {
  it('should return the initial state', () => {
    expect(configs(undefined, {})).toEqual({})
  })

  it('should handle CONFIGS_FAILURE', () => {
    const err = {}

    expect(configs(undefined, {
      type: CONFIGS_FAILURE,
      payload: err
    })).toEqual({
      inRequest: false,
      error: err
    })
  })

  it('should handle CONFIGS_REQUEST', () => {
    shouldHandleError(CONFIGS_REQUEST)

    expect(configs(undefined, {
      type: CONFIGS_REQUEST
    })).toEqual({
      inRequest: true,
      error: null
    })
  })

  it('should handle CONFIGS_FETCH_SUCCESS', () => {
    shouldHandleError(CONFIGS_FETCH_SUCCESS)

    const items = {}

    expect(configs(undefined, {
      type: CONFIGS_FETCH_SUCCESS,
      payload: { configs: items }
    })).toEqual({ items: items })
  })

  it('should handle CONFIGS_COMMIT_SUCCESS', () => {
    shouldHandleError(CONFIGS_COMMIT_SUCCESS)

    expect(configs(undefined, {
      type: CONFIGS_COMMIT_SUCCESS
    })).toEqual({})
  })

  function shouldHandleError(type) {
    const err = {}

    expect(configs(undefined, {
      type: type,
      error: true,
      payload: err
    })).toEqual({
      inRequest: false,
      error: err
    })
  }
})
