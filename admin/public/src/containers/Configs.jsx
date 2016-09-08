import React, { Component, PropTypes } from 'react'
import Config from '../components/Config.jsx'

export default class Configs extends Component {
  changeConfig(config) {
    const { loadConfig } = this.props
    return () => loadConfig(config)
  }

  render() {
    const { configs } = this.props
    const { items } = configs

    if (items === undefined) {
      return null
    }

    return (<div className='configs'>
      {items.map(c => {
        let key = c.version + c.date
        return (<Config config={c} key={key} handleClick={this.changeConfig(c)} />)
      })}
    </div>)
  }
}

Configs.propTypes = {
  configs: PropTypes.object.isRequired,
  loadConfig: PropTypes.func.isRequired
}
