import PropTypes from 'prop-types';
import React from 'react';
import { Component } from 'react';

export default class NamespaceTile extends Component {
  constructor() {
    super();
    this.state = {};
  }

  canMakeChanges() {
    // permissive by default, to maintain previous functionality in case of misconfiguration
    return this.state.canMakeChanges !== false;
  }

  onClick = () => {
    const { namespace, selectNamespace, capabilitiesEnabled } = this.props;
    selectNamespace(namespace.name, capabilitiesEnabled ? this.canMakeChanges() : true);
  }

  fetchCanMakeChanges() {
    if ('canMakeChangesFetched' in this.state)
      return;

    const { namespace: { name: namespaceName }, eventEmitter } = this.props;
    const callback = (canMakeChanges = true) =>
      this.setState(prevState => ({ ...prevState, canMakeChanges, canMakeChangesFetched: true }));

    eventEmitter.dispatchEvent(new CustomEvent(
      'QuotaService.getCapabilities',
      { detail: { namespaceName, callback } }
    ));
  }

  UNSAFE_componentWillMount() {
    const { capabilities, capabilitiesEnabled } = this.props;
    if (capabilitiesEnabled && !capabilities.inRequest && !capabilities.error) {
      this.fetchCanMakeChanges();
    }
  }

  render() {
    const { isSelected, namespace } = this.props;
    const buckets = namespace.buckets || {};
    const isGrayedOutBecauseCantMakeChanges = this.props.capabilitiesEnabled ? !this.canMakeChanges() : false;
    const className = 'flex-box flex-tile namespace'
      + (isSelected ? ' selected' : '')
      + (isGrayedOutBecauseCantMakeChanges ? ' disabled' : '');

    return (
      <div className={className} onClick={this.onClick}>
        <p className="title">{namespace.name}</p>
        <hr />
        {this.renderBucket(namespace.dynamic_bucket_template)}
        {this.renderBucket(namespace.default_bucket)}
        {Object.keys(buckets).map(key =>
          this.renderBucket(buckets[key])
        )}
      </div>
    )
  }

  renderBucket(bucket) {
    if (!bucket)
      return;

    return <div key={bucket.name} className="bucket">{bucket.name}</div>;
  }
}

NamespaceTile.propTypes = {
  isSelected: PropTypes.bool.isRequired,
  namespace: PropTypes.object.isRequired,
  capabilities: PropTypes.object.isRequired,
  eventEmitter: PropTypes.object.isRequired,
  capabilitiesEnabled: PropTypes.bool.isRequired,
  selectNamespace: PropTypes.func.isRequired
}
