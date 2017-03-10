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
      this.props.actions.addNamespace(namespace)
      this.setState({ namespace: '' })
    }
  }

  handleAddBucket = () => {
    const { bucket } = this.state
    const { selectedNamespace, actions } = this.props

    if (bucket !== '') {
      actions.addBucket(selectedNamespace.name, bucket)
      this.setState({ bucket: '' })
    }
  }

  renderError() {
    const { error } = this.props.configs

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

  handleCommit = () => {
    const { actions, namespaces, currentVersion } = this.props
    actions.commitConfig(namespaces.items, currentVersion)
  }

  renderChanges() {
    const { namespaces, actions } = this.props

    return (<Changes
      handleUndo={actions.undo}
      handleRedo={actions.redo}
      handleRefresh={actions.fetchConfigs}
      handleCommit={this.handleCommit}
      changes={namespaces.history}
    />)
  }

  renderConfigs() {
    const { configs, actions } = this.props

    return (<Configs
      configs={configs}
      loadConfig={actions.loadConfig}
    />)
  }

  renderVersion() {
    const { namespaces, currentVersion } = this.props
    const version = namespaces.version || 0

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
  actions: PropTypes.object.isRequired,
  namespaces: PropTypes.object.isRequired,
  configs: PropTypes.object.isRequired,
  selectedNamespace: PropTypes.object,
  currentVersion: PropTypes.number.isRequired,
  env: PropTypes.object.isRequired
}
