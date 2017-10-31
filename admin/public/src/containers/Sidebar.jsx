import PropTypes from 'prop-types';
import React from 'react';
import { Component } from 'react';

import AddField from '../components/AddField.jsx';
import Error from '../components/Error.jsx';
import { BUCKET_KEY_MAP } from '../reducers/namespaces.jsx';
import Changes from './Changes.jsx';
import Configs from './Configs.jsx';

export default class Sidebar extends Component {
  constructor() {
    super()

    this.state = {
      namespace: '',
      bucket: ''
    }
  }

  canMakeChanges() {
    const { selectedNamespace } = this.props;
    return selectedNamespace ? selectedNamespace.canMakeChanges : false;
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

  handleAddBucket(bucket) {
    return () => {
      if (bucket === undefined) {
        bucket = this.state.bucket
        this.setState({ bucket: '' })
      }

      const { selectedNamespace, actions } = this.props

      if (bucket !== '') {
        actions.addBucket(selectedNamespace.namespace.name, bucket)
      }
    }
  }

  renderError() {
    const { error } = this.props.configs

    if (!error)
      return

    return <Error error={error} />
  }

  renderAddBucket() {
    if (!this.props.selectedNamespace)
      return null;

    if (!this.canMakeChanges())
      return null;

    return <AddField
      value={this.state.bucket}
      handleChange={this.handleChange('bucket')}
      placeholder='Bucket name'
      submitText='Add Bucket'
      handleSubmit={this.handleAddBucket()}
    />
  }

  renderSpecialBucketButtons() {
    const { selectedNamespace } = this.props

    if (!selectedNamespace || !this.canMakeChanges())
      return null

    const buttons = []
    const { namespace } = selectedNamespace;

    for (let [name, key] of Object.entries(BUCKET_KEY_MAP)) {
      if (!namespace[key]) {
        buttons.push(<button
          key={key} className="btn btn-primary"
          onClick={this.handleAddBucket(name)}
        >Add {name}</button>)
      }
    }

    if (buttons.length == 0)
      return null;

    return (
      <div className="flex-container flex-wrap flex-end">
        {buttons}
      </div>
    )
  }

  renderAddNamespace() {
    return <AddField
      value={this.state.namespace}
      handleChange={this.handleChange('namespace')}
      placeholder='Namespace name'
      submitText='Add Namespace'
      handleSubmit={this.handleAddNamespace}
    />
  }

  handleCommit = () => {
    const { actions, namespaces, currentVersion } = this.props
    actions.commitConfig(namespaces.items, currentVersion)
  }

  renderChanges() {
    const { namespaces, actions } = this.props

    return <Changes
      handleUndo={actions.undo}
      handleRedo={actions.redo}
      handleRefresh={this.props.handleRefresh}
      handleCommit={this.handleCommit}
      changes={namespaces.history}
    />
  }

  renderConfigs() {
    const { configs, actions } = this.props

    return <Configs
      configs={configs}
      loadConfig={actions.loadConfig}
    />
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
      {this.renderAddBucket()}
      {this.renderSpecialBucketButtons()}
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
  handleRefresh: PropTypes.func.isRequired,
  env: PropTypes.object.isRequired
}
