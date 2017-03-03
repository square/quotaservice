export const UNDO = 'UNDO'
export const REDO = 'REDO'
export const CLEAR = 'CLEAR'

export function undo() {
  return dispatch => dispatch({ type: UNDO })
}

export function redo() {
  return dispatch => dispatch({ type: REDO })
}
