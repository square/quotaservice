import PropTypes from 'prop-types';
import React from 'react';
import { Component } from 'react';

export default class NamespaceTile extends Component {
  onClick = () => {
    const { namespace, selectNamespace } = this.props
    selectNamespace(namespace.name)
  }

  render() {
    const { isSelected, namespace } = this.props
    const buckets = namespace.buckets || {}

    return (<div className={'flex-box flex-tile namespace' + (isSelected ? ' selected' : '')} onClick={this.onClick}>
      <p className="title">{namespace.name}</p>
      <hr />
      {this.renderBucket(namespace.dynamic_bucket_template)}
      {this.renderBucket(namespace.default_bucket)}
      {Object.keys(buckets).map(key =>
        this.renderBucket(buckets[key])
      )}
    </div>)
  }

  renderBucket(bucket) {
    if (!bucket)
      return

    return (<div key={bucket.name} className="bucket">{bucket.name}</div>)
  }
}

NamespaceTile.propTypes = {
  isSelected: PropTypes.bool.isRequired,
  namespace: PropTypes.object.isRequired,
  selectNamespace: PropTypes.func.isRequired
}
