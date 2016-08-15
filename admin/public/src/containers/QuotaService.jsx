import React, { Component, PropTypes } from 'react'
import { connect } from 'react-redux'
import { bindActionCreators } from 'redux'

import Namespace from './Namespace.jsx'
import NamespaceTile from './NamespaceTile.jsx'
import Sidebar from './Sidebar.jsx'
import Confirmation from '../components/Confirmation.jsx'

import * as HistoryActions from '../actions/history.jsx'
import * as NamespacesActions from '../actions/namespaces.jsx'
import * as MutableActions from '../actions/mutable.jsx'

class QuotaService extends Component {
  componentDidMount() {
    this.props.actions.fetchNamespaces()
  }

  componentWillReceiveProps(nextProps) {
    const { namespaces } = nextProps

    if (!namespaces.inRequest && !namespaces.error && namespaces.items === undefined) {
      this.props.actions.fetchNamespaces()
    }
  }

  renderSelectedNamespace() {
    const { selectedNamespace, actions } = this.props

    if (!selectedNamespace)
      return

    return (<div className='flex-container flex-box-md selected-namespace'>
      <Namespace namespace={selectedNamespace} {...actions} />
    </div>)
  }

  renderNamespaces() {
    const { actions, namespaces, selectedNamespace } = this.props
    const { items, inRequest } = namespaces

    if (inRequest) {
      return (<div className='flex-container flex-box-lg'>
        <div className='loader'>Loading...</div>
      </div>)
    }

    const classNames = ['namespaces', 'flex-container', 'flex-box-lg']

    // Hides this div for small screens <= 580px
    if (selectedNamespace) {
      classNames.push('flexed')
    }

    return (<div className={classNames.join(' ')}>
      {items && Object.keys(items).map(key =>
          <NamespaceTile key={key} namespace={items[key]} {...actions} />
      )}
    </div>)
  }

  renderConfirmation() {
    const { namespaces, actions } = this.props
    const json = JSON.stringify(namespaces.items, null, 4)

    return (<Confirmation
      json={json}
      handleCancel={actions.cancelCommit}
      handleSubmit={actions.commitNamespaces}
    />)
  }

  render() {
    const { actions, namespaces, env, selectedNamespace } = this.props
    const { lastUpdated, error, history, commit } = namespaces

    const classNames = ['flex-container', 'fill-height-container']

    if (commit) {
      classNames.push('blur')
    }

    return (<div>
      {commit && this.renderConfirmation()}
      <div className={classNames.join(' ')}>
        <Sidebar selectedNamespace={selectedNamespace} env={env} lastUpdated={lastUpdated} changes={history} error={error} {...actions} />
        {this.renderNamespaces()}
        {this.renderSelectedNamespace()}
      </div>
    </div>)
  }
}

QuotaService.propTypes = {
  actions: PropTypes.object.isRequired,
  namespaces: PropTypes.object.isRequired,
  selectedNamespace: PropTypes.object,
  env: PropTypes.object.isRequired
}

export default connect(
  state => {
    let selectedNamespace = state.selectedNamespace

    if (selectedNamespace) {
      selectedNamespace = state.namespaces.items[state.selectedNamespace]
    }

    return Object.assign({}, state, {
      selectedNamespace: selectedNamespace
    })
  },
  dispatch => {
    return {
      actions: Object.assign({},
        bindActionCreators(NamespacesActions, dispatch),
        bindActionCreators(HistoryActions, dispatch),
        bindActionCreators(MutableActions, dispatch)
      )
    }
  }
)(QuotaService)
