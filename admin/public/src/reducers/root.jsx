import { combineReducers } from 'redux'
import { namespaces, selectedNamespace } from './namespaces.jsx'
import { stats } from './stats.jsx'
import { history } from './history.jsx'
import { configs, currentVersion } from './configs.jsx'
import { confirm } from './confirmation.jsx'

export default combineReducers({
  namespaces: history(namespaces),
  selectedNamespace,
  currentVersion,
  stats,
  configs,
  confirm
})
