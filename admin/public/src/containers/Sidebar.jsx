import React, { Component, PropTypes } from 'react'
import Changes from './Changes.jsx'
import AddField from '../components/AddField.jsx'
import Error from '../components/Error.jsx'

export default class Sidebar extends Component {
  constructor() {
    super()

    this.state = {
      namespace: '',
      bucket: ''
    }
  }

  handleChange(key) {
    return (e) => this.setState({ [key]: e.target.value })
  }

  handleAddNamespace = () => {
    const { namespace } = this.state

    if (namespace !== '') {
      this.props.addNamespace(namespace)
      this.setState({ namespace: '' })
    }
  }

  handleAddBucket = () => {
    const { bucket } = this.state
    const { selectedNamespace, addBucket } = this.props

    if (bucket !== '') {
      addBucket(selectedNamespace.name, bucket)
      this.setState({ bucket: '' })
    }
  }

  renderError() {
    const { error } = this.props

    if (!error)
      return

    return <Error error={error} />
  }

  renderAddBucket() {
    return (<AddField
      value={this.state.bucket}
      handleChange={this.handleChange('bucket')}
      placeholder='bucket name'
      submitText='Add Bucket'
      handleSubmit={this.handleAddBucket}
    />)
  }

  renderAddNamespace() {
    return (<AddField
      value={this.state.namespace}
      handleChange={this.handleChange('namespace')}
      placeholder='namespace name'
      submitText='Add Namespace'
      handleSubmit={this.handleAddNamespace}
    />)
  }

  renderChanges() {
    const {
      changes, undo, redo,
      fetchNamespaces, commit,
      lastUpdated
    } = this.props

    return (<Changes
      lastUpdated={lastUpdated}
      handleUndo={undo}
      handleRedo={redo}
      handleCommit={commit}
      handleRefresh={fetchNamespaces}
      changes={changes}
    />)
  }

  render() {
    const { env, selectedNamespace } = this.props

    return (<div className='flex-box-md sidebar'>
      <div>
        <h1 className={env.environment}>QuotaService</h1>
        <small>{env.version}</small>
      </div>
      {this.renderAddNamespace()}
      {selectedNamespace && this.renderAddBucket()}
      {this.renderError()}
      {this.renderChanges()}
    </div>)
  }
}

Sidebar.propTypes = {
  changes: PropTypes.object.isRequired,
  error: PropTypes.object,
  selectedNamespace: PropTypes.object,
  undo: PropTypes.func.isRequired,
  redo: PropTypes.func.isRequired,
  commit: PropTypes.func.isRequired,
  fetchNamespaces: PropTypes.func.isRequired,
  addNamespace: PropTypes.func.isRequired,
  addBucket: PropTypes.func.isRequired,
  lastUpdated: PropTypes.number,
  env: PropTypes.object.isRequired
}
