import PropTypes from 'prop-types';
import React from 'react';
import { Component } from 'react';

import NamespaceTile from './NamespaceTile.jsx';

export default class Namespaces extends Component {
  render() {
    const { actions, configs, namespaces, selectedNamespace } = this.props
    const { items } = namespaces

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
        <NamespaceTile
          key={key}
          isSelected={items[key] === selectedNamespace}
          namespace={items[key]}
          {...actions} />
      )}
    </div>)
  }
}

Namespaces.propTypes = {
  actions: PropTypes.object.isRequired,
  namespaces: PropTypes.object.isRequired,
  selectedNamespace: PropTypes.object,
  configs: PropTypes.object.isRequired
}
