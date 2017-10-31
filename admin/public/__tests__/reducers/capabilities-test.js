import { CAPABILITIES_FETCH_SUCCESS, CAPABILITIES_REQUEST } from '../../src/actions/capabilities.jsx';
import { capabilities } from '../../src/reducers/capabilities.jsx';

describe('capabilities reducer', () => {
  it('should return the initial state', () => {
    const state = capabilities(undefined, {});
    expect(state).toEqual({});
  });

  it('should handle CAPABILITIES_REQUEST', () => {
    const state = capabilities(undefined, { type: CAPABILITIES_REQUEST });
    expect(state).toEqual({ inRequest: true, error: null })
  });

  it('should handle CAPABILITIES_FETCH_SUCCESS', () => {
    const payload = { foo: 'bar' };
    const state = capabilities(undefined, { type: CAPABILITIES_FETCH_SUCCESS, payload });
    expect(state).toEqual({ ...payload });
  });
});
