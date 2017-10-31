export const SELECT_NAMESPACE = 'SELECT_NAMESPACE';

export function selectNamespace(namespace, canMakeChanges = true) {
  return dispatch => dispatch({
    type: SELECT_NAMESPACE,
    namespace,
    canMakeChanges,
  });
}
