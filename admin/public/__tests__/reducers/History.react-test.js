import Immutable from 'seamless-immutable'

import { history, INITIAL_HISTORY } from '../../src/reducers/history.jsx'
import { UNDO, REDO } from '../../src/actions/history.jsx'

import { namespaces } from '../../src/reducers/namespaces.jsx'
import * as MutableActions from '../../src/actions/mutable.jsx'

const reducer = history(namespaces)

describe('history reducer', () => {
  it('should return the initial state', () => {
    expect(reducer(undefined, {})).toEqual(INITIAL_HISTORY)
  })

  it('should handle UNDO', () => {
    expect(reducer(INITIAL_HISTORY, {
      type: UNDO
    })).toEqual(INITIAL_HISTORY)

    const undoState = Immutable.from({ foo: 1 })
    const pastState = Immutable.from({ bar: 2 })

    expect(reducer({
      items: undoState,
      history: {
        past: [{ items: pastState, change: 2 }],
        future: []
      }
    }, {
      type: UNDO
    })).toEqual({
      items: pastState,
      history: {
        past: [],
        future: [{ items: undoState, change: 2 }]
      }
    })
  })

  it('should handle REDO', () => {
    expect(reducer(INITIAL_HISTORY, {
      type: REDO
    })).toEqual(INITIAL_HISTORY)

    const undoState = Immutable.from({ foo: 1 })
    const pastState = Immutable.from({ bar: 2 })

    expect(reducer({
      items: pastState,
      history: {
        past: [],
        future: [{ items: undoState, change: 2 }]
      }
    }, {
      type: REDO
    })).toEqual({
      items: undoState,
      history: {
        past: [{ items: pastState, change: 2 }],
        future: []
      }
    })
  })

  it('should handle mutable actions', () => {
    let namespace = (n, a = {}) => {
      return Immutable.from({
        [n]: Object.assign({
          name: n,
          buckets: {}
        }, a)
      })
    }

    let state = INITIAL_HISTORY
    let nextState = {
      items: namespace("new.namespace"),
      history: {
        past: [{
          items: Immutable.from({}),
          change: {
            type: MutableActions.ADD_NAMESPACE,
            key: "new.namespace"
          }
        }],
        future: []
      }
    }

    expect(reducer(state, {
      type: MutableActions.ADD_NAMESPACE,
      namespace: "new.namespace"
    })).toEqual(nextState)

    // Test merging

    state = nextState
    nextState = {
      items: namespace("new.namespace", { foo: "bar" }),
      history: {
        past: [{
          change: {
            type: MutableActions.UPDATE_NAMESPACE,
            key: "new.namespace.foo",
            value: "bar"
          },
          items: namespace("new.namespace")
        }, {
          change: {
            type: MutableActions.ADD_NAMESPACE,
            key: "new.namespace"
          },
          items: Immutable.from({})
        }],
        future: []
      }
    }

    expect(reducer(state, {
      type: MutableActions.UPDATE_NAMESPACE,
      namespace: "new.namespace",
      key: "foo",
      value: "bar"
    })).toEqual(nextState)


    // We merge consective updates

    state = nextState
    nextState = {
      items: namespace("new.namespace", { foo: "bar2" }),
      history: {
        past: [{
          change: {
            type: MutableActions.UPDATE_NAMESPACE,
            key: "new.namespace.foo",
            value: "bar2"
          },
          items: namespace("new.namespace")
        }, {
          change: {
            type: MutableActions.ADD_NAMESPACE,
            key: "new.namespace"
          },
          items: Immutable.from({})
        }],
        future: []
      }
    }

    expect(reducer(state, {
      type: MutableActions.UPDATE_NAMESPACE,
      namespace: "new.namespace",
      key: "foo",
      value: "bar2"
    })).toEqual(nextState)

    state = nextState
    nextState = {
      items: Immutable.from(Object.assign(
        {},
        namespace("new.namespace2"),
        namespace("new.namespace", { foo: "bar2" }),
      )),
      history: {
        past: [{
          change: {
            type: MutableActions.ADD_NAMESPACE,
            key: "new.namespace2"
          },
          items: namespace("new.namespace", { foo: "bar2" })
        },{
          change: {
            type: MutableActions.UPDATE_NAMESPACE,
            key: "new.namespace.foo",
            value: "bar2"
          },
          items: namespace("new.namespace")
        }, {
          change: {
            type: MutableActions.ADD_NAMESPACE,
            key: "new.namespace"
          },
          items: Immutable.from({})
        }],
        future: []
      }
    }

    expect(reducer(state, {
      type: MutableActions.ADD_NAMESPACE,
      namespace: "new.namespace2"
    })).toEqual(nextState)


    // We don't merge across non-mergeable boundaries
    state = nextState
    nextState = {
      items: Immutable.from(Object.assign(
        {},
        namespace("new.namespace2"),
        namespace("new.namespace", { foo: "bar3" }),
      )),
      history: {
        past: [{
          change: {
            type: MutableActions.UPDATE_NAMESPACE,
            key: "new.namespace.foo",
            value: "bar3"
          },
          items: Immutable.from(Object.assign(
            {},
            namespace("new.namespace2"),
            namespace("new.namespace", { foo: "bar2" }),
          ))
        }, {
          change: {
            type: MutableActions.ADD_NAMESPACE,
            key: "new.namespace2"
          },
          items: namespace("new.namespace", { foo: "bar2" })
        },{
          change: {
            type: MutableActions.UPDATE_NAMESPACE,
            key: "new.namespace.foo",
            value: "bar2"
          },
          items: namespace("new.namespace")
        }, {
          change: {
            type: MutableActions.ADD_NAMESPACE,
            key: "new.namespace"
          },
          items: Immutable.from({})
        }],
        future: []
      }
    }

    expect(reducer(state, {
      type: MutableActions.UPDATE_NAMESPACE,
      namespace: "new.namespace",
      key: "foo",
      value: "bar3"
    })).toEqual(nextState)
  })

  it('should handle immutable actions', () => {
    expect(reducer(INITIAL_HISTORY, {
      type: "MY_ACTION",
      payload: "test"
    })).toEqual(INITIAL_HISTORY)
  })
})
