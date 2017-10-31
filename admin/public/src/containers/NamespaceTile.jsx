import PropTypes from 'prop-types';
import React from 'react';
import { Component } from 'react';

export default class NamespaceTile extends Component {
  onClick = () => {
    const { namespace, selectNamespace, capabilitiesEnabled } = this.props;
    const { canMakeChanges } = this.state || {};
    selectNamespace(namespace.name, capabilitiesEnabled ? canMakeChanges === true : true);
  }

  fetchCanMakeChanges() {
    const hasCanMakeChanges = 'canMakeChanges' in (this.state || {});

    if (hasCanMakeChanges)
      return;

    const { namespace: { name: namespaceName }, eventEmitter } = this.props;
    const callback = canMakeChanges =>
      this.setState(prevState => ({ ...prevState, canMakeChanges }));

    eventEmitter.dispatchEvent(new CustomEvent(
      'QuotaService.getCapabilities',
      { detail: { namespaceName, callback } }
    ));
  }

  componentWillMount() {
    const { capabilities, capabilitiesEnabled } = this.props;
    if (capabilitiesEnabled && !capabilities.inRequest && !capabilities.error) {
      this.fetchCanMakeChanges();
    }
  }

  render() {
    const { isSelected, namespace } = this.props;
    const buckets = namespace.buckets || {};
    const className = 'flex-box flex-tile namespace' + (isSelected ? ' selected' : '');

    return (
      <div className={className} onClick={this.onClick}>
        <p className="title">
          {this.renderCanMakeChanges()}
          {namespace.name}
        </p>
        <hr />
        {this.renderBucket(namespace.dynamic_bucket_template)}
        {this.renderBucket(namespace.default_bucket)}
        {Object.keys(buckets).map(key =>
          this.renderBucket(buckets[key])
        )}
      </div>
    )
  }

  renderCanMakeChanges() {
    const { canMakeChanges } = this.state || {};

    return canMakeChanges
      ? <span className="canMakeChanges">🛠 </span>
      : null;
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
