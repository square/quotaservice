import Immutable from 'seamless-immutable';

import { CONFIGS_FETCH_SUCCESS, LOAD_CONFIG } from '../../src/actions/configs.jsx';
import { SELECT_NAMESPACE } from '../../src/actions/namespaces.jsx';
import * as MutableActions from '../../src/actions/mutable.jsx';
import { namespaces, selectedNamespace } from '../../src/reducers/namespaces.jsx';

describe('namesapces reducer', () => {
  it('should handle initial state', () => {
    // namespaces has no default, it's passed through from INITIAL_HISTORY
    expect(namespaces({}, {})).toEqual({})
  })

  it('should handle CONFIGS_FETCH_SUCCESS', () => {
    expect(namespaces({}, {
      type: CONFIGS_FETCH_SUCCESS,
      payload: { configs: [] }
    })).toEqual({})

    expect(namespaces({}, {
      type: CONFIGS_FETCH_SUCCESS,
      payload: {
        configs: [
          {
            version: 1,
            namespaces: { test: 1 }
          }
        ]
      }
    })).toEqual({
      items: Immutable.from({ test: 1 }),
      version: 1,
      history: {
        past: [],
        future: []
      }
    })
  })

  it('should handle LOAD_CONFIG', () => {
    expect(namespaces({}, {
      type: LOAD_CONFIG,
      config: {
        version: 1,
        namespaces: { test: 1 }
      }
    })).toEqual({
      items: Immutable.from({ test: 1 }),
      version: 1,
      history: {
        past: [],
        future: []
      }
    })
  })

  const INITIAL_STATE = {
    items: Immutable.from({
      ['test.namespace']: {
        name: 'test.namespace',
        buckets: {
          foo: { size: 100 }
        }
      }
    }),
  }

  it('should handle ADD_NAMESPACE', () => {
    expect(namespaces(INITIAL_STATE, {
      type: MutableActions.ADD_NAMESPACE,
      namespace: 'test.namespace2'
    })).toEqual({
      change: {
        type: MutableActions.ADD_NAMESPACE,
        key: 'test.namespace2'
      },
      items: Immutable.from({
        ['test.namespace2']: {
          name: 'test.namespace2',
          buckets: {}
        },
        ['test.namespace']: {
          name: 'test.namespace',
          buckets: {
            foo: { size: 100 }
          }
        }
      })
    })
  })

  it('should handle UPDATE_NAMESPACE', () => {
    expect(namespaces(INITIAL_STATE, {
      type: MutableActions.UPDATE_NAMESPACE,
      namespace: 'test.namespace',
      key: 'foo',
      value: 'bar'
    })).toEqual({
      change: {
        type: MutableActions.UPDATE_NAMESPACE,
        key: 'test.namespace.foo',
        value: 'bar'
      },
      items: Immutable.from({
        ['test.namespace']: {
          name: 'test.namespace',
          foo: 'bar',
          buckets: {
            foo: { size: 100 }
          }
        }
      })
    })
  })

  it('should handle REMOVE_NAMESPACE', () => {
    expect(namespaces(INITIAL_STATE, {
      type: MutableActions.REMOVE_NAMESPACE,
      namespace: 'test.namespace',
    })).toEqual({
      change: {
        type: MutableActions.REMOVE_NAMESPACE,
        key: 'test.namespace',
      },
      items: Immutable.from({})
    })
  })

  it('should handle ADD_BUCKET', () => {
    expect(namespaces(INITIAL_STATE, {
      type: MutableActions.ADD_BUCKET,
      namespace: 'test.namespace',
      bucket: 'bar'
    })).toEqual({
      change: {
        type: MutableActions.ADD_BUCKET,
        key: 'test.namespace.bar'
      },
      items: Immutable.from({
        ['test.namespace']: {
          name: 'test.namespace',
          buckets: {
            foo: { size: 100 },
            bar: {
              name: 'bar',
              namespace: 'test.namespace'
            }
          }
        }
      })
    })

    expect(namespaces(INITIAL_STATE, {
      type: MutableActions.ADD_BUCKET,
      namespace: 'test.namespace',
      bucket: '___DYNAMIC_BUCKET_TPL___'
    })).toEqual({
      change: {
        type: MutableActions.ADD_BUCKET,
        key: 'test.namespace.dynamic_bucket_template'
      },
      items: Immutable.from({
        ['test.namespace']: {
          name: 'test.namespace',
          dynamic_bucket_template: {
            name: '___DYNAMIC_BUCKET_TPL___',
            namespace: 'test.namespace'
          },
          buckets: {
            foo: { size: 100 }
          }
        }
      })
    })
  })

  it('should handle UPDATE_BUCKET', () => {
    expect(namespaces(INITIAL_STATE, {
      type: MutableActions.UPDATE_BUCKET,
      namespace: 'test.namespace',
      bucket: 'foo',
      key: 'size',
      value: 1000
    })).toEqual({
      change: {
        type: MutableActions.UPDATE_BUCKET,
        key: 'test.namespace.foo.size',
        value: 1000
      },
      items: Immutable.from({
        ['test.namespace']: {
          name: 'test.namespace',
          buckets: {
            foo: { size: 1000 }
          }
        }
      })
    })

    const state = {
      items: INITIAL_STATE.items.setIn(
        ['test.namespace', 'dynamic_bucket_template'],
        { name: '___DYNAMIC_BUCKET_TPL___', namespace: 'test.namespace' }
      )
    }

    expect(namespaces(state, {
      type: MutableActions.UPDATE_BUCKET,
      namespace: 'test.namespace',
      bucket: '___DYNAMIC_BUCKET_TPL___',
      key: 'size',
      value: 1000
    })).toEqual({
      change: {
        type: MutableActions.UPDATE_BUCKET,
        key: 'test.namespace.dynamic_bucket_template.size',
        value: 1000
      },
      items: Immutable.from({
        ['test.namespace']: {
          name: 'test.namespace',
          dynamic_bucket_template: {
            name: '___DYNAMIC_BUCKET_TPL___',
            namespace: 'test.namespace',
            size: 1000
          },
          buckets: {
            foo: { size: 100 }
          }
        }
      })
    })
  })

  it('should handle REMOVE_BUCKET', () => {
    expect(namespaces(INITIAL_STATE, {
      type: MutableActions.REMOVE_BUCKET,
      namespace: 'test.namespace',
      bucket: 'foo'
    })).toEqual({
      change: {
        type: MutableActions.REMOVE_BUCKET,
        key: 'test.namespace.foo',
      },
      items: Immutable.from({
        ['test.namespace']: {
          name: 'test.namespace',
          buckets: {}
        }
      })
    })

    const state = {
      items: INITIAL_STATE.items.setIn(
        ['test.namespace', 'dynamic_bucket_template'],
        { name: '___DYNAMIC_BUCKET_TPL___', namespace: 'test.namespace' }
      )
    }

    expect(namespaces(state, {
      type: MutableActions.REMOVE_BUCKET,
      namespace: 'test.namespace',
      bucket: '___DYNAMIC_BUCKET_TPL___'
    })).toEqual({
      change: {
        type: MutableActions.REMOVE_BUCKET,
        key: 'test.namespace.dynamic_bucket_template'
      },
      items: Immutable.from({
        ['test.namespace']: {
          name: 'test.namespace',
          buckets: { foo: { size: 100 } }
        }
      })
    })
  })
})

describe('selectedNamespace reducer', () => {
  it('should handle initial state', () => {
    expect(selectedNamespace(undefined, {})).toEqual(null)
  })

  it('should handle ADD_NAMESPACE', () => {
    expect(selectedNamespace(undefined, {
      type: MutableActions.ADD_NAMESPACE,
      namespace: 'foo'
    })).toEqual({
      namespace: 'foo',
      canMakeChanges: true,
    });
  })

  it('should handle REMOVE_NAMESPACE', () => {
    expect(selectedNamespace('foo', {
      type: MutableActions.REMOVE_NAMESPACE
    })).toEqual(null)
  })

  it('should handle SELECT_NAMESPACE', () => {
    expect(selectedNamespace(undefined, {
      type: SELECT_NAMESPACE,
      namespace: 'foo',
      canMakeChanges: false,
    })).toEqual({
      namespace: 'foo',
      canMakeChanges: false,
    });
  })
})
