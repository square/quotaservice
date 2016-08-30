import React, { Component, PropTypes } from 'react'
import Field from '../components/Field.jsx'

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
    const { bucket, handleRemove, showDynamicStats } = this.props

    return (<div className="bucket flex-tile flex-box">
      <div className="flex-container legend">
        <h4>{bucket.name}</h4>
        <button className="btn btn-danger" onClick={handleRemove}>Remove Bucket</button>
      </div>
      <Field keyName="size"
        parent={bucket.name}
        value={bucket.size}
        handleChange={this.handleChange}
        placeholder="100" />
      <Field keyName="fill_rate"
        parent={bucket.name}
        value={bucket.fill_rate}
        handleChange={this.handleChange}
        placeholder="50" />
      <Field keyName="wait_timeout_millis"
        parent={bucket.name}
        value={bucket.wait_timeout_millis}
        handleChange={this.handleChange}
        placeholder="1000" />
      <Field keyName="max_idle_millis"
        parent={bucket.name}
        value={bucket.max_idle_millis}
        handleChange={this.handleChange}
        placeholder="-1" />
      <Field keyName="max_debt_millis"
        parent={bucket.name}
        value={bucket.max_debt_millis}
        handleChange={this.handleChange}
        placeholder="10000" />
      <Field keyName="max_tokens_per_request"
        parent={bucket.name}
        value={bucket.max_tokens_per_request}
        handleChange={this.handleChange}
        placeholder="50" />
      {showDynamicStats && this.renderShowDynamicStats()}
    </div>)
  }
}

Bucket.propTypes = {
  bucket: PropTypes.object.isRequired,
  showDynamicStats: PropTypes.bool.isRequired,
  handleRemove: PropTypes.func.isRequired,
  handleChange: PropTypes.func.isRequired,
  handleShowDynamicStats: PropTypes.func.isRequired
}
