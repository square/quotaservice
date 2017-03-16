import React, { Component, PropTypes } from 'react'

export default class AddField extends Component {
  handleSubmit = (e) => {
    e.preventDefault()
    this.props.handleSubmit()
  }

  render() {
    const {
      placeholder, value,
      handleChange, submitText
    } = this.props

    return (<form
      className="flex-container input-box flex-wrap flex-end"
      onSubmit={this.handleSubmit}
    >
      <input
        type="text"
        placeholder={placeholder}
        className="flex-box"
        value={value}
        onChange={handleChange}
      />
      <button type="submit" className="btn btn-primary btn-attached">{submitText}</button>
    </form>)
  }
}

AddField.propTypes = {
  placeholder: PropTypes.string.isRequired,
  value: PropTypes.string.isRequired,
  submitText: PropTypes.string.isRequired,
  handleChange: PropTypes.func.isRequired,
  handleSubmit: PropTypes.func.isRequired
}
