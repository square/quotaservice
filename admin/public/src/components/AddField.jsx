import React, { Component, PropTypes } from 'react'

export default class AddField extends Component {
  render() {
    const {
      placeholder, value, handleChange,
      submitText, handleSubmit
    } = this.props

    return (<div className="flex-container input-box flex-wrap flex-end">
      <input
        type="text"
        placeholder={placeholder}
        className="flex-box"
        value={value}
        onChange={handleChange}
      />
      <button className="btn btn-primary btn-attached" onClick={handleSubmit}>{submitText}</button>
    </div>)
  }
}

AddField.propTypes = {
  placeholder: PropTypes.string.isRequired,
  value: PropTypes.string.isRequired,
  submitText: PropTypes.string.isRequired,
  handleChange: PropTypes.func.isRequired,
  handleSubmit: PropTypes.func.isRequired
}
