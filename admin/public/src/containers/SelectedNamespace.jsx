import PropTypes from 'prop-types';
import React from 'react';
import { Component } from 'react';

import Namespace from './Namespace.jsx';
import Stats from './Stats.jsx';

export default class SelectedNamespace extends Component {
  render() {
    const { selectedNamespace, stats, actions } = this.props;

    if (!selectedNamespace)
      return null;

    const { namespace, canMakeChanges } = selectedNamespace;

    return (
      <div className="flex-container flex-box-lg selected-namespace">
        {stats.show ?
          <Stats namespace={namespace} stats={stats} {...actions} /> :
          <Namespace namespace={namespace} canMakeChanges={canMakeChanges} {...actions} />}
      </div>
    )
  }
}

SelectedNamespace.propTypes = {
  actions: PropTypes.object.isRequired,
  stats: PropTypes.object.isRequired,
  selectedNamespace: PropTypes.object
}
