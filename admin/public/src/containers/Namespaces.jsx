import PropTypes from 'prop-types';
import React from 'react';
import { Component } from 'react';

import NamespaceTile from './NamespaceTile.jsx';

export default class Namespaces extends Component {
  render() {
    const { actions, namespaces, selectedNamespace, capabilities, env } = this.props
    const { items } = namespaces
    const classNames = ['namespaces', 'flex-container', 'flex-box-lg']

    // Hides this div for small screens <= 1000px
    if (selectedNamespace) {
      classNames.push('flexed')
    }

    return (
      <div className={classNames.join(' ')}>
        {items && Object.keys(items).map(key =>
          <NamespaceTile
            capabilities={capabilities}
            capabilitiesEnabled={env.capabilities === true}
            eventEmitter={window}
            isSelected={items[key].name === (selectedNamespace ? selectedNamespace.namespace.name : '')}
            key={key}
            namespace={items[key]}
            selectNamespace={actions.selectNamespace}
          />
        )}
      </div>
    )
  }
}

Namespaces.propTypes = {
  actions: PropTypes.object.isRequired,
  namespaces: PropTypes.object.isRequired,
  env: PropTypes.object.isRequired,
  capabilities: PropTypes.object.isRequired,
  selectedNamespace: PropTypes.object,
}
