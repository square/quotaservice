import React, { Component, PropTypes } from 'react'
import Change from '../components/Change.jsx'

export default class Changes extends Component {
  renderChanges() {
    let { past, future } = this.props.changes

    if (past.length == 0 && future.length == 0) {
      return (<div className='changes'>
        <div className='change future'>no changes recorded</div>
      </div>)
    } else {
      return (<div className='changes'>
        {future.map((ch, i) => {
          return <Change key={`future-${i}`} className='future' change={ch.change} />
        })}
        {past.map((ch, i) => {
          return <Change key={`past-${i}`} className='past' change={ch.change} />
        })}
      </div>)
    }
  }

  render() {
    const {
      handleUndo, handleRedo,
      handleRefresh, handleCommit,
    } = this.props

    let { past, future } = this.props.changes
    const canUndo = past.length > 0
    const canRedo = future.length > 0

    return (<div>
      <div className='actions'>
        <div className='flex-box'>
          <button className='btn' title='Undo last change' onClick={handleUndo} disabled={!canUndo}>Undo</button>
          <button className='btn' title='Redo last change' onClick={handleRedo} disabled={!canRedo}>Redo</button>
        </div>
        <div className='flex-box'>
          <button className='btn btn-danger' title='Refresh configuration' onClick={handleRefresh}>
            Refresh
          </button>
          <button className='btn btn-primary' onClick={handleCommit}>Save</button>
        </div>
      </div>
      {this.renderChanges()}
    </div>)
  }
}

Changes.propTypes = {
  changes: PropTypes.object.isRequired,
  handleUndo: PropTypes.func.isRequired,
  handleRedo: PropTypes.func.isRequired,
  handleCommit: PropTypes.func.isRequired,
  handleRefresh: PropTypes.func.isRequired
}
