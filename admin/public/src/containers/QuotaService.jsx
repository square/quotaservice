import React, { Component, PropTypes } from 'react'
import { connect } from 'react-redux'
import { bindActionCreators } from 'redux'

import Stats from './Stats.jsx'
import Namespace from './Namespace.jsx'
import NamespaceTile from './NamespaceTile.jsx'
import Sidebar from './Sidebar.jsx'
import Confirmation from '../components/Confirmation.jsx'

import * as HistoryActions from '../actions/history.jsx'
import * as NamespacesActions from '../actions/namespaces.jsx'
import * as MutableActions from '../actions/mutable.jsx'
import * as StatsActions from '../actions/stats.jsx'
import * as ConfigsActions from '../actions/configs.jsx'
import * as ConfirmationActions from '../actions/confirmation.jsx'

class QuotaService extends Component {
  componentDidMount() {
    this.props.actions.fetchConfigs()
  }

  componentWillReceiveProps(nextProps) {
    const { configs, actions } = nextProps

    if (!configs.inRequest && !configs.error && configs.items === undefined) {
      actions.fetchConfigs()
    }
  }

  renderSelectedNamespace() {
    const {
      selectedNamespace, stats, actions
    } = this.props

    if (!selectedNamespace)
      return

    return (<div className='flex-container flex-box-lg selected-namespace'>
      {stats.show ?
        <Stats namespace={selectedNamespace} stats={stats} {...actions} /> :
        <Namespace namespace={selectedNamespace} {...actions} />}
    </div>)
  }

  renderNamespaces() {
    const { actions, configs, namespaces, selectedNamespace } = this.props
    const { items, } = namespaces

    if (configs.inRequest) {
      return (<div className='flex-container flex-box-lg'>
        <div className='loader'>Loading...</div>
      </div>)
    }

    const classNames = ['namespaces', 'flex-container', 'flex-box-lg']

    // Hides this div for small screens <= 1000px
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
    const { actions, dispatch, confirm } = this.props
    return (<Confirmation
      cancel={actions.clearConfirm}
      dispatch={dispatch}
      {...confirm}
    />)
  }

  handleConfigCommit = () => {
    const { actions, namespaces, currentVersion } = this.props
    actions.commitConfig(namespaces.items, currentVersion)
  }

  render() {
    const {
      actions, namespaces, env, confirm,
      selectedNamespace, configs, currentVersion
    } = this.props
    const { history, version } = namespaces
    const { lastUpdated, error } = configs

    const classNames = ['flex-container', 'fill-height-container']

    if (confirm) {
      classNames.push('blur')
    }

    return (<div>
      {confirm && this.renderConfirmation()}
      <div className={classNames.join(' ')}>
        <Sidebar
          selectedNamespace={selectedNamespace}
          env={env}
          currentVersion={currentVersion}
          version={version || 0}
          lastUpdated={lastUpdated}
          changes={history}
          configs={configs}
          error={error}
          commit={this.handleConfigCommit}
          {...actions}
        />
        {this.renderNamespaces()}
        {this.renderSelectedNamespace()}
      </div>
    </div>)
  }
}

QuotaService.propTypes = {
  dispatch: PropTypes.func.isRequired,
  actions: PropTypes.object.isRequired,
  namespaces: PropTypes.object.isRequired,
  stats: PropTypes.object.isRequired,
  configs: PropTypes.object.isRequired,
  selectedNamespace: PropTypes.object,
  currentVersion: PropTypes.number.isRequired,
  env: PropTypes.object.isRequired,
  confirm: PropTypes.object
}

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
