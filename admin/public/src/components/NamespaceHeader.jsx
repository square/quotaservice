import PropTypes from 'prop-types';
import React from 'react';
import { Component } from 'react';

export default class NamespaceHeader extends Component {
  handleNamespaceRemove = () => {
    const { removeNamespace, namespace } = this.props
    removeNamespace(namespace.name)
  }

  render() {
    const { namespace, handleBack } = this.props

    return (<div className="namespace-navbar flex-container">
      <button className="btn" onClick={handleBack}>Back</button>
      <p className="title">{namespace.name}</p>
      <button className="btn btn-danger" onClick={this.handleNamespaceRemove}>Remove Namespace</button>
    </div>)
  }
}

NamespaceHeader.propTypes = {
  namespace: PropTypes.object.isRequired,
  removeNamespace: PropTypes.func.isRequired,
  handleBack: PropTypes.func.isRequired
}
