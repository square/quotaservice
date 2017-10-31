import PropTypes from 'prop-types';
import React from 'react';
import { Component } from 'react';

import Field from '../components/Field.jsx';
import NamespaceHeader from '../components/NamespaceHeader.jsx';
import Bucket from './Bucket.jsx';

export default class Namespace extends Component {
  handleNamespaceChange = (keyName) => {
    return (value) => {
      const { updateNamespace, namespace } = this.props
      updateNamespace(namespace.name, keyName, value)
    }
  }

  handleBucketChange = (bucket) => {
    return (keyName, value) => {
      const { updateBucket, namespace } = this.props
      updateBucket(namespace.name, bucket.name, keyName, value)
    }
  }

  handleBucketRemove = (bucket) => {
    return () => {
      const { removeBucket, namespace } = this.props
      removeBucket(namespace.name, bucket.name)
    }
  }

  handleBack = () => {
    this.props.selectNamespace(null)
  }

  handleShowDynamicStats = () => {
    this.props.toggleStats()
  }

  render() {
    const { canMakeChanges = true, namespace, removeNamespace } = this.props
    const buckets = namespace.buckets

    return (
      <div className="namespace flex-box flex-tile">
        <NamespaceHeader namespace={namespace}
          canMakeChanges={canMakeChanges}
          handleBack={this.handleBack}
          removeNamespace={removeNamespace}
        />
        <div className="buckets flex-container flex-column flex-wrap">
          <div className="bucket flex-box flex-tile">
            <div className="flex-container legend">
              <h4>Namespace Configuration</h4>
            </div>
            <Field
              parent={namespace.name}
              disabled={canMakeChanges === false}
              keyName="max_dynamic_buckets"
              handleChange={this.handleNamespaceChange}
              value={namespace.max_dynamic_buckets}
              title="Maximum amount of dynamic buckets allowed before requests are rejected."
            />
          </div>
          {this.renderBucket(namespace.dynamic_bucket_template, true)}
          {this.renderBucket(namespace.default_bucket, false)}
          {buckets && Object.keys(buckets).map(key =>
            this.renderBucket(buckets[key], false)
          )}
        </div>
      </div>
    )
  }

  renderBucket(bucket, showDynamicStats) {
    if (!bucket)
      return

    const { canMakeChanges } = this.props;

    return <Bucket
      key={bucket.name}
      bucket={bucket}
      canMakeChanges={canMakeChanges}
      showDynamicStats={showDynamicStats}
      handleChange={this.handleBucketChange(bucket)}
      handleRemove={this.handleBucketRemove(bucket)}
      handleShowDynamicStats={this.handleShowDynamicStats}
    />
  }
}

Namespace.propTypes = {
  canMakeChanges: PropTypes.bool.isRequired,
  namespace: PropTypes.object.isRequired,
  selectNamespace: PropTypes.func.isRequired,
  updateNamespace: PropTypes.func.isRequired,
  removeNamespace: PropTypes.func.isRequired,
  updateBucket: PropTypes.func.isRequired,
  removeBucket: PropTypes.func.isRequired,
  toggleStats: PropTypes.func.isRequired
}
