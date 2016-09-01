export const SELECT_NAMESPACE = 'SELECT_NAMESPACE'

export function selectNamespace(namespace) {
  return dispatch => dispatch({
    type: SELECT_NAMESPACE,
    namespace: namespace
  })
}
