import { CLEAR_CONFIRM, CONFIRM } from '../actions/confirmation.jsx'

export function confirm(state = null, action) {
  switch (action.type) {
    case CONFIRM:
      return action
    default:
      return null
  }
}
