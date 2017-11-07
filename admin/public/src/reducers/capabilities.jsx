import { CAPABILITIES_REQUEST, CAPABILITIES_FETCH_SUCCESS } from '../actions/capabilities.jsx';

export function capabilities(state = {}, action) {
  switch (action.type) {
    case CAPABILITIES_REQUEST:
      return {
        ...state,
        inRequest: true,
        error: null
      };

    case CAPABILITIES_FETCH_SUCCESS:
      return { ...action.payload };

    default:
      return state;
  }
}
