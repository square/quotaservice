import { combineReducers } from 'redux'
import { namespaces, selectedNamespace } from './namespaces.jsx'
import { history } from './history.jsx'

export default combineReducers({
  namespaces: history(namespaces),
  selectedNamespace
})
