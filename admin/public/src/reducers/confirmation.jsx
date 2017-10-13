import { CONFIRM } from '../actions/confirmation.jsx';

export function confirm(state, action) {
  switch (action.type) {
    case CONFIRM:
      return action
    default:
      return null
  }
}
