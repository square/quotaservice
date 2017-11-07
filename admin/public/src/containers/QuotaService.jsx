import PropTypes from 'prop-types';
import React from 'react';
import { Component } from 'react';

import Confirmation from '../components/Confirmation.jsx';
import Namespaces from './Namespaces.jsx';
import SelectedNamespace from './SelectedNamespace.jsx';
import Sidebar from './Sidebar.jsx';

export default class QuotaService extends Component {
  fetchData() {
    const { actions, env } = this.props;
    const { fetchConfigs, fetchCapabilities } = actions;

    return fetchConfigs().then(configs => {
      if (env.capabilities) {
        return fetchCapabilities(configs.payload);
      }
      return null;
    });
  }

  componentDidMount() {
    this.fetchData();
  }

  componentWillReceiveProps(nextProps) {
    const { configs } = nextProps

    // Refetch the configs after the save because there's a chance the backend
    // may do some additional processing on the data.
    if (!configs.inRequest && !configs.error && configs.items === undefined) {
      this.fetchData();
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

  renderLoading() {
    return (
      <div className='flex-container flex-box-lg'>
        <div className='loader'>Loading...</div>
      </div>
    );
  }

  render() {
    const { selectedNamespace, configs, capabilities, confirm } = this.props
    const classNames = ['flex-container', 'fill-height-container']
    const isLoading = configs.inRequest || (capabilities && capabilities.inRequest);
    const canMakeChanges = selectedNamespace ? selectedNamespace.canMakeChanges : true;

    if (confirm) {
      classNames.push('blur')
    }

    return (
      <div>
        {confirm && this.renderConfirmation()}

        {!canMakeChanges &&
          <div className="warning">You do not have permissions to make changes to this namespace.</div>
        }

        <div className={classNames.join(' ')}>
          <Sidebar {...this.props} handleRefresh={this.fetchData.bind(this)} />
          {isLoading ? this.renderLoading() : <Namespaces {...this.props} />}
          <SelectedNamespace {...this.props} />
        </div>
      </div>
    )
  }
}

QuotaService.propTypes = {
  dispatch: PropTypes.func.isRequired,
  actions: PropTypes.object.isRequired,
  namespaces: PropTypes.object.isRequired,
  stats: PropTypes.object.isRequired,
  configs: PropTypes.object.isRequired,
  capabilities: PropTypes.object.isRequired,
  selectedNamespace: PropTypes.object,
  currentVersion: PropTypes.number.isRequired,
  env: PropTypes.object.isRequired,
  confirm: PropTypes.object
}
