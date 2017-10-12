import PropTypes from 'prop-types';
import React from 'react';
import { Component } from 'react';

import Namespace from './Namespace.jsx';
import Stats from './Stats.jsx';

export default class SelectedNamespace extends Component {
  render() {
    const {
      selectedNamespace, stats, actions
    } = this.props

    if (!selectedNamespace)
      return null

    return (<div className='flex-container flex-box-lg selected-namespace'>
      {stats.show ?
        <Stats namespace={selectedNamespace} stats={stats} {...actions} /> :
        <Namespace namespace={selectedNamespace} {...actions} />}
    </div>)
  }
}

SelectedNamespace.propTypes = {
  actions: PropTypes.object.isRequired,
  stats: PropTypes.object.isRequired,
  selectedNamespace: PropTypes.object
}
