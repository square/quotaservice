export const CLEAR_CONFIRM = 'CLEAR_CONFIRM'
export const CONFIRM = 'CONFIRM'

export function confirm(action, header, body) {
  return dispatch => dispatch({
    type: CONFIRM,
    action: action,
    header: header,
    body: body
  })
}

export function clearConfirm() {
  return dispatch => dispatch({ type: CLEAR_CONFIRM })
}
