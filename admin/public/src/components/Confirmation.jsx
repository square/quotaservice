import React, { Component, PropTypes } from 'react'

export default class Confirmation extends Component {
  render() {
    const { handleCancel, handleSubmit, json } = this.props

    return (<div className="overlay fill-height-container flex-container flex-centered">
      <div className="confirmation flex-container flex-column">
        <div>
          <div className="pull-right">
            <button className="btn" onClick={handleCancel}>Cancel</button>
            <button className="btn btn-danger" onClick={handleSubmit}>Submit</button>
          </div>
          <h4>You are about to submit the following configuration.</h4>
        </div>
        <div className="code">{json}</div>
      </div>
    </div>)
  }
}

Confirmation.propTypes = {
  json: PropTypes.string.isRequired,
  handleCancel: PropTypes.func.isRequired,
  handleSubmit: PropTypes.func.isRequired
}
