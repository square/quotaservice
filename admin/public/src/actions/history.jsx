export const UNDO = 'UNDO'
export const REDO = 'REDO'
export const COMMIT = 'COMMIT'
export const CANCEL_COMMIT = 'CANCEL_COMMIT'
export const CLEAR = 'CLEAR'

export function undo() {
  return dispatch => dispatch({ type: UNDO })
}

export function redo() {
  return dispatch => dispatch({ type: REDO })
}

export function commit() {
  return dispatch => dispatch({ type: COMMIT })
}

export function cancelCommit() {
  return dispatch => dispatch({ type: CANCEL_COMMIT })
}
