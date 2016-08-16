import React, { Component, PropTypes } from 'react'

export default class Error extends Component {
  render() {
    const { error } = this.props

    // TODO
    return (<div className="error">
      Error! {error.message}
    </div>)
  }
}

Error.propTypes = {
  error: PropTypes.object.isRequired
}
