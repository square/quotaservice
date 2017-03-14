import { currentVersion } from '../../src/reducers/configs.jsx'
import { CONFIGS_FETCH_SUCCESS } from '../../src/actions/configs.jsx'

describe('currentVersion reducer', () => {
  it('should return the initial state', () => {
    expect(currentVersion(undefined, {})).toEqual(0)
  })

  it('should handle CONFIGS_FETCH_SUCCESS', () => {
    expect(currentVersion(undefined, {
      type: CONFIGS_FETCH_SUCCESS,
      payload: {
        configs: [{
          version: 3
        }]
      }
    })).toEqual(3)

    expect(currentVersion(undefined, {
      type: CONFIGS_FETCH_SUCCESS,
      payload: {
        configs: [{}]
      }
    })).toEqual(0)

    expect(currentVersion(undefined, {
      type: CONFIGS_FETCH_SUCCESS,
      payload: {
        configs: []
      }
    })).toEqual(0)
  })
})
