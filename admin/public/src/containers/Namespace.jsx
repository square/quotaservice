import React, { Component, PropTypes } from 'react'
import Bucket from './Bucket.jsx'
import Field from '../components/Field.jsx'

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

  handleNamespaceRemove = () => {
    const { removeNamespace, namespace } = this.props
    removeNamespace(namespace.name)
  }

  render() {
    const { namespace } = this.props
    const buckets = namespace.buckets

    return (<div className="namespace flex-box flex-tile">
      <div className="namespace-navbar flex-container">
        <button className="btn" onClick={this.handleBack}>Back</button>
        <p className="title">{namespace.name}</p>
        <button className="btn btn-danger" onClick={this.handleNamespaceRemove}>Remove Namespace</button>
      </div>
      <div className="buckets flex-container flex-column flex-wrap">
        <div className="bucket flex-box flex-tile">
          <div className="flex-container legend">
            <h4>namespace config</h4>
          </div>
          <Field
            parent={namespace.name}
            keyName="max_dynamic_buckets"
            handleChange={this.handleNamespaceChange}
            value={namespace.max_dynamic_buckets}
          />
        </div>
        {this.renderBucket(namespace.dynamic_bucket_template)}
        {this.renderBucket(namespace.default_bucket)}
        {buckets && Object.keys(buckets).map(key =>
            this.renderBucket(buckets[key])
        )}
      </div>
    </div>)
  }

  renderBucket(bucket) {
    if (!bucket)
      return

    return (<Bucket
      key={bucket.name} bucket={bucket}
      handleChange={this.handleBucketChange(bucket)}
      handleRemove={this.handleBucketRemove(bucket)}
    />)
  }
}

Namespace.propTypes = {
  namespace: PropTypes.object.isRequired,
  selectNamespace: PropTypes.func.isRequired,
  updateNamespace: PropTypes.func.isRequired,
  removeNamespace: PropTypes.func.isRequired,
  updateBucket: PropTypes.func.isRequired,
  removeBucket: PropTypes.func.isRequired
}
