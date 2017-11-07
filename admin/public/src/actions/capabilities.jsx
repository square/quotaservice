import Promise from '../promise';

export const CAPABILITIES_REQUEST = 'CAPABILITIES_REQUEST';
export const CAPABILITIES_FETCH_SUCCESS = 'CAPABILITIES_FETCH_SUCCESS';

export function fetchCapabilities(configs) {
  function request() {
    return new Promise(resolve => {
      function callback(payload) {
        resolve({ type: CAPABILITIES_FETCH_SUCCESS, payload });
      }

      window.dispatchEvent(new CustomEvent(
        'QuotaService.fetchCapabilities',
        { detail: { configs, callback } }
      ));
    });
  }

  return async dispatch => {
    dispatch({ type: CAPABILITIES_REQUEST, configs });
    dispatch(await request());
  }
}
