import React, { Component, PropTypes } from 'react'

export default class Field extends Component {
  constructor(props) {
    super(props)

    const {
      handleChange, keyName,
      parent, value
    } = props

    this.state = {
      id: `${parent}_${keyName}`,
      handleChange: handleChange(keyName),
      value: value || ''
    }
  }

  componentWillReceiveProps(nextProps) {
    this.setState({ value: nextProps.value || '' })
  }

  handleChange = (e) => {
    const { handleChange } = this.state
    const value = e.target.value
    const intValue = parseInt(value)

    this.setState({ value: value })

    if (value === '') {
      handleChange(null)
    } else if (!Number.isNaN(intValue)) {
      handleChange(intValue)
    }
  }

  render() {
    const { placeholder, keyName }= this.props
    const { id, value } = this.state

    return (<div className='flex-container input-box'>
      <label htmlFor={id} className='input-label'>{keyName}</label>
      <div className='input-field'>
        <input
          type='text'
          id={id}
          value={value}
          onChange={this.handleChange}
          placeholder={placeholder}
        />
      </div>
    </div>)
  }
}

Field.propTypes = {
  value: PropTypes.any,
  handleChange: PropTypes.func.isRequired,
  placeholder: PropTypes.string,
  keyName: PropTypes.string.isRequired,
  parent: PropTypes.string.isRequired
}
