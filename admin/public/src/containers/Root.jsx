import { connect } from 'react-redux'
import { bindActionCreators } from 'redux'

import QuotaService from './QuotaService.jsx'

import * as HistoryActions from '../actions/history.jsx'
import * as NamespacesActions from '../actions/namespaces.jsx'
import * as MutableActions from '../actions/mutable.jsx'
import * as StatsActions from '../actions/stats.jsx'
import * as ConfigsActions from '../actions/configs.jsx'
import * as ConfirmationActions from '../actions/confirmation.jsx'

export default connect(
  state => {
    let { selectedNamespace } = state

    if (selectedNamespace) {
      let namespace = state.namespaces.items[selectedNamespace]

      return Object.assign({}, state, {
        selectedNamespace: namespace
      })
    }

    return state
  },
  dispatch => {
    return {
      dispatch: dispatch, // only used for Confirmation
      actions: Object.assign({},
        bindActionCreators(NamespacesActions, dispatch),
        bindActionCreators(HistoryActions, dispatch),
        bindActionCreators(MutableActions, dispatch),
        bindActionCreators(StatsActions, dispatch),
        bindActionCreators(ConfigsActions, dispatch),
        bindActionCreators(ConfirmationActions, dispatch)
      )
    }
  }
)(QuotaService)
