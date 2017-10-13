import PropTypes from 'prop-types';
import React from 'react';
import { Component } from 'react';

export default class Confirmation extends Component {
  handleConfirm = () => {
    const { dispatch, action } = this.props
    dispatch(action)
  }

  render() {
    const { header, body, cancel } = this.props

    return (<div className="overlay fill-height-container flex-container flex-centered">
      <div className="confirmation flex-container flex-column">
        <h4>{header}</h4>
        {body}
        <div className="confirmation-footer">
          <button className="btn" onClick={cancel}>Cancel</button>
          <button className="btn btn-danger" onClick={this.handleConfirm}>Confirm</button>
        </div>
      </div>
    </div>)
  }
}

Confirmation.propTypes = {
  header: PropTypes.string.isRequired,
  action: PropTypes.object.isRequired,
  dispatch: PropTypes.func.isRequired,
  cancel: PropTypes.func.isRequired,
  body: PropTypes.object,
}
