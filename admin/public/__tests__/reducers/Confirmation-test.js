import { CONFIRM } from '../../src/actions/confirmation.jsx';
import { confirm } from '../../src/reducers/confirmation.jsx';

describe('confirm reducer', () => {
  it('should return the initial state', () => {
    expect(confirm(undefined, {})).toEqual(null)
  })

  it('should handle CONFIRM', () => {
    expect(confirm(undefined, {
      type: CONFIRM,
      payload: {}
    })).toEqual({ type: CONFIRM, payload: {} })
  })
})
