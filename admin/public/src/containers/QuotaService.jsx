import PropTypes from 'prop-types';
import React from 'react';
import { Component } from 'react';

import Confirmation from '../components/Confirmation.jsx';
import Namespaces from './Namespaces.jsx';
import SelectedNamespace from './SelectedNamespace.jsx';
import Sidebar from './Sidebar.jsx';

export default class QuotaService extends Component {
  componentDidMount() {
    this.props.actions.fetchConfigs()
  }

  componentWillReceiveProps(nextProps) {
    const { configs, actions } = nextProps

    if (!configs.inRequest && !configs.error && configs.items === undefined) {
      actions.fetchConfigs()
    }
  }

  renderConfirmation() {
    const { actions, dispatch, confirm } = this.props
    return (<Confirmation
      cancel={actions.clearConfirm}
      dispatch={dispatch}
      {...confirm}
    />)
  }

  render() {
    const { confirm } = this.props

    const classNames = ['flex-container', 'fill-height-container']

    if (confirm) {
      classNames.push('blur')
    }

    return (<div>
      {confirm && this.renderConfirmation()}
      <div className={classNames.join(' ')}>
        <Sidebar {...this.props} />
        <Namespaces {...this.props} />
        <SelectedNamespace {...this.props} />
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
