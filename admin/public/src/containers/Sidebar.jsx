import React, { Component, PropTypes } from 'react'
import Changes from './Changes.jsx'
import Configs from './Configs.jsx'
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
      fetchConfigs, commit,
      lastUpdated
    } = this.props

    return (<Changes
      lastUpdated={lastUpdated}
      handleUndo={undo}
      handleRedo={redo}
      handleCommit={commit}
      handleRefresh={fetchConfigs}
      changes={changes}
    />)
  }

  renderConfigs() {
    const { configs, loadConfig } = this.props

    return (<Configs
      configs={configs}
      loadConfig={loadConfig}
    />)
  }

  renderVersion() {
    const { version, currentVersion } = this.props

    let currentVersionStr = `v${currentVersion}`
    let versionStr = ''

    if (version != currentVersion) {
      versionStr += ` (viewing v${version})`
    }

    return (<h4>{currentVersionStr}<small>{versionStr}</small></h4>)
  }

  render() {
    const { env, selectedNamespace } = this.props

    let classNames = ['flex-box-md', 'sidebar']

    if (selectedNamespace) {
      classNames.push('flexed')
    }

    return (<div className={classNames.join(' ')}>
      <div>
        <h1 className={env.environment}>QuotaService</h1>
        {this.renderVersion()}
      </div>
      {this.renderAddNamespace()}
      {selectedNamespace && this.renderAddBucket()}
      {this.renderError()}
      {this.renderChanges()}
      {this.renderConfigs()}
    </div>)
  }
}

Sidebar.propTypes = {
  changes: PropTypes.object.isRequired,
  error: PropTypes.object,
  currentVersion: PropTypes.number.isRequired,
  version: PropTypes.number.isRequired,
  selectedNamespace: PropTypes.object,
  configs: PropTypes.object.isRequired,
  undo: PropTypes.func.isRequired,
  redo: PropTypes.func.isRequired,
  commit: PropTypes.func.isRequired,
  fetchConfigs: PropTypes.func.isRequired,
  addNamespace: PropTypes.func.isRequired,
  addBucket: PropTypes.func.isRequired,
  loadConfig: PropTypes.func.isRequired,
  lastUpdated: PropTypes.number,
  env: PropTypes.object.isRequired
}
