import React, { Component, PropTypes } from 'react'
import { formatDate } from './FormattedDate.jsx'

export default class Config extends Component {
  render() {
    const { config, handleClick } = this.props

    return (<div className='config' onClick={handleClick}>
      <span className="sha">v{config.version || 0}</span>
      <span className="user"> by {config.user || 'unknown'} at </span>
      <span className="date">{formatDate(config.date) || 'unknown'}</span>
    </div>)
  }
}

Config.propTypes = {
  config: PropTypes.object.isRequired,
  handleClick: PropTypes.func.isRequired
}
