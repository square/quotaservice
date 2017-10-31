import PropTypes from 'prop-types';
import React from 'react';
import { Component } from 'react';

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
    } else if (this.validInteger(intValue)) {
      handleChange(intValue)
    }
  }

  validInteger(int) {
    return !Number.isNaN(int) &&
      int > Number.MIN_SAFE_INTEGER &&
      int < Number.MAX_SAFE_INTEGER
  }

  render() {
    const { disabled, placeholder, title, keyName } = this.props
    const { id, value } = this.state

    let keyTitle = keyName
    if (title) {
      keyTitle = <abbr title={title}>{keyName}</abbr>
    }

    return (
      <div className='flex-container input-box'>
        <label htmlFor={id} className='input-label'>
          {keyTitle}
        </label>
        <div className='input-field'>
          <input
            type='text'
            id={id}
            value={value}
            disabled={disabled === true}
            onChange={this.handleChange}
            placeholder={placeholder}
          />
        </div>
      </div>
    )
  }
}

Field.propTypes = {
  disabled: PropTypes.bool,
  value: PropTypes.any,
  handleChange: PropTypes.func.isRequired,
  placeholder: PropTypes.string,
  keyName: PropTypes.string.isRequired,
  parent: PropTypes.string.isRequired,
  title: PropTypes.string
}
