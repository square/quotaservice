import React, { Component, PropTypes } from 'react'
import Error from '../components/Error.jsx'
import NamespaceHeader from '../components/NamespaceHeader.jsx'

export default class Stats extends Component {
  constructor() {
    super()
    this.state = { searchValue: '' }
  }

  handleSearchChange = (e) => {
    let { searchTimer } = this.state

    if (searchTimer) {
      clearTimeout(searchTimer)
    }

    const value = e.target.value
    searchTimer = setTimeout(
      this.searchStats(value), 300
    )

    this.setState({
      searchValue: value,
      searchTimer: searchTimer
    })
  }

  searchStats = (bucket) => {
    const { namespace, fetchStats } = this.props
    return () => { fetchStats(namespace.name, bucket) }
  }

  handleBack = () => {
    this.props.toggleStats()
  }

  renderBucketStats(items, bucketName) {
    if (!items) {
      return
    }

    const stats = items[bucketName]

    if (!stats) {
      return
    }

    return (<div>
      <div className='flex-container input-box'>
        <label className='input-label'>hits</label>
        <div className="input-field">{stats.hits}</div>
      </div>
      <div className='flex-container input-box'>
        <label className='input-label'>misses</label>
        <div className="input-field">{stats.misses}</div>
      </div>
    </div>)
  }

  renderError(error) {
    if (!error) {
      return
    }

    return <Error error={error} />
  }

  renderBucketSearch() {
    const { searchValue } = this.state
    const { inRequest, error, items } = this.props.stats

    let classNames = ['flex-container', 'input-box', 'flex-end']

    if (inRequest) {
      classNames.push('loading')
    }

    return (<div className="bucket flex-tile flex-box">
      <div className={classNames.join(' ')}>
        <input
          type="text"
          placeholder="search dynamic bucket name"
          className="flex-box"
          value={searchValue}
          onChange={this.handleSearchChange}
        />
        <i className="icon" />
      </div>
      {this.renderBucketStats(items, searchValue)}
      {this.renderError(error)}
    </div>)
  }

  renderTopList(title, list) {
    if (list.length == 0)
      return

    return (<div className="bucket flex-tile flex-box">
      <div className="flex-container legend">
        <h4>{title}</h4>
      </div>
      {list.map(stat => {
        return (<div key={stat.bucket} className='flex-container input-box'>
          <label className='input-label'>{stat.bucket}</label>
          <div className="input-field">{stat.value}</div>
        </div>)
      })}
    </div>)
  }

  render() {
    const { namespace, stats, removeNamespace } = this.props
    let { topHits, topMisses } = stats.items || {}

    return (<div className="namespace flex-box flex-tile">
      <NamespaceHeader namespace={namespace}
        handleBack={this.handleBack}
        removeNamespace={removeNamespace}
      />
      <div className="buckets">
        {this.renderBucketSearch()}
        {topHits && this.renderTopList('top dynamic bucket hits', topHits) }
        {topMisses && this.renderTopList('top dynamic bucket misses', topMisses) }
      </div>
    </div>)
  }
}

Stats.propTypes = {
  namespace: PropTypes.object.isRequired,
  stats: PropTypes.object.isRequired,
  toggleStats: PropTypes.func.isRequired,
  removeNamespace: PropTypes.func.isRequired,
  fetchStats: PropTypes.func.isRequired
}
