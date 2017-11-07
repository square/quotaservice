import PropTypes from 'prop-types';
import React from 'react';
import { Component } from 'react';

import Field from '../components/Field.jsx';


export default class Bucket extends Component {
  handleChange = (keyName) => {
    return (value) => {
      this.props.handleChange(keyName, value)
    }
  }

  renderShowDynamicStats() {
    const { handleShowDynamicStats } = this.props

    return (<div className="input-btn">
      <button className="btn" onClick={handleShowDynamicStats}>Dynamic Bucket Stats</button>
    </div>)
  }

  render() {
    const { canMakeChanges = true, bucket, handleRemove, showDynamicStats } = this.props

    return (<div className="bucket flex-tile flex-box">
      <div className="flex-container legend">
        <h4>{bucket.name}</h4>
        {canMakeChanges &&
          <button className="btn btn-danger" onClick={handleRemove}>Remove Bucket</button>
        }
      </div>
      <Field keyName="size"
        parent={bucket.name}
        disabled={canMakeChanges === false}
        value={bucket.size}
        handleChange={this.handleChange}
        title="Maximum number of tokens in a bucket."
        placeholder="100" />
      <Field keyName="fill_rate"
        parent={bucket.name}
        disabled={canMakeChanges === false}
        value={bucket.fill_rate}
        handleChange={this.handleChange}
        title="Token fill rate per second."
        placeholder="50" />
      <Field keyName="wait_timeout_millis"
        parent={bucket.name}
        disabled={canMakeChanges === false}
        value={bucket.wait_timeout_millis}
        handleChange={this.handleChange}
        title="Maximum time a request can wait for future tokens (milliseconds)."
        placeholder="1000" />
      <Field keyName="max_idle_millis"
        parent={bucket.name}
        disabled={canMakeChanges === false}
        value={bucket.max_idle_millis}
        handleChange={this.handleChange}
        title="When a bucket is idle (not serving requests), amount of time before a bucket resets to the initial size (milliseconds)."
        placeholder="-1" />
      <Field keyName="max_debt_millis"
        parent={bucket.name}
        disabled={canMakeChanges === false}
        value={bucket.max_debt_millis}
        handleChange={this.handleChange}
        title="Maximum amount of time in the future a request can pre-reserve tokens (milliseconds)."
        placeholder="10000" />
      <Field keyName="max_tokens_per_request"
        parent={bucket.name}
        disabled={canMakeChanges === false}
        value={bucket.max_tokens_per_request}
        handleChange={this.handleChange}
        title="Maximum number of tokens allowed per request."
        placeholder="50" />
      {showDynamicStats && this.renderShowDynamicStats()}
    </div>)
  }
}

Bucket.propTypes = {
  canMakeChanges: PropTypes.bool,
  bucket: PropTypes.object.isRequired,
  showDynamicStats: PropTypes.bool.isRequired,
  handleRemove: PropTypes.func.isRequired,
  handleChange: PropTypes.func.isRequired,
  handleShowDynamicStats: PropTypes.func.isRequired
}
