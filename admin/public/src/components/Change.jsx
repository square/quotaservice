import React, { Component, PropTypes } from 'react'
import {
  ADD_NAMESPACE, UPDATE_NAMESPACE, REMOVE_NAMESPACE,
  ADD_BUCKET, UPDATE_BUCKET, REMOVE_BUCKET
} from '../actions/mutable.jsx'

export default class Change extends Component {
  render() {
    const { className } = this.props

    return (<div className={`change ${className}`}>
      {this.description()}
    </div>)
  }

  description() {
    const { change } = this.props

    switch(change.type) {
      case ADD_NAMESPACE:
      case ADD_BUCKET:
        return <span className="change-text">add {change.key}</span>
      case UPDATE_NAMESPACE:
      case UPDATE_BUCKET:
        return <span className="change-text">set {change.key} to "{change.value}"</span>
      case REMOVE_NAMESPACE:
      case REMOVE_BUCKET:
        return <span className="change-text">remove {change.key}</span>
      default:
        return `Unknown change: ${JSON.stringify(change)}`
    }
  }
}

Change.propTypes = {
  className: PropTypes.string.isRequired,
  change: PropTypes.object.isRequired
}
