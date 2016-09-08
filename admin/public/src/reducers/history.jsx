import Immutable from 'seamless-immutable'

import { UNDO, REDO, CANCEL_COMMIT, COMMIT, CLEAR } from '../actions/history.jsx'
import * as MutableActions from '../actions/mutable.jsx'

export const INITIAL_HISTORY = {
  items: Immutable.from({}),
  history: {
    past: [],
    future: []
  }
}

const MERGEABLE_CHANGES = [MutableActions.UPDATE_NAMESPACE, MutableActions.UPDATE_BUCKET]

function undo(state) {
  const { history, ...present } = state
  const { past, future } = history

  if (past.length < 1)
    return state

  const { change, ...newPresent } = past[0]
  const newFuture = Object.assign({}, present, { change: change })

  return Object.assign({}, newPresent, {
    history: {
      past: past.slice(1, past.length),
      future: [...future, newFuture]
    }
  })
}

function redo(state) {
  const { history, ...present } = state
  const { past, future } = history

  if (future.length < 1)
    return state

  const { change, ...newPresent } = future[future.length - 1]
  const redoPast = Object.assign({}, present, { change: change })

  return Object.assign({}, newPresent, {
    history: {
      past: [redoPast, ...past],
      future: future.slice(0, future.length - 1)
    }
  })
}

function commit(state) {
  return Object.assign({}, state, { commit: true })
}

function cancelCommit(state) {
  const newState = Object.assign({}, state)
  delete newState.commit
  return newState
}

function changesMergeable(previous, current) {
  return MERGEABLE_CHANGES.includes(current.type) &&
    previous.type == current.type && previous.key == current.key
}

function reduce(reducer, state, action) {
  const { history, ...currentState } = state

  let res = reducer(currentState, action)

  if (!Object.keys(MutableActions).includes(action.type)) {
    return Object.assign({}, { history: history }, res)
  }

  const { past: pastHistory } = history
  let { change: nextChange, ...nextState } = res

  const [lastPastEntry, ...restPastHistory] = pastHistory
  let nextPastHistory = pastHistory
  let nextPastState = currentState

  // If the last change in the past history is to the same
  // object as the current one, we merge them so that the
  // history isn't a bunch of single-character changes
  if (lastPastEntry) {
    const { change: lastPastChange, ...lastPastState } = lastPastEntry

    if (changesMergeable(lastPastChange, nextChange)) {
      nextPastState = lastPastState
      nextPastHistory = restPastHistory
    }
  }

  const historyEntry = Object.assign({ change: nextChange }, nextPastState)

  return Object.assign({}, nextState, {
    history: {
      past: [historyEntry, ...nextPastHistory],
      future: []
    }
  })
}

export function history(reducer) {
  return (state = INITIAL_HISTORY, action) => {
    switch (action.type) {
      case UNDO:
        return undo(state)
      case REDO:
        return redo(state)
      case COMMIT:
        return commit(state)
      case CANCEL_COMMIT:
        return cancelCommit(state)
      case CLEAR:
        return INITIAL_HISTORY
      default:
        return reduce(reducer, state, action)
    }
  }
}
