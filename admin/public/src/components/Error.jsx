import PropTypes from 'prop-types';
import React from 'react';
import { Component } from 'react';

export default class Error extends Component {
  renderError(error) {
    switch (error.name) {
      case 'RequestError':
        return `A network error occurred: "${error.message}"`
      case 'ApiError':
        if (error.response) {
          return error.response.description
        } else {
          return `An error occurred: "${error.statusText}"`
        }
      case 'InternalError':
      default:
        return 'An unknown error occurred. Please contact your friendly QuotaService owners for help.'
    }
  }

  render() {
    return (<div className="error">
      {this.renderError(this.props.error)}
    </div>)
  }
}

Error.propTypes = {
  error: PropTypes.object.isRequired
}
